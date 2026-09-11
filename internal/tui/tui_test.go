package tui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"skillyfiy/internal/config"
	"skillyfiy/internal/model"
	"skillyfiy/internal/tui"
)

func createTestItems() []model.AgentItem {
	now := time.Now()
	return []model.AgentItem{
		{
			ID:          "skill:web-search.md",
			Name:        "web-search.md",
			Type:        model.TypeSkill,
			SourcePath:  "/path/skills/web-search.md",
			Description: "Searches Google or DuckDuckGo",
			Tokens:      100,
			Size:        400,
			ModTime:     now.Add(-2 * time.Hour),
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
			ModTime:     now.Add(-1 * time.Hour),
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
			ModTime:     now,
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
	// Initial sort is Newest
	view := app.View()
	if !strings.Contains(view, "[Sort: Newest]") && !strings.Contains(view, "[Newest]") {
		t.Errorf("Expected initial [Sort: Newest], got: %s", view)
	}

	// 1. Cycle Sort with Ctrl+S -> Oldest
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m := updated.(tui.AppModel)
	view = m.View()
	if !strings.Contains(view, "[Sort: Oldest]") && !strings.Contains(view, "[Oldest]") {
		t.Errorf("Expected [Sort: Oldest], got: %s", view)
	}

	// 2. Cycle Sort again -> Largest
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(tui.AppModel)
	view = m.View()
	if !strings.Contains(view, "[Sort: Largest]") && !strings.Contains(view, "[Largest]") {
		t.Errorf("Expected [Sort: Largest], got: %s", view)
	}

	// 3. Switch Pane with Tab
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(tui.AppModel)

	// Scroll inspector with down arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(tui.AppModel)

	// Switch Pane back to List
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(tui.AppModel)
}

func TestTUICycleFilter(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)
	view := app.View()
	if !strings.Contains(view, "[Filter: All]") && !strings.Contains(view, "[All]") {
		t.Errorf("Expected initial [Filter: All], got: %s", view)
	}

	// 1. Cycle Filter with Ctrl+T -> Skills
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m := updated.(tui.AppModel)
	view = m.View()
	if !strings.Contains(view, "[Filter: Skills]") && !strings.Contains(view, "[Skills]") {
		t.Errorf("Expected [Filter: Skills], got: %s", view)
	}
	// MCP item github should NOT be visible
	if strings.Contains(view, "github") {
		t.Errorf("Expected github MCP to be filtered out in Skills filter")
	}

	// 2. Cycle Filter again -> MCP
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	m = updated.(tui.AppModel)
	view = m.View()
	if !strings.Contains(view, "[Filter: MCP]") && !strings.Contains(view, "[MCP]") {
		t.Errorf("Expected [Filter: MCP], got: %s", view)
	}
	// Skill items should NOT be visible
	if strings.Contains(view, "web-search.md") {
		t.Errorf("Expected web-search.md to be filtered out in MCP filter")
	}
}

func TestTUIVisualModeWithSKey(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Enter visual mode with 's'
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updated.(tui.AppModel)

	view := m.View()
	if !strings.Contains(view, "VISUAL RANGE") {
		t.Errorf("Expected footer to indicate VISUAL RANGE after 's'")
	}

	// Move cursor down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(tui.AppModel)

	// Commit with 's'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(tui.AppModel)

	viewFinal := m.View()
	if strings.Contains(viewFinal, "VISUAL RANGE") {
		t.Errorf("Expected visual range exited after committing with 's'")
	}
}

func TestTUIMobileLayout(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Simulate narrow mobile screen (60x24)
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m := updated.(tui.AppModel)

	view := m.View()
	lines := strings.Split(view, "\n")
	for i, l := range lines {
		// Verify no line exceeds terminal visual width (eliminates auto-wrap)
		w := lipgloss.Width(l)
		if w > 60 {
			t.Errorf("Line %d exceeds mobile width 60: visual width=%d", i, w)
		}
	}
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

func TestTUITieredSearch(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// 1. Focus search
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m := updated.(tui.AppModel)

	// 2. Type "reviewer" (Tier 1 Name match)
	for _, r := range "reviewer" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(tui.AppModel)
	}

	view := m.View()
	if !strings.Contains(view, "code-reviewer.py") {
		t.Errorf("Expected code-reviewer.py in Tier 1 search match")
	}
	if strings.Contains(view, "web-search.md") {
		t.Errorf("web-search.md should not match 'reviewer'")
	}

	// 3. Clear search input and search "google" (Tier 2 Description match)
	for i := 0; i < 15; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = updated.(tui.AppModel)
	}
	for _, r := range "google" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(tui.AppModel)
	}

	viewDesc := m.View()
	if !strings.Contains(viewDesc, "web-search.md") {
		t.Errorf("Expected web-search.md to match 'google' in Tier 2 description")
	}

	// 4. Clear search and test Tier 3 Path match ("claude")
	for i := 0; i < 15; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = updated.(tui.AppModel)
	}
	for _, r := range "claude" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(tui.AppModel)
	}

	viewPath := m.View()
	if !strings.Contains(viewPath, "github") {
		t.Errorf("Expected github MCP to match 'claude' in Tier 3 path")
	}
}

func TestTUISortingCriteria(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Initial is SortNewest
	view := app.View()
	if !strings.Contains(view, "Newest") {
		t.Errorf("Expected initial sort to be Newest")
	}

	expectedSortBadges := []string{
		"Oldest",
		"Largest",
		"Smallest",
		"A-Z",
		"Z-A",
		"Newest",
	}

	m := app
	for _, expected := range expectedSortBadges {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
		m = updated.(tui.AppModel)
		v := m.View()
		if !strings.Contains(v, expected) {
			t.Errorf("Expected badge %s in view, got: %s", expected, v)
		}
	}
}

func TestTUIDefensiveClamping(t *testing.T) {
	cfg := &config.Config{DryRun: true}
	items := createTestItems()

	app := tui.NewAppModel(cfg, items)

	// Window too small
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 20, Height: 8})
	m := updated.(tui.AppModel)

	view := m.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("Expected 'Terminal too small' message on tiny dimensions")
	}
}
