package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"skillyfiy/internal/config"
	"skillyfiy/internal/model"
)

func internalTestItems() []model.AgentItem {
	return []model.AgentItem{
		{
			ID:          "skill:alpha.md",
			Name:        "alpha.md",
			Type:        model.TypeSkill,
			SourcePath:  "/skills/alpha.md",
			Description: "first skill",
			Tokens:      100,
			Size:        400,
			ModTime:     time.Now(),
			Details:     map[string]string{"Category": "Agent Skill"},
			RawPreview:  "# alpha",
		},
		{
			ID:          "mcp:beta",
			Name:        "beta",
			Type:        model.TypeMCP,
			SourcePath:  "/settings.json",
			Description: "beta server",
			Tokens:      250,
			Size:        1000,
			ModTime:     time.Now(),
			Details:     map[string]string{"Category": "MCP Server"},
			RawPreview:  "{}",
		},
	}
}

// runningApp builds a model with a known-good, env-independent clock.
func runningApp(t *testing.T) AppModel {
	t.Helper()
	app := NewAppModel(&config.Config{}, internalTestItems())
	app.animUnsupported = false
	app.anim.SetMode(AnimNormal)
	return app
}

func TestTickCmdSchedulesWhileRunning(t *testing.T) {
	app := runningApp(t)
	cmd := app.tickCmd()
	if cmd == nil {
		t.Fatalf("expected a tick command while the animation is running")
	}
	if !app.anim.tickInFlight {
		t.Errorf("scheduling a frame should mark it in flight")
	}
}

// TestTickCmdDropsFramesWhenBehind is the back-pressure contract: a frame that is
// still in flight must never queue a second one behind it.
func TestTickCmdDropsFramesWhenBehind(t *testing.T) {
	app := runningApp(t)

	if cmd := app.tickCmd(); cmd == nil {
		t.Fatalf("expected the first frame to be scheduled")
	}
	if cmd := app.tickCmd(); cmd != nil {
		t.Errorf("a duplicate frame should be dropped while one is in flight")
	}

	// Delivering the frame clears the guard and lets the clock continue.
	updated, cmd := app.Update(animTickMsg{tick: 0})
	app = updated.(AppModel)
	if cmd == nil {
		t.Errorf("expected the clock to continue after a frame was delivered")
	}
	if app.anim.tickInFlight == false {
		t.Errorf("a delivered frame should immediately schedule its successor")
	}
}

func TestTickCmdReturnsNilWhenDisabled(t *testing.T) {
	app := runningApp(t)
	app.anim.SetMode(AnimOff)

	if cmd := app.tickCmd(); cmd != nil {
		t.Errorf("an OFF clock must not schedule frames")
	}
	if app.Init() != nil {
		t.Errorf("Init must not schedule frames when the animation is off")
	}
}

func TestTickCmdReturnsNilWhenUnsupported(t *testing.T) {
	app := runningApp(t)
	app.animUnsupported = true

	if cmd := app.tickCmd(); cmd != nil {
		t.Errorf("an unsupported terminal must not schedule frames")
	}
}

// TestAnimationPausesOnModals covers both overlays. While either is open the clock
// stops, and the phase must not advance while it is stopped.
func TestAnimationPausesOnModals(t *testing.T) {
	cases := []struct {
		name string
		open func(m *AppModel)
	}{
		{"help overlay", func(m *AppModel) { m.showHelpOverlay = true }},
		{"confirm modal", func(m *AppModel) { m.state = stateConfirm }},
	}

	for _, tc := range cases {
		app := runningApp(t)
		for i := 0; i < 10; i++ {
			app.anim.Advance()
		}
		frozen := app.anim.Phase()

		tc.open(&app)
		if !app.animPaused() {
			t.Fatalf("%s: expected the animation to report paused", tc.name)
		}

		updated, cmd := app.Update(animTickMsg{})
		app = updated.(AppModel)

		if cmd != nil {
			t.Errorf("%s: expected no further frames while paused", tc.name)
		}
		if app.anim.Phase() != frozen {
			t.Errorf("%s: phase advanced from %v to %v while paused", tc.name, frozen, app.anim.Phase())
		}
		if app.anim.tickInFlight {
			t.Errorf("%s: the in-flight guard should be cleared", tc.name)
		}
	}
}

// TestAnimationResumesFromFrozenPhase proves the pause is seamless: the sweep picks
// up exactly where it stopped instead of jumping.
func TestAnimationResumesFromFrozenPhase(t *testing.T) {
	app := runningApp(t)
	for i := 0; i < 25; i++ {
		app.anim.Advance()
	}
	frozen := app.anim.Phase()

	// Open the help overlay, then deliver a frame that must be swallowed.
	app.showHelpOverlay = true
	app.Update(animTickMsg{})
	if app.anim.Phase() != frozen {
		t.Fatalf("phase moved while the overlay was open")
	}

	updated, cmd := app.Update(asciiKey('i'))
	app = updated.(AppModel)
	if app.showHelpOverlay {
		t.Fatalf("expected the help overlay to close")
	}
	if cmd == nil {
		t.Errorf("closing the overlay should restart the clock")
	}
	if app.anim.Phase() != frozen {
		t.Errorf("closing the overlay jumped the phase from %v to %v", frozen, app.anim.Phase())
	}
}

func TestConfirmCancelRestartsClock(t *testing.T) {
	app := runningApp(t)
	app.state = stateConfirm

	updated, cmd := app.Update(asciiKey('n'))
	app = updated.(AppModel)

	if app.state != stateMain {
		t.Fatalf("expected the confirm modal to close")
	}
	if cmd == nil {
		t.Errorf("cancelling the modal should restart the animation clock")
	}
}

func TestRunCommandRespectsUnsupportedTerminal(t *testing.T) {
	app := runningApp(t)
	app.animUnsupported = true
	before := app.anim.Mode()

	cmd := app.runCommand("/anim")
	if cmd == nil {
		t.Fatalf("expected a toast for an unsupported terminal")
	}
	if app.anim.Mode() != before {
		t.Errorf("an unsupported terminal must not change the animation mode")
	}
	if !strings.Contains(app.toast, "unavailable") {
		t.Errorf("expected an unavailable toast, got %q", app.toast)
	}
}

func TestAnimCommandRestartsClockFromOff(t *testing.T) {
	app := NewAppModel(&config.Config{NoAnimation: true}, internalTestItems())
	if app.anim.Mode() != AnimOff {
		t.Fatalf("--no-animation should start OFF, got %s", app.anim.Mode().Label())
	}

	cmd := app.runCommand("/anim")
	if app.anim.Mode() != AnimSlow {
		t.Errorf("expected OFF -> SLOW, got %s", app.anim.Mode().Label())
	}
	if cmd == nil {
		t.Errorf("cycling the animation on should schedule work")
	}
	if !app.anim.tickInFlight {
		t.Errorf("cycling on should start the clock immediately")
	}
}

func TestToastExpiresAndIgnoresStaleTimers(t *testing.T) {
	app := runningApp(t)
	app.setToast("Animation: FAST")
	if app.toast == "" {
		t.Fatalf("expected a toast to be set")
	}

	firstSeq := app.toastSeq

	// A second toast supersedes the first, so the first timer must be ignored.
	app.setToast("Animation: HUE")
	updated, _ := app.Update(toastMsg{seq: firstSeq})
	app = updated.(AppModel)
	if app.toast != "Animation: HUE" {
		t.Errorf("a stale toast timer cleared the current toast: %q", app.toast)
	}

	updated, _ = app.Update(toastMsg{seq: app.toastSeq})
	app = updated.(AppModel)
	if app.toast != "" {
		t.Errorf("the toast should have expired, got %q", app.toast)
	}
}

func TestUnsupportedTerminalEnvironmentForcesOff(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	app := NewAppModel(&config.Config{}, internalTestItems())

	if app.anim.Mode() != AnimOff {
		t.Errorf("NO_COLOR should start the animation off, got %s", app.anim.Mode().Label())
	}
	if !app.animUnsupported {
		t.Errorf("NO_COLOR should mark the terminal as unsupported")
	}
	if cmd := app.Init(); cmd != nil {
		t.Errorf("no frames should be scheduled on an unsupported terminal")
	}
}

func TestNoAnimationFlagIsNotReportedUnsupported(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	app := NewAppModel(&config.Config{NoAnimation: true}, internalTestItems())
	if app.animUnsupported {
		t.Errorf("--no-animation is a preference, not an unsupported terminal")
	}
	if app.anim.Mode() != AnimOff {
		t.Errorf("--no-animation should start OFF")
	}
}

// asciiKey builds the KeyMsg a plain letter key produces.
func asciiKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}
