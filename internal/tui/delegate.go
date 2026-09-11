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
	SearchQuery  *string
}

// NewAgentItemDelegate creates an item delegate with references to visual anchor and search query.
func NewAgentItemDelegate(visualAnchor *int, searchQuery *string) AgentItemDelegate {
	return AgentItemDelegate{
		VisualAnchor: visualAnchor,
		SearchQuery:  searchQuery,
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

// Render renders a single item with checkbox, badge, title, tokens, and description with search highlighting.
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

	// 5. Title with Search Highlighting
	titleText := item.Name
	overhead := 2 + 3 + 1 + 7 + 1 + len(fmt.Sprintf("~%dtk", item.Tokens)) + 2
	availTitleWidth := width - overhead
	if availTitleWidth < 6 {
		availTitleWidth = 6
	}
	titleTrunc := truncate.StringWithTail(titleText, uint(availTitleWidth), "…")

	var queryStr string
	if d.SearchQuery != nil {
		queryStr = *d.SearchQuery
	}

	baseTitleStyle := NormalTitleStyle
	if item.Selected {
		baseTitleStyle = SelectedTitleStyle
	} else if isCurrent {
		baseTitleStyle = CursorTitleStyle
	}

	titleStyled := HighlightMatches(titleTrunc, queryStr, baseTitleStyle, HighlightMatchStyle)
	line1 := fmt.Sprintf("%s%s %s %s %s", cursorStr, checkboxStr, badgeStr, titleStyled, tokenStr)

	// Line 2: Description with Search Highlighting
	descIndent := "      " // Align under item title
	availDescWidth := width - len(descIndent) - 2
	if availDescWidth < 10 {
		availDescWidth = 10
	}
	cleanDesc := strings.ReplaceAll(item.Description, "\n", " ")
	cleanDesc = strings.TrimSpace(cleanDesc)
	descTrunc := truncate.StringWithTail(cleanDesc, uint(availDescWidth), "…")
	descStyled := HighlightMatches(descTrunc, queryStr, NormalDescStyle, HighlightMatchStyle)
	line2 := descIndent + descStyled

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

// HighlightMatches highlights matched words or letters from query within target string.
func HighlightMatches(target, query string, baseStyle, matchStyle lipgloss.Style) string {
	cleanQuery := strings.ToLower(strings.TrimSpace(query))
	if cleanQuery == "" || target == "" {
		return baseStyle.Render(target)
	}

	targetRunes := []rune(target)
	targetLower := []rune(strings.ToLower(target))
	matched := make([]bool, len(targetRunes))

	words := strings.Fields(cleanQuery)
	for _, w := range words {
		wRunes := []rune(w)
		if len(wRunes) == 0 {
			continue
		}
		wordMatched := false
		for i := 0; i <= len(targetLower)-len(wRunes); i++ {
			found := true
			for k := 0; k < len(wRunes); k++ {
				if targetLower[i+k] != wRunes[k] {
					found = false
					break
				}
			}
			if found {
				wordMatched = true
				for k := 0; k < len(wRunes); k++ {
					matched[i+k] = true
				}
			}
		}

		// If this word did not match contiguously, match its individual letters in sequence
		if !wordMatched {
			wIdx := 0
			for i, tr := range targetLower {
				if !matched[i] && wIdx < len(wRunes) && tr == wRunes[wIdx] {
					matched[i] = true
					wIdx++
				}
			}
		}
	}

	// Full query subsequence fallback if nothing matched
	anyMatched := false
	for _, m := range matched {
		if m {
			anyMatched = true
			break
		}
	}
	if !anyMatched {
		qRunes := []rune(cleanQuery)
		qIdx := 0
		for i, tr := range targetLower {
			if qIdx < len(qRunes) && tr == qRunes[qIdx] {
				matched[i] = true
				qIdx++
			}
		}
	}

	// Group contiguous segments and render
	var b strings.Builder
	segStart := 0
	isMatch := matched[0]

	for i := 1; i < len(targetRunes); i++ {
		if matched[i] != isMatch {
			chunk := string(targetRunes[segStart:i])
			if isMatch {
				b.WriteString(matchStyle.Render(chunk))
			} else {
				b.WriteString(baseStyle.Render(chunk))
			}
			segStart = i
			isMatch = matched[i]
		}
	}

	chunk := string(targetRunes[segStart:])
	if isMatch {
		b.WriteString(matchStyle.Render(chunk))
	} else {
		b.WriteString(baseStyle.Render(chunk))
	}

	return b.String()
}
