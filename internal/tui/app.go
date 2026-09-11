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
	"github.com/muesli/reflow/truncate"

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
	cfg             *config.Config
	keys            config.KeyMap
	state           appState
	focusedPane     paneFocus
	width           int
	height          int
	sortMode        model.SortMode
	filterType      model.FilterType
	visualAnchor    int
	multiSelectMode bool
	showHelpOverlay bool
	searchQuery     string

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
	searchQuery := ""
	delegate := NewAgentItemDelegate(&visualAnchor, &searchQuery)

	l := list.New([]list.Item{}, delegate, 40, 20)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.DisableQuitKeybindings()

	inspector := NewInspector(50, 20)

	selectedMap := make(map[string]bool)
	for i := range items {
		items[i].EnsureCache()
		if items[i].Selected {
			selectedMap[items[i].ID] = true
		}
	}

	app := AppModel{
		cfg:             cfg,
		keys:            config.DefaultKeyMap,
		state:           stateMain,
		focusedPane:     paneList,
		width:           80,
		height:          24,
		sortMode:        model.SortNewest,
		filterType:      model.FilterAll,
		visualAnchor:    visualAnchor,
		multiSelectMode: false,
		showHelpOverlay: false,
		searchQuery:     searchQuery,
		allItems:        items,
		selectedMap:     selectedMap,
		searchInput:     ti,
		list:            l,
		inspector:       inspector,
		delegate:        delegate,
	}

	app.recalculateSizes()
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

		// State: Help Modal Overlay
		if m.showHelpOverlay {
			switch msg.String() {
			case "i", "I", "esc", "q", "enter", "?":
				m.showHelpOverlay = false
				return m, nil
			default:
				return m, nil
			}
		}

		// State: Main Dashboard
		// If Search Input is focused:
		if m.searchInput.Focused() {
			if msg.String() == "/" || msg.Type == tea.KeyEsc || msg.Type == tea.KeyEnter {
				m.searchInput.Blur()
				m.syncDelegateQuery()
				return m, nil
			}
			var cmd tea.Cmd
			prevVal := m.searchInput.Value()
			m.searchInput, cmd = m.searchInput.Update(msg)
			if m.searchInput.Value() != prevVal {
				m.updateListItems()
			}
			return m, cmd
		}

		// Search Input is not focused:
		switch {
		case msg.String() == "i" || msg.String() == "I" || key.Matches(msg, m.keys.Help):
			m.showHelpOverlay = true
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			if m.visualAnchor != -1 {
				m.visualAnchor = -1
				m.syncDelegateAnchor()
				return m, nil
			}
			if m.multiSelectMode {
				m.multiSelectMode = false
				return m, nil
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.FocusSearch) || msg.String() == "/":
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.CycleSort):
			m.sortMode = (m.sortMode + 1) % 6
			m.updateListItems()
			return m, nil

		case key.Matches(msg, m.keys.CycleFilter):
			m.filterType = (m.filterType + 1) % 3
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

		case key.Matches(msg, m.keys.VisualRange) || msg.String() == "v" || msg.String() == "V":
			m.toggleMultiSelectMode()
			return m, nil

		case key.Matches(msg, m.keys.VisualSpan) || msg.String() == "s" || msg.String() == "S":
			m.handleSpanKey()
			return m, nil

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

// View renders the TUI layout with strict zero-wrapping guards.
func (m AppModel) View() string {
	if m.width < 32 || m.height < 10 {
		return fmt.Sprintf("Terminal too small (%dx%d).\nPlease enlarge screen.\n", m.width, m.height)
	}

	if m.state == stateConfirm {
		return RenderConfirmModal(m.allItems, m.width, m.height, m.cfg.DryRun)
	}

	if m.showHelpOverlay {
		return RenderHelpModal(m.width, m.height)
	}

	var b strings.Builder

	// 1. Header Banner
	bannerStr := RenderBanner(m.width)
	b.WriteString(bannerStr + "\n")

	// 2. Search Prompt Box with Sort and Filter Badges
	isNarrow := m.width < 90
	var sortStr, filterStr string
	if isNarrow {
		sortStr = m.sortMode.Compact()
		filterStr = m.filterType.Compact()
	} else {
		sortStr = m.sortMode.String()
		filterStr = m.filterType.String()
	}

	sortBadge := SortBadgeStyle.Render(sortStr)
	filterBadge := FilterBadgeStyle.Render(filterStr)
	badges := lipgloss.JoinHorizontal(lipgloss.Center, sortBadge, " ", filterBadge)

	boxWidth := m.width - 6
	if boxWidth < 20 {
		boxWidth = 20
	}
	availSearchWidth := boxWidth - lipgloss.Width(badges) - 3
	if availSearchWidth < 8 {
		availSearchWidth = 8
	}

	searchBoxContent := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(availSearchWidth).Render(m.searchInput.View()),
		badges,
	)
	b.WriteString(SearchBoxStyle.Width(boxWidth).Render(searchBoxContent) + "\n")

	// 3. Middle Content (Split View or Stacked/Toggle View)
	listStyle := LeftPaneStyle
	if m.focusedPane == paneList {
		listStyle = LeftPaneFocusedStyle
	}

	listInnerH, _, isStacked := m.calculateContentHeights()

	var mainContent string
	if m.width >= 90 {
		leftView := listStyle.Width(m.list.Width()).Height(listInnerH).Render(m.list.View())
		rightView := m.inspector.View()
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	} else {
		if isStacked {
			leftView := listStyle.Width(m.list.Width()).Height(listInnerH).Render(m.list.View())
			rightView := m.inspector.View()
			mainContent = lipgloss.JoinVertical(lipgloss.Left, leftView, rightView)
		} else {
			if m.focusedPane == paneInspector {
				mainContent = m.inspector.View()
			} else {
				mainContent = listStyle.Width(m.list.Width()).Height(listInnerH).Render(m.list.View())
			}
		}
	}
	b.WriteString(mainContent + "\n")

	// 4. Status Footer
	footer := m.renderFooter()
	b.WriteString(footer)

	// Final pass: clamp lines and prevent any line-wrapping jitter
	rawView := b.String()
	lines := strings.Split(rawView, "\n")
	var cleanLines []string
	maxH := m.height
	if maxH <= 0 {
		maxH = 24
	}
	maxW := m.width
	if maxW <= 0 {
		maxW = 80
	}

	for i, line := range lines {
		if i >= maxH {
			break
		}
		if lipgloss.Width(line) >= maxW {
			line = truncate.StringWithTail(line, uint(maxW-1), "")
		}
		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}

func (m *AppModel) calculateContentHeights() (listInner, inspInner int, isStacked bool) {
	bannerH := BannerHeight(m.width)
	searchH := 3
	footerH := 1
	availMainOuter := m.height - bannerH - searchH - footerH
	if availMainOuter < 4 {
		availMainOuter = 4
	}

	if m.width >= 90 {
		// Dual-column split view
		innerH := availMainOuter - 2
		if innerH < 3 {
			innerH = 3
		}
		return innerH, innerH, false
	}

	// Narrow / mobile mode (<90 cols)
	if availMainOuter >= 14 {
		// Stacked mode: List on top, Inspector on bottom
		listOuter := int(float64(availMainOuter) * 0.55)
		if listOuter < 7 {
			listOuter = 7
		}
		inspOuter := availMainOuter - listOuter
		if inspOuter < 5 {
			inspOuter = 5
		}
		lInner := listOuter - 2
		if lInner < 2 {
			lInner = 2
		}
		iInner := inspOuter - 2
		if iInner < 2 {
			iInner = 2
		}
		return lInner, iInner, true
	}

	// Single pane toggle mode
	singleInner := availMainOuter - 2
	if singleInner < 2 {
		singleInner = 2
	}
	return singleInner, singleInner, false
}

func (m *AppModel) renderFooter() string {
	if m.visualAnchor != -1 {
		msg := fmt.Sprintf(" -- SPAN PINNED (#%d) -- Scroll to target item & tap 's' to lock (Esc cancels) ", m.visualAnchor+1)
		return truncate.StringWithTail(FooterVisualModeStyle.Render(msg), uint(m.width-1), "")
	}

	if m.multiSelectMode {
		msg := " -- MULTI-SELECT ACTIVE -- Move to item & tap 's' to pin span start (Esc exits) "
		return truncate.StringWithTail(FooterVisualModeStyle.Render(msg), uint(m.width-1), "")
	}

	selectedCount := m.getSelectedCount()
	reclaimedTokens := m.getReclaimedTokens()

	var leftHints string
	if m.width >= 75 {
		leftHints = "[i] Help Modal • [/] Search • [v] Multi-Select • [Enter] Purge"
	} else if m.width >= 52 {
		leftHints = "[i] Help • [/] Search • [v] Multi • [Enter] Purge"
	} else {
		leftHints = "[i] Help • [Enter] Purge"
	}
	leftFormatted := FooterStyle.Render(leftHints)

	var statusInfo string
	if m.width >= 65 {
		statusInfo = fmt.Sprintf("Selected: %s/%d  Reclaimed: %s",
			FooterSelectedCountStyle.Render(fmt.Sprintf("%d", selectedCount)),
			len(m.allItems),
			FooterReclaimedStyle.Render(fmt.Sprintf("~%d tk", reclaimedTokens)),
		)
	} else {
		statusInfo = fmt.Sprintf("%s/%d (~%dtk)",
			FooterSelectedCountStyle.Render(fmt.Sprintf("%d", selectedCount)),
			len(m.allItems),
			reclaimedTokens,
		)
	}
	rightFormatted := lipgloss.NewStyle().Padding(0, 1).Render(statusInfo)

	availGap := m.width - lipgloss.Width(leftFormatted) - lipgloss.Width(rightFormatted) - 1
	if availGap < 1 {
		availGap = 1
	}
	gap := strings.Repeat(" ", availGap)

	line := leftFormatted + gap + rightFormatted
	if lipgloss.Width(line) >= m.width {
		line = truncate.StringWithTail(line, uint(m.width-1), "")
	}
	return line
}

func (m *AppModel) recalculateSizes() {
	listInnerH, inspInnerH, _ := m.calculateContentHeights()

	if m.width >= 90 {
		// Dual-Column Split View
		availInner := m.width - 8
		if availInner < 40 {
			availInner = 40
		}
		leftW := int(float64(availInner) * 0.45)
		if leftW < 30 {
			leftW = 30
		}
		rightW := availInner - leftW
		if rightW < 30 {
			rightW = 30
		}

		m.list.SetSize(leftW, listInnerH)
		m.inspector.SetSize(rightW, inspInnerH)
	} else {
		// Responsive Mobile Mode (<90 cols)
		contentW := m.width - 4
		if contentW < 16 {
			contentW = 16
		}

		m.list.SetSize(contentW, listInnerH)
		m.inspector.SetSize(contentW-2, inspInnerH)
	}
}

func (m *AppModel) updateListItems() {
	// Sync selection states into allItems
	for i := range m.allItems {
		m.allItems[i].EnsureCache()
		m.allItems[i].Selected = m.selectedMap[m.allItems[i].ID]
	}

	m.searchQuery = m.searchInput.Value()
	m.delegate.SearchQuery = &m.searchQuery
	m.list.SetDelegate(m.delegate)

	// Filter and sort items according to query, sortMode, and filterType
	filtered := filterAndSortItems(m.allItems, m.searchQuery, m.sortMode, m.filterType)

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

func (m *AppModel) toggleMultiSelectMode() {
	if !m.multiSelectMode {
		m.multiSelectMode = true
		m.visualAnchor = -1
	} else {
		m.multiSelectMode = false
		m.visualAnchor = -1
		m.syncDelegateAnchor()
	}
}

func (m *AppModel) handleSpanKey() {
	if m.visualAnchor == -1 {
		idx := m.list.Index()
		m.visualAnchor = idx
		m.multiSelectMode = true
		m.syncDelegateAnchor()
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

func (m *AppModel) syncDelegateQuery() {
	m.searchQuery = m.searchInput.Value()
	m.delegate.SearchQuery = &m.searchQuery
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

// Helpers for sorting and tiered search filtering

func sortItems(items []model.AgentItem, mode model.SortMode) {
	switch mode {
	case model.SortNewest:
		sort.Slice(items, func(i, j int) bool {
			if !items[i].ModTime.Equal(items[j].ModTime) {
				return items[i].ModTime.After(items[j].ModTime)
			}
			return items[i].NameLower < items[j].NameLower
		})
	case model.SortOldest:
		sort.Slice(items, func(i, j int) bool {
			if !items[i].ModTime.Equal(items[j].ModTime) {
				return items[i].ModTime.Before(items[j].ModTime)
			}
			return items[i].NameLower < items[j].NameLower
		})
	case model.SortLargest:
		sort.Slice(items, func(i, j int) bool {
			if items[i].Tokens != items[j].Tokens {
				return items[i].Tokens > items[j].Tokens
			}
			return items[i].NameLower < items[j].NameLower
		})
	case model.SortSmallest:
		sort.Slice(items, func(i, j int) bool {
			if items[i].Tokens != items[j].Tokens {
				return items[i].Tokens < items[j].Tokens
			}
			return items[i].NameLower < items[j].NameLower
		})
	case model.SortAZ:
		sort.Slice(items, func(i, j int) bool {
			return items[i].NameLower < items[j].NameLower
		})
	case model.SortZA:
		sort.Slice(items, func(i, j int) bool {
			return items[i].NameLower > items[j].NameLower
		})
	}
}

func isSubsequence(target, query string) bool {
	if query == "" {
		return true
	}
	tRunes := []rune(strings.ToLower(target))
	qRunes := []rune(strings.ToLower(query))
	qIdx := 0
	for _, tr := range tRunes {
		if tr == qRunes[qIdx] {
			qIdx++
			if qIdx == len(qRunes) {
				return true
			}
		}
	}
	return false
}

func filterAndSortItems(items []model.AgentItem, query string, sortMode model.SortMode, filterType model.FilterType) []model.AgentItem {
	// 1. Type Filter
	var typeFiltered []model.AgentItem
	for _, it := range items {
		switch filterType {
		case model.FilterSkills:
			if it.Type == model.TypeSkill {
				typeFiltered = append(typeFiltered, it)
			}
		case model.FilterMCP:
			if it.Type == model.TypeMCP {
				typeFiltered = append(typeFiltered, it)
			}
		default: // FilterAll
			typeFiltered = append(typeFiltered, it)
		}
	}

	cleanQuery := strings.ToLower(strings.TrimSpace(query))
	if cleanQuery == "" {
		sortItems(typeFiltered, sortMode)
		return typeFiltered
	}

	tokens := strings.Fields(cleanQuery)
	seen := make(map[string]bool)

	var tier1, tier2, tier3, tier4, tier5, tier6 []model.AgentItem

	for _, it := range typeFiltered {
		it.EnsureCache()

		// Tier 1: All query words match in Name
		allTokensName := true
		for _, tok := range tokens {
			if !strings.Contains(it.NameLower, tok) {
				allTokensName = false
				break
			}
		}
		if allTokensName {
			tier1 = append(tier1, it)
			seen[it.ID] = true
			continue
		}

		// Tier 2: All query words match in Description
		allTokensDesc := true
		for _, tok := range tokens {
			if !strings.Contains(it.DescLower, tok) && !strings.Contains(it.NameLower, tok) {
				allTokensDesc = false
				break
			}
		}
		if allTokensDesc {
			tier2 = append(tier2, it)
			seen[it.ID] = true
			continue
		}

		// Tier 3: Subsequence match in Name (letters typed match in order, e.g. "anm" -> "animejs-animation")
		if isSubsequence(it.NameLower, cleanQuery) {
			tier3 = append(tier3, it)
			seen[it.ID] = true
			continue
		}

		// Tier 4: Any query token matches in Name or Description
		anyTokenMatch := false
		for _, tok := range tokens {
			if strings.Contains(it.NameLower, tok) || strings.Contains(it.DescLower, tok) {
				anyTokenMatch = true
				break
			}
		}
		if anyTokenMatch {
			tier4 = append(tier4, it)
			seen[it.ID] = true
			continue
		}

		// Tier 5: Subsequence match in Description
		if isSubsequence(it.DescLower, cleanQuery) {
			tier5 = append(tier5, it)
			seen[it.ID] = true
			continue
		}

		// Tier 6: Match in Path (token substring or subsequence)
		matchPath := false
		for _, tok := range tokens {
			if strings.Contains(it.PathLower, tok) {
				matchPath = true
				break
			}
		}
		if !matchPath && isSubsequence(it.PathLower, cleanQuery) {
			matchPath = true
		}
		if matchPath {
			tier6 = append(tier6, it)
			seen[it.ID] = true
		}
	}

	sortItems(tier1, sortMode)
	sortItems(tier2, sortMode)
	sortItems(tier3, sortMode)
	sortItems(tier4, sortMode)
	sortItems(tier5, sortMode)
	sortItems(tier6, sortMode)

	result := make([]model.AgentItem, 0, len(seen))
	result = append(result, tier1...)
	result = append(result, tier2...)
	result = append(result, tier3...)
	result = append(result, tier4...)
	result = append(result, tier5...)
	result = append(result, tier6...)
	return result
}
