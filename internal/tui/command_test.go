package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"skillyfiy/internal/config"
	"skillyfiy/internal/tui"
)

func TestParseCommand(t *testing.T) {
	type parseCase struct {
		input  string
		name   string
		arg    string
		wantOK bool
	}

	cases := []parseCase{
		{"/anim", "anim", "", true},
		{"anim", "anim", "", true},
		{"  /anim  ", "anim", "", true},
		{"/ANIM", "anim", "", true},
		{"/anim fast", "anim", "fast", true},
		{"/anim  slow  extra", "anim", "slow extra", true},
		{"", "", "", false},
		{"   ", "", "", false},
		{"/", "", "", false},
		{"  /  ", "", "", false},
	}

	for _, tc := range cases {
		name, arg, ok := tui.ParseCommand(tc.input)
		if ok != tc.wantOK {
			t.Errorf("ParseCommand(%q) ok = %v, want %v", tc.input, ok, tc.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if name != tc.name || arg != tc.arg {
			t.Errorf("ParseCommand(%q) = (%q, %q), want (%q, %q)", tc.input, name, arg, tc.name, tc.arg)
		}
	}
}

// enterCommandMode drives the Ctrl+K shortcut and returns the updated model.
func enterCommandMode(t *testing.T, app tui.AppModel) tui.AppModel {
	t.Helper()
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m, ok := updated.(tui.AppModel)
	if !ok {
		t.Fatalf("Update did not return an AppModel")
	}
	return m
}

// typeText feeds each rune to the model as a key press.
func typeText(t *testing.T, app tui.AppModel, text string) tui.AppModel {
	t.Helper()
	m := app
	for _, r := range text {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		var ok bool
		m, ok = updated.(tui.AppModel)
		if !ok {
			t.Fatalf("Update did not return an AppModel")
		}
	}
	return m
}

// pressEnter runs the currently typed command.
func pressEnter(t *testing.T, app tui.AppModel) (tui.AppModel, tea.Cmd) {
	t.Helper()
	updated, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, ok := updated.(tui.AppModel)
	if !ok {
		t.Fatalf("Update did not return an AppModel")
	}
	return m, cmd
}

// runCommand types a command into command mode and executes it.
func runCommand(t *testing.T, app tui.AppModel, command string) tui.AppModel {
	t.Helper()
	m := typeText(t, enterCommandMode(t, app), command)
	m, cmd := pressEnter(t, m)
	if cmd == nil {
		t.Fatalf("running %q produced no command (expected a toast timer)", command)
	}
	return m
}

func TestCommandModeOpensViaShortcut(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())
	m := enterCommandMode(t, app)

	view := m.View()
	if !strings.Contains(view, "type a command") {
		t.Errorf("expected the command placeholder in command mode, got: %s", view)
	}
	if !strings.Contains(view, "COMMAND MODE") {
		t.Errorf("expected the footer to announce COMMAND MODE")
	}
}

func TestCommandModeSuspendsFiltering(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())
	before := app.View()

	m := typeText(t, enterCommandMode(t, app), "/anim")
	view := m.View()

	// "/anim" must not be treated as a filter, so every item stays visible.
	for _, item := range []string{"web-search.md", "code-reviewer.py", "github"} {
		if !strings.Contains(view, item) {
			t.Errorf("command text filtered the list: %q disappeared", item)
		}
	}
	for _, item := range []string{"web-search.md", "code-reviewer.py", "github"} {
		if !strings.Contains(before, item) {
			t.Fatalf("test setup problem: %q was not visible before command mode", item)
		}
	}
}

func TestAnimCommandCyclesEveryMode(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())

	// The tool launches in NORMAL, so the first /anim moves to FAST.
	for _, label := range []string{"FAST", "HUE", "OFF", "SLOW", "NORMAL"} {
		m := runCommand(t, app, "/anim")
		view := m.View()
		if !strings.Contains(view, "Animation: "+label) {
			t.Errorf("expected the footer toast %q, got footer: %q",
				"Animation: "+label, lastNonEmptyLine(view))
		}
		app = m
	}
}

func TestAnimCommandAcceptsBareNameToo(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())

	withSlash := runCommand(t, app, "/anim")
	withoutSlash := runCommand(t, app, "anim")

	if !strings.Contains(withSlash.View(), "Animation: FAST") {
		t.Errorf("/anim did not advance the mode")
	}
	if !strings.Contains(withoutSlash.View(), "Animation: HUE") {
		t.Errorf("bare 'anim' did not advance the mode a second time")
	}
}

func TestUnknownCommandIsRejectedAndKeepsMode(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())

	m := runCommand(t, app, "/bogus")
	if !strings.Contains(m.View(), "Unknown command: /bogus") {
		t.Errorf("expected an unknown-command toast, got: %q", lastNonEmptyLine(m.View()))
	}

	// The mode must be untouched: the next /anim still lands on FAST.
	m2 := runCommand(t, m, "/anim")
	if !strings.Contains(m2.View(), "Animation: FAST") {
		t.Errorf("an unknown command changed the animation mode")
	}
}

func TestAnimCommandRejectsArguments(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())

	m := runCommand(t, app, "/anim fast")
	if !strings.Contains(m.View(), "Use bare /anim") {
		t.Errorf("expected an argument-rejection toast, got: %q", lastNonEmptyLine(m.View()))
	}

	// Still on the launch default, so /anim works normally afterwards.
	m2 := runCommand(t, m, "/anim")
	if !strings.Contains(m2.View(), "Animation: FAST") {
		t.Errorf("/anim with an argument changed the mode anyway")
	}
}

func TestEmptyCommandIsRejected(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())

	m, cmd := pressEnter(t, enterCommandMode(t, app))
	if cmd == nil {
		t.Fatalf("expected a toast command for an empty command")
	}
	if !strings.Contains(m.View(), "No command entered") {
		t.Errorf("expected an empty-command toast, got: %q", lastNonEmptyLine(m.View()))
	}
}

func TestEscLeavesCommandModeAndRestoresFiltering(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())

	m := typeText(t, enterCommandMode(t, app), "/anim")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tui.AppModel)

	view := m.View()
	if strings.Contains(view, "COMMAND MODE") {
		t.Errorf("Esc should have left command mode")
	}
	if strings.Contains(view, "/anim") {
		t.Errorf("Esc should have cleared the command text")
	}
	// Focus must be back on the list: space toggles an item.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(tui.AppModel)
	if !strings.Contains(m.View(), "[✓]") {
		t.Errorf("expected space to select an item after leaving command mode")
	}
}

func TestCtrlKTogglesCommandModeOff(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())
	m := enterCommandMode(t, app)
	m = enterCommandMode(t, m)

	if strings.Contains(m.View(), "COMMAND MODE") {
		t.Errorf("Ctrl+K again should leave command mode")
	}
}

func TestUnsupportedTerminalBlocksAnimCommand(t *testing.T) {
	// NO_COLOR opts the terminal out, so the command must refuse to cycle.
	t.Setenv("NO_COLOR", "1")
	unsupported := tui.NewAppModel(&config.Config{}, createTestItems())

	view := runCommand(t, unsupported, "/anim").View()
	if !strings.Contains(view, "Animation unavailable") {
		t.Errorf("expected an unavailable toast, got: %q", lastNonEmptyLine(view))
	}
	if strings.Contains(view, "Animation: SLOW") {
		t.Errorf("an unsupported terminal must not be cycled into an animation mode")
	}

	// The same terminal with the opt-out removed does cycle, which proves the
	// guard is what blocked it rather than the command being broken.
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	supported := tui.NewAppModel(&config.Config{}, createTestItems())
	if !strings.Contains(runCommand(t, supported, "/anim").View(), "Animation: FAST") {
		t.Errorf("a supported terminal should still cycle animation modes")
	}
}

func TestNoAnimationFlagStartsStatic(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{NoAnimation: true}, createTestItems())

	if cmd := app.Init(); cmd != nil {
		t.Errorf("--no-animation should not start the animation clock")
	}

	// It must remain possible to switch it back on.
	m := runCommand(t, app, "/anim")
	if !strings.Contains(m.View(), "Animation: SLOW") {
		t.Errorf("expected OFF -> SLOW when cycling from a static start, got: %q",
			lastNonEmptyLine(m.View()))
	}
}

func TestAnimationStartsOnByDefault(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := tui.NewAppModel(&config.Config{}, createTestItems())
	if cmd := app.Init(); cmd == nil {
		t.Errorf("the animation should start automatically without any key press")
	}
}

// lastNonEmptyLine returns the footer, which is where toasts are rendered.
func lastNonEmptyLine(view string) string {
	lines := strings.Split(view, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}
