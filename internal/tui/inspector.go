package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"skillyfiy/internal/model"
)

// InspectorModel wraps a bubbles/viewport to display detailed metadata and source previews.
type InspectorModel struct {
	Viewport viewport.Model
	Item     *model.AgentItem
	Width    int
	Height   int
	Focused  bool
}

// NewInspector creates a new InspectorModel with the given dimensions.
func NewInspector(width, height int) InspectorModel {
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle()

	return InspectorModel{
		Viewport: vp,
		Width:    width,
		Height:   height,
		Focused:  false,
	}
}

// SetSize updates the viewport dimensions and reformats existing content.
func (i *InspectorModel) SetSize(width, height int) {
	i.Width = width
	i.Height = height
	i.Viewport.Width = width
	i.Viewport.Height = height
	if i.Item != nil {
		i.SetItem(i.Item)
	}
}

// SetItem updates the inspector content with the details of the active AgentItem.
func (i *InspectorModel) SetItem(item *model.AgentItem) {
	i.Item = item
	if item == nil {
		i.Viewport.SetContent(NormalDescStyle.Render("No item selected or list is empty."))
		return
	}

	content := i.renderContent(item)
	i.Viewport.SetContent(content)
}

// Update delegates messages to the internal viewport.
func (i *InspectorModel) Update(msg tea.Msg) (InspectorModel, tea.Cmd) {
	var cmd tea.Cmd
	i.Viewport, cmd = i.Viewport.Update(msg)
	return *i, cmd
}

// View renders the inspector pane with appropriate focused/unfocused border styling.
func (i InspectorModel) View() string {
	style := RightPaneStyle
	if i.Focused {
		style = RightPaneFocusedStyle
	}

	inner := i.Viewport.View()
	return style.Width(i.Width).Height(i.Height).Render(inner)
}

// renderContent generates the formatted inspection text for skills or MCP items.
func (i *InspectorModel) renderContent(item *model.AgentItem) string {
	var b strings.Builder

	// 1. Header with Badge & Name
	var badge string
	if item.Type == model.TypeSkill {
		badge = SkillBadgeStyle.Render(" [AGENT SKILL] ")
	} else {
		badge = MCPBadgeStyle.Render(" [MCP SERVER] ")
	}

	header := fmt.Sprintf("%s  %s", badge, InspectorHeaderStyle.Render(item.Name))
	b.WriteString(header + "\n\n")

	// 2. Structured Metadata
	b.WriteString(InspectorKeyStyle.Render("Resource ID:") + InspectorValStyle.Render(item.ID) + "\n")
	b.WriteString(InspectorKeyStyle.Render("Disk Location:") + InspectorValStyle.Render(item.SourcePath) + "\n")
	b.WriteString(InspectorKeyStyle.Render("File Size:") + InspectorValStyle.Render(fmt.Sprintf("%d bytes", item.Size)) + "\n")
	b.WriteString(InspectorKeyStyle.Render("Context Cost:") + TokenWeightStyle.Render(fmt.Sprintf("~%d prompt tokens", item.Tokens)) + "\n")

	// Sorted custom details
	keys := make([]string, 0, len(item.Details))
	for k := range item.Details {
		if k == "Path" || k == "Size" || k == "Tokens" || k == "Category" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := item.Details[k]
		if val != "" {
			b.WriteString(InspectorKeyStyle.Render(k+":") + InspectorValStyle.Render(val) + "\n")
		}
	}

	b.WriteString("\n" + SubtitleStyle.Render(strings.Repeat("─", i.Width-4)) + "\n\n")

	// 3. Source Preview
	if item.Type == model.TypeSkill {
		b.WriteString(SearchPromptStyle.Render("▶ SOURCE PREVIEW (First 25 lines):") + "\n")
	} else {
		b.WriteString(SearchPromptStyle.Render("▶ MCP SERVER DEFINITION:") + "\n")
	}

	previewText := item.RawPreview
	if previewText == "" {
		previewText = "(Preview unavailable)"
	}

	// Wrap preview to avoid horizontal overrun
	wrapped := wordwrap.String(previewText, i.Width-6)
	b.WriteString(InspectorCodeBlock.Render(wrapped))

	return b.String()
}
