package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"

	"skillyfiy/internal/model"
)

// AgentItemDelegate implements list.ItemDelegate for rendering AgentItems.
type AgentItemDelegate struct {
	VisualAnchor *int
}

// NewAgentItemDelegate creates an item delegate with a reference to the visual anchor.
func NewAgentItemDelegate(visualAnchor *int) AgentItemDelegate {
	return AgentItemDelegate{
		VisualAnchor: visualAnchor,
	}
}

// Height returns the vertical line height per item.
func (d AgentItemDelegate) Height() int {
	return 2
}

// Spacing returns the vertical gap between items.
func (d AgentItemDelegate) Spacing() int {
	return 0
}

// Update handles item-level updates.
func (d AgentItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders a single item with checkbox, badge, title, tokens, and description.
func (d AgentItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(model.AgentItem)
	if !ok {
		return
	}

	width := m.Width()
	if width <= 0 {
		width = 40
	}

	isCurrent := index == m.Index()
	inVisualRange := false
	if d.VisualAnchor != nil && *d.VisualAnchor >= 0 {
		start := *d.VisualAnchor
		end := m.Index()
		if start > end {
			start, end = end, start
		}
		if index >= start && index <= end {
			inVisualRange = true
		}
	}

	// 1. Checkbox
	checkboxStr := UncheckedBoxStyle.Render("[ ]")
	if item.Selected {
		checkboxStr = CheckedBoxStyle.Render("[✓]")
	} else if inVisualRange {
		checkboxStr = VisualHighlightStyle.Render("[~]")
	}

	// 2. Cursor Indicator
	cursorStr := "  "
	if isCurrent {
		cursorStr = SearchPromptStyle.Render("❯ ")
	}

	// 3. Category Badge
	badgeStr := SkillBadgeStyle.Render("SKILL")
	if item.Type == model.TypeMCP {
		badgeStr = MCPBadgeStyle.Render(" MCP ")
	}

	// 4. Token Weight
	tokenStr := TokenWeightStyle.Render(fmt.Sprintf("~%dtk", item.Tokens))

	// 5. Title
	titleText := item.Name
	// Calculate available width for title
	// cursor(2) + checkbox(3) + space(1) + badge(7) + space(1) + title + space(1) + tokenStr(len)
	overhead := 2 + 3 + 1 + 7 + 1 + len(fmt.Sprintf("~%dtk", item.Tokens)) + 2
	availTitleWidth := width - overhead
	if availTitleWidth < 6 {
		availTitleWidth = 6
	}
	titleTrunc := truncate.StringWithTail(titleText, uint(availTitleWidth), "…")

	titleStyled := NormalTitleStyle.Render(titleTrunc)
	if item.Selected {
		titleStyled = SelectedTitleStyle.Render(titleTrunc)
	} else if isCurrent {
		titleStyled = CursorTitleStyle.Render(titleTrunc)
	}

	line1 := fmt.Sprintf("%s%s %s %s %s", cursorStr, checkboxStr, badgeStr, titleStyled, tokenStr)

	// Line 2: Description
	descIndent := "      " // Align under item title
	availDescWidth := width - len(descIndent) - 2
	if availDescWidth < 10 {
		availDescWidth = 10
	}
	cleanDesc := strings.ReplaceAll(item.Description, "\n", " ")
	cleanDesc = strings.TrimSpace(cleanDesc)
	descTrunc := truncate.StringWithTail(cleanDesc, uint(availDescWidth), "…")
	line2 := descIndent + NormalDescStyle.Render(descTrunc)

	if inVisualRange && !isCurrent {
		line1 = VisualHighlightStyle.Render(line1)
		line2 = VisualHighlightStyle.Render(line2)
	}

	if lipgloss.Width(line1) > width {
		line1 = truncate.StringWithTail(line1, uint(width), "")
	}
	if lipgloss.Width(line2) > width {
		line2 = truncate.StringWithTail(line2, uint(width), "")
	}

	fmt.Fprintf(w, "%s\n%s", line1, line2)
}
