package tui_test

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"skillyfiy/internal/tui"
)

const phaseTolerance = 1e-6

// phaseDistanceToZero reports how close a wrapped phase is to the start of a loop.
func phaseDistanceToZero(p float64) float64 {
	d := math.Abs(p)
	if alt := math.Abs(1 - p); alt < d {
		d = alt
	}
	return d
}

func TestAnimModeCycleOrder(t *testing.T) {
	want := []tui.AnimMode{tui.AnimOff, tui.AnimSlow, tui.AnimNormal, tui.AnimFast, tui.AnimHue, tui.AnimOff}

	got := tui.AnimOff
	for i, expected := range want {
		if got != expected {
			t.Fatalf("cycle step %d: got %v, want %v", i, got.Label(), expected.Label())
		}
		got = got.Next()
	}
}

func TestAnimModeLabels(t *testing.T) {
	labels := map[tui.AnimMode]string{
		tui.AnimOff:    "OFF",
		tui.AnimSlow:   "SLOW",
		tui.AnimNormal: "NORMAL",
		tui.AnimFast:   "FAST",
		tui.AnimHue:    "HUE",
	}
	for mode, want := range labels {
		if got := mode.Label(); got != want {
			t.Errorf("mode %d label = %q, want %q", mode, got, want)
		}
	}
}

func TestAnimModePhaseDeltasMatchLoopDurations(t *testing.T) {
	cases := []struct {
		mode  tui.AnimMode
		ticks int
	}{
		{tui.AnimSlow, 240},   // 12s at a 50ms tick
		{tui.AnimNormal, 120}, // 6s
		{tui.AnimFast, 60},    // 3s
		{tui.AnimHue, 120},    // HUE sweeps at the normal rate
	}

	for _, tc := range cases {
		if got := tc.mode.LoopTicks(); got != tc.ticks {
			t.Errorf("%s LoopTicks = %d, want %d", tc.mode.Label(), got, tc.ticks)
		}
		want := 1.0 / float64(tc.ticks)
		if got := tc.mode.PhaseDelta(); math.Abs(got-want) > phaseTolerance {
			t.Errorf("%s PhaseDelta = %v, want %v", tc.mode.Label(), got, want)
		}
	}

	if got := tui.AnimOff.PhaseDelta(); got != 0 {
		t.Errorf("AnimOff PhaseDelta = %v, want 0", got)
	}
}

func TestAnimHueRotationRate(t *testing.T) {
	if got := tui.AnimHue.HueDegPerTick(); math.Abs(got-0.9) > phaseTolerance {
		t.Errorf("HUE should rotate 0.9 degrees per tick, got %v", got)
	}
	for _, mode := range []tui.AnimMode{tui.AnimOff, tui.AnimSlow, tui.AnimNormal, tui.AnimFast} {
		if got := mode.HueDegPerTick(); got != 0 {
			t.Errorf("%s should not rotate hue, got %v deg/tick", mode.Label(), got)
		}
	}
}

// TestAdvanceCompletesExactlyOneLoop is the core timing contract: driving a mode
// for its loop length must return the phase to the start.
func TestAdvanceCompletesExactlyOneLoop(t *testing.T) {
	modes := []tui.AnimMode{tui.AnimSlow, tui.AnimNormal, tui.AnimFast, tui.AnimHue}

	for _, mode := range modes {
		a := tui.NewAnimState(mode)
		for i := 0; i < mode.LoopTicks(); i++ {
			a.Advance()
		}
		if d := phaseDistanceToZero(a.Phase()); d > 1e-4 {
			t.Errorf("%s: after %d ticks phase = %v (off by %v), want a full loop",
				mode.Label(), mode.LoopTicks(), a.Phase(), d)
		}
		if a.Tick() != mode.LoopTicks() {
			t.Errorf("%s: tick counter = %d, want %d", mode.Label(), a.Tick(), mode.LoopTicks())
		}
	}
}

func TestAdvanceKeepsPhaseInRange(t *testing.T) {
	a := tui.NewAnimState(tui.AnimFast)
	for i := 0; i < 500; i++ {
		a.Advance()
		if p := a.Phase(); p < 0 || p >= 1 {
			t.Fatalf("phase %v out of range after %d ticks", p, i+1)
		}
	}
}

func TestAnimOffNeverAdvances(t *testing.T) {
	a := tui.NewAnimState(tui.AnimOff)
	if a.Enabled() {
		t.Errorf("AnimOff should not be enabled")
	}
	a.Advance()
	a.Advance()
	if a.Tick() != 0 || a.Phase() != 0 {
		t.Errorf("AnimOff advanced to tick=%d phase=%v", a.Tick(), a.Phase())
	}
}

// TestRegionPhasesLagTheBanner pins the agreed layering: badges trail by 10% and
// the footer by 20% of the loop.
func TestRegionPhasesLagTheBanner(t *testing.T) {
	a := tui.NewAnimState(tui.AnimNormal)
	for i := 0; i < 30; i++ {
		a.Advance()
	}

	wantBadge := tui.WrapPhase(a.Phase() - 0.10)
	wantFooter := tui.WrapPhase(a.Phase() - 0.20)

	if got := a.BadgePhase(); math.Abs(got-wantBadge) > phaseTolerance {
		t.Errorf("BadgePhase = %v, want %v", got, wantBadge)
	}
	if got := a.FooterPhase(); math.Abs(got-wantFooter) > phaseTolerance {
		t.Errorf("FooterPhase = %v, want %v", got, wantFooter)
	}
	if a.BadgePhase() == a.Phase() {
		t.Errorf("badges must be phase-offset from the banner, not in lockstep")
	}
}

func TestSetModeDoesNotJumpPhase(t *testing.T) {
	a := tui.NewAnimState(tui.AnimNormal)
	for i := 0; i < 45; i++ {
		a.Advance()
	}
	before := a.Phase()

	a.SetMode(tui.AnimFast)
	if a.Phase() != before {
		t.Errorf("switching modes jumped the phase from %v to %v", before, a.Phase())
	}
	if a.Mode() != tui.AnimFast {
		t.Errorf("SetMode did not take effect, mode = %s", a.Mode().Label())
	}
}

func TestCycleFromLaunchDefaultGoesToFast(t *testing.T) {
	a := tui.NewAnimState(tui.AnimNormal)
	if got := a.Cycle(); got != tui.AnimFast {
		t.Errorf("first /anim from the NORMAL launch default should give FAST, got %s", got.Label())
	}
	if got := a.Cycle(); got != tui.AnimHue {
		t.Errorf("second /anim should give HUE, got %s", got.Label())
	}
	if got := a.Cycle(); got != tui.AnimOff {
		t.Errorf("third /anim should switch the animation OFF, got %s", got.Label())
	}
	if a.Enabled() {
		t.Errorf("animation should be disabled after cycling to OFF")
	}
	if got := a.Cycle(); got != tui.AnimSlow {
		t.Errorf("cycling on from OFF should give SLOW, got %s", got.Label())
	}
}

func TestHueModeRotatesWhileSweeping(t *testing.T) {
	a := tui.NewAnimState(tui.AnimHue)

	startSweep := a.Phase()
	startHue := a.HueShift()

	for i := 0; i < 50; i++ {
		a.Advance()
	}

	if a.Phase() == startSweep {
		t.Errorf("HUE mode should still sweep, but the phase never moved")
	}
	if a.HueShift() == startHue {
		t.Errorf("HUE mode should rotate the palette hue, but it never moved")
	}
	if got := a.HueShift(); got < 0 || got >= 360 {
		t.Errorf("hue shift %v out of range", got)
	}

	// A full loop of ticks returns the hue to the start.
	b := tui.NewAnimState(tui.AnimHue)
	for i := 0; i < tui.AnimHueLoopTicks; i++ {
		b.Advance()
	}
	if d := phaseDistanceToZero(b.HueShift() / 360); d > 1e-4 {
		t.Errorf("after a full hue rotation hue = %v, want 0/360", b.HueShift())
	}
}

func TestNonHueModesNeverRotateHue(t *testing.T) {
	for _, mode := range []tui.AnimMode{tui.AnimSlow, tui.AnimNormal, tui.AnimFast} {
		a := tui.NewAnimState(mode)
		for i := 0; i < 40; i++ {
			a.Advance()
		}
		if got := a.HueShift(); got != 0 {
			t.Errorf("%s reported a hue shift of %v, want 0", mode.Label(), got)
		}
	}
}

func TestAnimationSupportedRespectsOptOuts(t *testing.T) {
	env := func(vars map[string]string) tui.EnvLookup {
		return func(key string) string { return vars[key] }
	}

	cases := []struct {
		name string
		vars map[string]string
		want bool
	}{
		{"plain terminal", map[string]string{"TERM": "xterm-256color"}, true},
		{"no color", map[string]string{"NO_COLOR": "1", "TERM": "xterm"}, false},
		{"no color empty", map[string]string{"NO_COLOR": "", "TERM": "xterm"}, true},
		{"no color whitespace", map[string]string{"NO_COLOR": "   "}, true},
		{"dumb terminal", map[string]string{"TERM": "dumb"}, false},
		{"dumb terminal uppercase", map[string]string{"TERM": "DUMB"}, false},
		{"missing term", map[string]string{}, true},
	}

	for _, tc := range cases {
		if got := tui.AnimationSupported(env(tc.vars)); got != tc.want {
			t.Errorf("%s: AnimationSupported = %v, want %v", tc.name, got, tc.want)
		}
	}

	if !tui.AnimationSupported(nil) {
		t.Errorf("a nil env lookup should default to supported")
	}
}

// TestSuccessiveFramesDiffer is the end-to-end proof that the banner actually
// animates: driving the clock one tick at a time must produce a different rendered
// frame each time. Colors are forced on because the default test profile strips
// every escape sequence, which would make all frames look identical.
func TestSuccessiveFramesDiffer(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)

	for _, mode := range []tui.AnimMode{tui.AnimSlow, tui.AnimNormal, tui.AnimFast, tui.AnimHue} {
		a := tui.NewAnimState(mode)
		frames := make([]string, 0, 4)
		for i := 0; i < 4; i++ {
			frames = append(frames, tui.RenderBannerWithRamp(80, a.Phase(), a.LUT()))
			a.Advance()
		}

		for i := 1; i < len(frames); i++ {
			if frames[i] == frames[i-1] {
				t.Errorf("%s: frame %d is identical to frame %d, so the banner is not moving",
					mode.Label(), i, i-1)
			}
		}

		for i := range frames {
			if strings.Count(frames[i], "38;2;") < 10 {
				t.Errorf("%s: frame %d is not swept per column", mode.Label(), i)
			}
		}
	}
}

func TestLUTIsSharedWithTheClock(t *testing.T) {
	a := tui.NewAnimState(tui.AnimNormal)
	if len(a.LUT()) != tui.RampLUTSize {
		t.Errorf("clock LUT has %d samples, want %d", len(a.LUT()), tui.RampLUTSize)
	}
	if a.LUT()[0] != tui.BuildRampLUT(0)[0] {
		t.Errorf("non-HUE clock should use the un-rotated ramp")
	}

	h := tui.NewAnimState(tui.AnimHue)
	for i := 0; i < 60; i++ {
		h.Advance()
	}
	if h.LUT()[0] == tui.BuildRampLUT(0)[0] {
		t.Errorf("HUE clock should rotate its ramp, but it matches the base ramp")
	}
}
