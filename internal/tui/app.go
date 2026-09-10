package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"skillyfiy/internal/config"
	"skillyfiy/internal/engine"
	"skillyfiy/internal/model"
)

type appState int

const (
	stateMain appState = iota
	stateConfirm
)

type paneFocus int

const (
	paneList paneFocus = iota
	paneInspector
)

// AppModel is the top-level Bubble Tea model orchestrating the Skillyfiy TUI.
type AppModel struct {
	cfg          *config.Config
	keys         config.KeyMap
	state        appState
	focusedPane  paneFocus
	width        int
	height       int
	sortMode     model.SortMode
	visualAnchor int

	allItems    []model.AgentItem
	selectedMap map[string]bool

	searchInput textinput.Model
	list        list.Model
	inspector   InspectorModel

	PurgeReport *engine.PurgeReport
	delegate    AgentItemDelegate
}

// NewApp provides an ergonomic alias for NewAppModel.
func NewApp(cfg *config.Config, items []model.AgentItem) AppModel {
	return NewAppModel(cfg, items)
}

// NewAppModel initializes the Bubble Tea application state.
func NewAppModel(cfg *config.Config, items []model.AgentItem) AppModel {
	ti := textinput.New()
	ti.Placeholder = "Search skills, MCPs, paths..."
	ti.Prompt = SearchPromptStyle.Render("❯ ")
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorMintBlue).Bold(true)
	ti.TextStyle = SearchTextStyle
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorDullSilver)
	ti.CharLimit = 120

	visualAnchor := -1
	delegate := NewAgentItemDelegate(&visualAnchor)

	l := list.New([]list.Item{}, delegate, 40, 20)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.DisableQuitKeybindings()

	inspector := NewInspector(50, 20)

	selectedMap := make(map[string]bool)
	for _, it := range items {
		if it.Selected {
			selectedMap[it.ID] = true
		}
	}

	app := AppModel{
		cfg:          cfg,
		keys:         config.DefaultKeyMap,
		state:        stateMain,
		focusedPane:  paneList,
		width:        80,
		height:       24,
		sortMode:     model.SortAlphabetical,
		visualAnchor: visualAnchor,
		allItems:     items,
		selectedMap:  selectedMap,
		searchInput:  ti,
		list:         l,
		inspector:    inspector,
		delegate:     delegate,
	}

	app.updateListItems()
	return app
}

// Init triggers initial commands.
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Update processes terminal messages and key inputs.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateSizes()
		return m, nil

	case tea.KeyMsg:
		// Always handle ForceQuit (Ctrl+C)
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}

		// State: Confirmation Modal
		if m.state == stateConfirm {
			switch msg.String() {
			case "y", "Y":
				report, err := engine.ExecutePurge(m.allItems, m.cfg.DryRun)
				if err != nil {
					report = &engine.PurgeReport{
						Errors: []string{err.Error()},
					}
				}
				m.PurgeReport = report
				return m, tea.Quit
			case "n", "N", "esc":
				m.state = stateMain
				return m, nil
			default:
				return m, nil
			}
		}

		// State: Main Dashboard
		// If Search Input is focused:
		if m.searchInput.Focused() {
			switch msg.Type {
			case tea.KeyEsc, tea.KeyEnter:
				m.searchInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				prevVal := m.searchInput.Value()
				m.searchInput, cmd = m.searchInput.Update(msg)
				if m.searchInput.Value() != prevVal {
					m.updateListItems()
				}
				return m, cmd
			}
		}

		// Search Input is not focused:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.visualAnchor != -1 {
				m.visualAnchor = -1
				m.syncDelegateAnchor()
				return m, nil
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.FocusSearch):
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.CycleSort):
			m.sortMode = (m.sortMode + 1) % 3
			m.updateListItems()
			return m, nil

		case key.Matches(msg, m.keys.SwitchPane):
			if m.focusedPane == paneList {
				m.focusedPane = paneInspector
				m.inspector.Focused = true
			} else {
				m.focusedPane = paneList
				m.inspector.Focused = false
			}
			return m, nil

		case key.Matches(msg, m.keys.ConfirmPurge):
			if m.getSelectedCount() > 0 {
				m.state = stateConfirm
			}
			return m, nil
		}

		// Pane-specific navigation
		if m.focusedPane == paneInspector {
			var cmd tea.Cmd
			m.inspector, cmd = m.inspector.Update(msg)
			return m, cmd
		}

		// List Pane focused:
		switch {
		case key.Matches(msg, m.keys.ToggleSelect):
			m.toggleCurrentSelection()
			return m, nil

		case key.Matches(msg, m.keys.SelectAll):
			m.toggleSelectAll()
			return m, nil

		case key.Matches(msg, m.keys.VisualRange):
			m.toggleVisualMode()
			return m, nil

		case msg.String() == "s":
			if m.visualAnchor != -1 {
				m.commitVisualRange()
				return m, nil
			}

		default:
			var cmd tea.Cmd
			prevIdx := m.list.Index()
			m.list, cmd = m.list.Update(msg)
			newIdx := m.list.Index()
			if newIdx != prevIdx {
				m.syncInspectorToCursor()
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI layout.
func (m AppModel) View() string {
	if m.state == stateConfirm {
		return RenderConfirmModal(m.allItems, m.width, m.height, m.cfg.DryRun)
	}

	var b strings.Builder

	// 1. Header Banner
	bannerStr := RenderBanner(m.width)
	b.WriteString(bannerStr + "\n\n")

	// 2. Search Prompt Box with active Sort Badge
	sortBadge := SortBadgeStyle.Render(m.sortMode.String())
	searchPrompt := m.searchInput.View()
	availSearchWidth := m.width - lipgloss.Width(sortBadge) - 8
	if availSearchWidth < 20 {
		availSearchWidth = 20
	}
	searchBoxContent := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(availSearchWidth).Render(searchPrompt),
		sortBadge,
	)
	b.WriteString(SearchBoxStyle.Width(m.width-4).Render(searchBoxContent) + "\n")

	// 3. Dual-Column Split View
	listStyle := LeftPaneStyle
	if m.focusedPane == paneList {
		listStyle = LeftPaneFocusedStyle
	}

	leftView := listStyle.Width(m.list.Width()).Height(m.list.Height()).Render(m.list.View())
	rightView := m.inspector.View()

	splitView := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	b.WriteString(splitView + "\n")

	// 4. Status Footer
	footer := m.renderFooter()
	b.WriteString(footer)

	return b.String()
}

func (m *AppModel) renderFooter() string {
	if m.visualAnchor != -1 {
		msg := fmt.Sprintf(" -- VISUAL RANGE (Pinned at #%d) -- Move cursor with ↑/↓/j/k and tap 'v' or 's' to lock selection. Tap Esc to cancel. ", m.visualAnchor+1)
		return FooterVisualModeStyle.Width(m.width).Render(msg)
	}

	selectedCount := m.getSelectedCount()
	reclaimedTokens := m.getReclaimedTokens()

	leftHints := "[↑/↓/j/k] Move • [Space/x] Toggle • [v] Visual • [Ctrl+A] Select All • [/] Search • [Tab] Pane • [Ctrl+S] Sort • [Enter] Purge • [Ctrl+Q] Quit"
	leftFormatted := FooterStyle.Render(leftHints)

	statusInfo := fmt.Sprintf("Selected: %s/%d  Reclaimed: %s",
		FooterSelectedCountStyle.Render(fmt.Sprintf("%d", selectedCount)),
		len(m.allItems),
		FooterReclaimedStyle.Render(fmt.Sprintf("~%d tk", reclaimedTokens)),
	)
	rightFormatted := lipgloss.NewStyle().Padding(0, 1).Render(statusInfo)

	availGap := m.width - lipgloss.Width(leftFormatted) - lipgloss.Width(rightFormatted)
	if availGap < 1 {
		availGap = 1
	}
	gap := strings.Repeat(" ", availGap)

	return leftFormatted + gap + rightFormatted
}

func (m *AppModel) recalculateSizes() {
	bannerH := BannerHeight() + 1
	searchH := 3
	footerH := 2
	availableH := m.height - bannerH - searchH - footerH
	if availableH < 8 {
		availableH = 8
	}

	// 45% left, 55% right
	leftW := int(float64(m.width) * 0.45)
	if leftW < 30 {
		leftW = 30
	}
	rightW := m.width - leftW - 4
	if rightW < 30 {
		rightW = 30
	}

	m.list.SetSize(leftW, availableH)
	m.inspector.SetSize(rightW, availableH)
}

func (m *AppModel) updateListItems() {
	// Sort all items
	sorted := make([]model.AgentItem, len(m.allItems))
	copy(sorted, m.allItems)

	sortItems(sorted, m.sortMode)

	// Sync selection states
	for i := range sorted {
		sorted[i].Selected = m.selectedMap[sorted[i].ID]
	}

	// Apply fuzzy search
	filtered := filterAgentItems(sorted, m.searchInput.Value())

	listItems := make([]list.Item, len(filtered))
	for i, it := range filtered {
		listItems[i] = it
	}

	m.list.SetItems(listItems)
	m.syncInspectorToCursor()
}

func (m *AppModel) syncInspectorToCursor() {
	items := m.list.Items()
	if len(items) == 0 {
		m.inspector.SetItem(nil)
		return
	}

	idx := m.list.Index()
	if idx < 0 {
		idx = 0
		m.list.Select(0)
	}
	if idx >= len(items) {
		idx = len(items) - 1
		m.list.Select(idx)
	}

	if item, ok := items[idx].(model.AgentItem); ok {
		m.inspector.SetItem(&item)
	}
}

func (m *AppModel) toggleCurrentSelection() {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}

	idx := m.list.Index()
	if idx < 0 || idx >= len(items) {
		return
	}

	item, ok := items[idx].(model.AgentItem)
	if !ok {
		return
	}

	newVal := !m.selectedMap[item.ID]
	m.selectedMap[item.ID] = newVal
	item.Selected = newVal
	m.list.SetItem(idx, item)

	// Update allItems mirror
	for i := range m.allItems {
		if m.allItems[i].ID == item.ID {
			m.allItems[i].Selected = newVal
			break
		}
	}

	// Automatically step cursor down 1 position (if not at bottom)
	if idx < len(items)-1 {
		m.list.CursorDown()
		m.syncInspectorToCursor()
	}
}

func (m *AppModel) toggleSelectAll() {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}

	allSelected := true
	for _, it := range items {
		agentItem := it.(model.AgentItem)
		if !m.selectedMap[agentItem.ID] {
			allSelected = false
			break
		}
	}

	targetState := !allSelected
	for i, it := range items {
		agentItem := it.(model.AgentItem)
		m.selectedMap[agentItem.ID] = targetState
		agentItem.Selected = targetState
		m.list.SetItem(i, agentItem)

		for j := range m.allItems {
			if m.allItems[j].ID == agentItem.ID {
				m.allItems[j].Selected = targetState
				break
			}
		}
	}
}

func (m *AppModel) toggleVisualMode() {
	if m.visualAnchor == -1 {
		idx := m.list.Index()
		m.visualAnchor = idx
		m.syncDelegateAnchor()

		// Select the anchor item
		items := m.list.Items()
		if idx >= 0 && idx < len(items) {
			item := items[idx].(model.AgentItem)
			m.selectedMap[item.ID] = true
			item.Selected = true
			m.list.SetItem(idx, item)
			for j := range m.allItems {
				if m.allItems[j].ID == item.ID {
					m.allItems[j].Selected = true
					break
				}
			}
		}
	} else {
		m.commitVisualRange()
	}
}

func (m *AppModel) commitVisualRange() {
	if m.visualAnchor == -1 {
		return
	}

	start := m.visualAnchor
	end := m.list.Index()
	if start > end {
		start, end = end, start
	}

	items := m.list.Items()
	for i := start; i <= end && i < len(items); i++ {
		item := items[i].(model.AgentItem)
		m.selectedMap[item.ID] = true
		item.Selected = true
		m.list.SetItem(i, item)

		for j := range m.allItems {
			if m.allItems[j].ID == item.ID {
				m.allItems[j].Selected = true
				break
			}
		}
	}

	m.visualAnchor = -1
	m.syncDelegateAnchor()
}

func (m *AppModel) syncDelegateAnchor() {
	m.delegate.VisualAnchor = &m.visualAnchor
	m.list.SetDelegate(m.delegate)
}

func (m *AppModel) getSelectedCount() int {
	count := 0
	for _, selected := range m.selectedMap {
		if selected {
			count++
		}
	}
	return count
}

func (m *AppModel) getReclaimedTokens() int {
	tokens := 0
	for _, item := range m.allItems {
		if m.selectedMap[item.ID] {
			tokens += item.Tokens
		}
	}
	return tokens
}

// Helpers for sorting and fuzzy filtering
func sortItems(items []model.AgentItem, mode model.SortMode) {
	switch mode {
	case model.SortAlphabetical:
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		})
	case model.SortTokenWeight:
		sort.Slice(items, func(i, j int) bool {
			if items[i].Tokens == items[j].Tokens {
				return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
			}
			return items[i].Tokens > items[j].Tokens
		})
	case model.SortItemType:
		sort.Slice(items, func(i, j int) bool {
			if items[i].Type == items[j].Type {
				return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
			}
			return items[i].Type == model.TypeSkill && items[j].Type == model.TypeMCP
		})
	}
}

type agentItemSource []model.AgentItem

func (s agentItemSource) Len() int            { return len(s) }
func (s agentItemSource) String(i int) string { return s[i].FilterValue() }

func filterAgentItems(items []model.AgentItem, query string) []model.AgentItem {
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return items
	}

	matches := fuzzy.FindFrom(cleanQuery, agentItemSource(items))
	result := make([]model.AgentItem, len(matches))
	for i, m := range matches {
		result[i] = items[m.Index]
	}
	return result
}
