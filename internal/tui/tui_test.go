package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"skillyfiy/internal/config"
	"skillyfiy/internal/model"
	"skillyfiy/internal/tui"
)

func createTestItems() []model.AgentItem {
	return []model.AgentItem{
		{
			ID:          "skill:web-search.md",
			Name:        "web-search.md",
			Type:        model.TypeSkill,
			SourcePath:  "/path/skills/web-search.md",
			Description: "Searches Google or DuckDuckGo",
			Tokens:      100,
			Size:        400,
			Details: map[string]string{
				"Category": "Agent Skill",
				"Lines":    "20",
			},
			RawPreview: "# Web Search Tool\nUseful for searching online.",
		},
		{
			ID:          "skill:code-reviewer.py",
			Name:        "code-reviewer.py",
			Type:        model.TypeSkill,
			SourcePath:  "/path/skills/code-reviewer.py",
			Description: "AST based static code reviewer",
			Tokens:      500,
			Size:        2000,
			Details: map[string]string{
				"Category": "Agent Skill",
				"Lines":    "80",
			},
			RawPreview: "import ast\n# Code reviewer tool",
		},
		{
			ID:          "mcp:github",
			Name:        "github",
			Type:        model.TypeMCP,
			SourcePath:  "/path/claude/settings.json",
			Description: "npx -y @modelcontextprotocol/server-github",
			Tokens:      250,
			Size:        1000,
			Details: map[string]string{
				"Category":  "MCP Server",
				"Command":   "npx",
				"Arguments": "-y @modelcontextprotocol/server-github",
			},
			RawPreview: `{\n  "command": "npx"\n}`,
		},
	}
}

func TestBannerHelpers(t *testing.T) {
	banner := tui.RenderBanner(80)
	if !strings.Contains(banner, "SKILLYFIY") && !strings.Contains(banner, "Optimizer") {
		t.Errorf("RenderBanner missing banner text")
	}

	compact := tui.CompactBanner()
	if !strings.Contains(compact, "SKILLYFIY") {
		t.Errorf("CompactBanner missing name")
	}

	h := tui.BannerHeight()
	if h <= 0 {
		t.Errorf("BannerHeight should be positive, got %d", h)
	}
}

func TestTUIAppInitialization(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)
	view := app.View()

	if !strings.Contains(view, "Optimizer") {
		t.Errorf("App view missing banner subtitle")
	}
	if !strings.Contains(view, "web-search.md") {
		t.Errorf("App view missing item web-search.md")
	}
}

func TestTUISelectionAndCursorStep(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Send space key to toggle selection on first item
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeySpace})
	m, ok := updated.(tui.AppModel)
	if !ok {
		t.Fatalf("Failed to cast tea.Model to AppModel")
	}

	view := m.View()
	if !strings.Contains(view, "[✓]") {
		t.Errorf("Expected checked checkbox [✓] after space toggle")
	}
}

func TestTUISelectAll(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Send Ctrl+A to select all items
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	m, ok := updated.(tui.AppModel)
	if !ok {
		t.Fatalf("Failed to cast tea.Model to AppModel")
	}

	view := m.View()
	checkedCount := strings.Count(view, "[✓]")
	if checkedCount < 3 {
		t.Errorf("Expected at least 3 [✓] checkmarks, got %d", checkedCount)
	}

	// Send Ctrl+A again to deselect all
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	m2 := updated.(tui.AppModel)
	view2 := m2.View()
	checkedCount2 := strings.Count(view2, "[✓]")
	if checkedCount2 != 0 {
		t.Errorf("Expected 0 [✓] checkmarks after second Ctrl+A, got %d", checkedCount2)
	}
}

func TestTUIVisualMode(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// 1. Enter visual mode by pressing 'v'
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m := updated.(tui.AppModel)

	view := m.View()
	if !strings.Contains(view, "VISUAL RANGE") {
		t.Errorf("Expected footer to indicate VISUAL RANGE")
	}

	// 2. Move cursor down with 'j'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(tui.AppModel)

	// 3. Confirm range with 'v'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = updated.(tui.AppModel)

	viewFinal := m.View()
	if strings.Contains(viewFinal, "VISUAL RANGE") {
		t.Errorf("Expected visual mode to be exited after second 'v'")
	}
}

func TestTUIVisualModeCancel(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Enter visual mode
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m := updated.(tui.AppModel)

	// Cancel with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := updated.(tui.AppModel)

	view := m2.View()
	if strings.Contains(view, "VISUAL RANGE") {
		t.Errorf("Expected visual mode cancelled on Esc")
	}
}

func TestTUICycleSortAndPaneSwitch(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// 1. Cycle Sort with Ctrl+S
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m := updated.(tui.AppModel)
	view := m.View()
	if !strings.Contains(view, "[Sort: Tokens ↓]") {
		t.Errorf("Expected [Sort: Tokens ↓], got: %s", view)
	}

	// Cycle sort again to ItemType
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(tui.AppModel)
	view = m.View()
	if !strings.Contains(view, "[Sort: Type]") {
		t.Errorf("Expected [Sort: Type], got: %s", view)
	}

	// 2. Switch Pane with Tab
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(tui.AppModel)

	// Scroll inspector with down arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(tui.AppModel)

	// Switch Pane back to List
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(tui.AppModel)
}

func TestTUISearchFocus(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Focus search with '/'
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m := updated.(tui.AppModel)

	// Type 'git' into search
	for _, r := range "git" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(tui.AppModel)
	}

	view := m.View()
	// Should show github item and hide web-search
	if !strings.Contains(view, "github") {
		t.Errorf("Expected github in filtered search view")
	}

	// Exit search with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tui.AppModel)
}

func TestInspectorModel(t *testing.T) {
	insp := tui.NewInspector(60, 20)
	items := createTestItems()

	// 1. Set Skill Item
	insp.SetItem(&items[0])
	view := insp.View()
	if !strings.Contains(view, "[AGENT SKILL]") {
		t.Errorf("Inspector view missing [AGENT SKILL] badge")
	}
	if !strings.Contains(view, "web-search.md") {
		t.Errorf("Inspector view missing item name")
	}

	// 2. Set MCP Item
	insp.SetItem(&items[2])
	viewMCP := insp.View()
	if !strings.Contains(viewMCP, "[MCP SERVER]") {
		t.Errorf("Inspector view missing [MCP SERVER] badge")
	}

	// 3. Set nil Item
	insp.SetItem(nil)
	viewNil := insp.View()
	if !strings.Contains(viewNil, "No item selected") {
		t.Errorf("Inspector view missing empty state message")
	}

	// 4. SetSize
	insp.SetSize(80, 25)
	if insp.Width != 80 || insp.Height != 25 {
		t.Errorf("SetSize did not update dimensions")
	}
}

func TestTUIConfirmationModal(t *testing.T) {
	items := createTestItems()
	items[0].Selected = true

	modal := tui.RenderConfirmModal(items, 100, 30, true)

	if !strings.Contains(modal, "CONFIRM") && !strings.Contains(modal, "DRY-RUN") {
		t.Errorf("Modal missing title")
	}
	if !strings.Contains(modal, "Skills to permanently unlink:") {
		t.Errorf("Modal missing skills label")
	}
	if !strings.Contains(modal, "Are you sure you want to purge these resources?") {
		t.Errorf("Modal missing prompt question")
	}

	// Test non-dryrun modal
	modalLive := tui.RenderConfirmModal(items, 100, 30, false)
	if !strings.Contains(modalLive, "IRREVERSIBLE PURGE") {
		t.Errorf("Live modal missing irreversible warning")
	}
}
