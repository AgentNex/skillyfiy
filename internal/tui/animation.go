package tui

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Animation timing constants. The tick interval is fixed for every mode; modes
// differ only in how much phase they advance per tick, which keeps the redraw
// rate (and therefore CPU/battery cost) predictable on mobile terminals.
const (
	AnimTickInterval = 50 * time.Millisecond // 20 FPS

	// Ticks required for one full palette loop in each mode. At a 50ms tick:
	// slow = 12s, normal = 6s, fast = 3s.
	AnimSlowLoopTicks   = 240
	AnimNormalLoopTicks = 120
	AnimFastLoopTicks   = 60

	// A full hue rotation takes 20 seconds (400 ticks) => 0.9 degrees per tick.
	AnimHueLoopTicks = 400

	// Phase lags, expressed as fractions of a loop. Badges trail the banner by
	// 10% (~0.6s at normal speed) and the footer trails by a further 10%.
	BadgePhaseLag  = 0.10
	FooterPhaseLag = 0.20

	// ToastDuration is how long a footer notice stays on screen.
	ToastDuration = 2 * time.Second
)

// AnimMode is the animation mode selected by the /anim command.
type AnimMode int

const (
	AnimOff AnimMode = iota
	AnimSlow
	AnimNormal
	AnimFast
	AnimHue
)

// Label returns the uppercase label shown in the footer toast.
func (m AnimMode) Label() string {
	switch m {
	case AnimSlow:
		return "SLOW"
	case AnimNormal:
		return "NORMAL"
	case AnimFast:
		return "FAST"
	case AnimHue:
		return "HUE"
	default:
		return "OFF"
	}
}

// Next returns the next mode in the cycle, which is exactly:
// OFF -> SLOW -> NORMAL -> FAST -> HUE -> OFF
func (m AnimMode) Next() AnimMode {
	return (m + 1) % 5
}

// PhaseDelta is the fraction of a full loop advanced per tick.
func (m AnimMode) PhaseDelta() float64 {
	switch m {
	case AnimSlow:
		return 1.0 / float64(AnimSlowLoopTicks)
	case AnimNormal:
		return 1.0 / float64(AnimNormalLoopTicks)
	case AnimFast:
		return 1.0 / float64(AnimFastLoopTicks)
	case AnimHue:
		return 1.0 / float64(AnimNormalLoopTicks) // HUE sweeps at the normal rate
	default:
		return 0
	}
}

// HueDegPerTick is how far the palette hue rotates each tick.
func (m AnimMode) HueDegPerTick() float64 {
	if m != AnimHue {
		return 0
	}
	return 360.0 / float64(AnimHueLoopTicks)
}

// LoopTicks is the number of ticks in one full sweep for this mode (0 when off).
func (m AnimMode) LoopTicks() int {
	switch m {
	case AnimSlow:
		return AnimSlowLoopTicks
	case AnimNormal:
		return AnimNormalLoopTicks
	case AnimFast:
		return AnimFastLoopTicks
	case AnimHue:
		return AnimNormalLoopTicks
	default:
		return 0
	}
}

// animTickMsg drives the animation clock.
type animTickMsg struct {
	tick int
}

func animTickCmd(tick int) tea.Cmd {
	return tea.Tick(AnimTickInterval, func(time.Time) tea.Msg {
		return animTickMsg{tick: tick}
	})
}

// toastMsg expires a footer notice.
type toastMsg struct {
	seq int
}

// animState is the animation clock. AppModel owns exactly one, shared by pointer
// so that copies of the model (Bubble Tea uses value receivers) all observe the
// same clock and the ramp cache stays warm.
//
// Phase is accumulated per tick rather than derived from wall-clock time. That is
// what lets a pause freeze the sweep and resume from the identical position, and
// it means switching modes never makes the gradient jump.
type animState struct {
	mode         AnimMode
	phase        float64
	tick         int
	tickInFlight bool
	lut          []lipgloss.Color
	lutHue       float64
}

// NewAnimState creates an animation clock in the given mode.
func NewAnimState(mode AnimMode) *animState {
	a := &animState{mode: mode}
	a.RefreshLUT()
	return a
}

// Enabled reports whether the clock should be ticking at all.
func (a *animState) Enabled() bool {
	return a != nil && a.mode != AnimOff
}

// Mode returns the current mode.
func (a *animState) Mode() AnimMode {
	if a == nil {
		return AnimOff
	}
	return a.mode
}

// Tick returns how many ticks have elapsed.
func (a *animState) Tick() int {
	if a == nil {
		return 0
	}
	return a.tick
}

// Phase returns the current sweep phase in [0,1).
func (a *animState) Phase() float64 {
	if a == nil {
		return 0
	}
	return a.phase
}

// SetMode switches modes, refreshing the ramp for the new hue shift.
func (a *animState) SetMode(mode AnimMode) {
	if a == nil {
		return
	}
	a.mode = mode
	a.RefreshLUT()
}

// Cycle advances to the next mode and returns it. This is what bare /anim does.
func (a *animState) Cycle() AnimMode {
	if a == nil {
		return AnimOff
	}
	a.SetMode(a.mode.Next())
	return a.mode
}

// Advance moves the clock forward by exactly one tick. It is only ever called
// from the tick handler, so pausing simply stops calling it.
func (a *animState) Advance() {
	if a == nil || !a.Enabled() {
		return
	}
	a.tick++
	a.phase = WrapPhase(a.phase + a.mode.PhaseDelta())
	a.RefreshLUT()
}

// HueShift is the current palette hue rotation in degrees.
func (a *animState) HueShift() float64 {
	if a == nil || a.mode != AnimHue {
		return 0
	}
	return math.Mod(float64(a.tick)*a.mode.HueDegPerTick(), 360)
}

// BadgePhase is the phase used by the item badge pills.
func (a *animState) BadgePhase() float64 {
	if a == nil {
		return 0
	}
	return WrapPhase(a.phase - BadgePhaseLag)
}

// FooterPhase is the phase used by the footer accent.
func (a *animState) FooterPhase() float64 {
	if a == nil {
		return 0
	}
	return WrapPhase(a.phase - FooterPhaseLag)
}

// RefreshLUT rebuilds the ramp only when the hue shift actually changed, so
// non-HUE modes build the ramp once and then simply sample it every frame.
func (a *animState) RefreshLUT() {
	if a == nil {
		return
	}
	hue := a.HueShift()
	if a.lut != nil && a.lutHue == hue {
		return
	}
	a.lut = BuildRampLUT(hue)
	a.lutHue = hue
}

// LUT returns the current ramp. Non-mutating: the ramp is refreshed in Advance
// and SetMode so that rendering never has to build one.
func (a *animState) LUT() []lipgloss.Color {
	if a == nil || a.lut == nil {
		return BaseRampLUT()
	}
	return a.lut
}

// EnvLookup reads an environment variable; injectable so support detection is
// testable without mutating the real process environment.
type EnvLookup func(string) string

// AnimationSupported reports whether the terminal can render an animated
// gradient. NO_COLOR and TERM=dumb opt out entirely; every other terminal is
// allowed through and lipgloss auto-degrades the palette as needed.
func AnimationSupported(env EnvLookup) bool {
	if env == nil {
		return true
	}
	if strings.TrimSpace(env("NO_COLOR")) != "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(env("TERM")), "dumb") {
		return false
	}
	return true
}

// ParseCommand splits a command-mode input such as "/anim" into its name and
// remaining argument. The leading slash is optional, matching is case-insensitive.
func ParseCommand(input string) (name string, arg string, ok bool) {
	s := strings.TrimSpace(input)
	s = strings.TrimSpace(strings.TrimPrefix(s, "/"))
	if s == "" {
		return "", "", false
	}
	fields := strings.Fields(s)
	name = strings.ToLower(fields[0])
	if len(fields) > 1 {
		arg = strings.Join(fields[1:], " ")
	}
	return name, arg, true
}
