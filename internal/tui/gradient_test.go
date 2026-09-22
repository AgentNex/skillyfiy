package tui_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"skillyfiy/internal/tui"
)

var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

// hexDelta returns the largest single-channel difference between two #rrggbb colors.
func hexDelta(t *testing.T, a, b lipgloss.Color) int {
	t.Helper()
	ar, ag, ab, ok1 := hexChannels(string(a))
	br, bg, bb, ok2 := hexChannels(string(b))
	if !ok1 || !ok2 {
		t.Fatalf("unparsable colors: %q %q", string(a), string(b))
	}
	delta := 0
	for _, d := range []int{ar - br, ag - bg, ab - bb} {
		if d < 0 {
			d = -d
		}
		if d > delta {
			delta = d
		}
	}
	return delta
}

func hexChannels(s string) (int, int, int, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	var out [3]int
	for i := 0; i < 3; i++ {
		hi := hexVal(s[i*2])
		lo := hexVal(s[i*2+1])
		if hi < 0 || lo < 0 {
			return 0, 0, 0, false
		}
		out[i] = hi*16 + lo
	}
	return out[0], out[1], out[2], true
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func TestRampLUTSizeAndDeterminism(t *testing.T) {
	a := tui.BuildRampLUT(0)
	b := tui.BuildRampLUT(0)

	if len(a) != tui.RampLUTSize {
		t.Fatalf("expected %d ramp samples, got %d", tui.RampLUTSize, len(a))
	}
	if len(a) != len(b) {
		t.Fatalf("ramp length is not stable: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("ramp is not deterministic at index %d: %q vs %q", i, a[i], b[i])
		}
		if strings.TrimSpace(string(a[i])) == "" {
			t.Fatalf("ramp index %d is an empty color", i)
		}
	}
}

func TestRampEndpointsMatchPaletteAnchors(t *testing.T) {
	lut := tui.BuildRampLUT(0)

	first := strings.ToLower(string(lut[0]))
	last := strings.ToLower(string(lut[len(lut)-1]))

	if first != "#10b981" {
		t.Errorf("ramp should start at the mint anchor #10b981, got %q", first)
	}
	if last != "#f1f5f9" {
		t.Errorf("ramp should end at the silver anchor #f1f5f9, got %q", last)
	}
}

func TestRampIsContinuousNotBanded(t *testing.T) {
	lut := tui.BuildRampLUT(0)

	// A stepped implementation would emit only 9 distinct colors.
	distinct := make(map[lipgloss.Color]struct{}, len(lut))
	for _, c := range lut {
		distinct[c] = struct{}{}
	}
	if len(distinct) < 200 {
		t.Errorf("ramp looks banded: only %d distinct colors out of %d", len(distinct), len(lut))
	}

	// Neighbouring samples must be close, which is what makes the sweep smooth
	// instead of jumping between palette anchors.
	for i := 1; i < len(lut); i++ {
		if d := hexDelta(t, lut[i-1], lut[i]); d > 12 {
			t.Errorf("ramp jumps by %d on channel at index %d (%q -> %q)", d, i, lut[i-1], lut[i])
		}
	}
}

func TestBuildRampLUTAppliesHueShift(t *testing.T) {
	base := tui.BuildRampLUT(0)
	shifted := tui.BuildRampLUT(90)

	if len(shifted) != len(base) {
		t.Fatalf("hue shift changed the ramp length: %d vs %d", len(shifted), len(base))
	}

	// The mint anchor sits at hue 160; rotating it by 90 degrees lands on this
	// blue-violet, which is the defining behaviour of HUE mode.
	if got := strings.ToLower(string(shifted[0])); got != "#2d10b9" {
		t.Errorf("expected hue-shifted mint to be #2d10b9, got %q", got)
	}
	if shifted[0] == base[0] {
		t.Errorf("hue shift left the first ramp sample unchanged")
	}

	// Hue shifting must stay deterministic.
	again := tui.BuildRampLUT(90)
	for i := range again {
		if again[i] != shifted[i] {
			t.Fatalf("hue-shifted ramp is not deterministic at index %d", i)
		}
	}
}

func TestSampleRampWrapsAndInterpolates(t *testing.T) {
	lut := tui.BuildRampLUT(0)

	if got, want := tui.SampleRamp(lut, 0), lut[0]; got != want {
		t.Errorf("SampleRamp(0) = %q, want %q", got, want)
	}
	if got, want := tui.SampleRamp(lut, 1), lut[0]; got != want {
		t.Errorf("SampleRamp(1) should wrap to the first sample, got %q want %q", got, want)
	}
	if got, want := tui.SampleRamp(lut, -1), lut[0]; got != want {
		t.Errorf("SampleRamp(-1) should wrap to the first sample, got %q want %q", got, want)
	}
	if got, want := tui.SampleRamp(lut, 1.5), lut[tui.RampLUTSize/2]; got != want {
		t.Errorf("SampleRamp(1.5) should wrap to the midpoint, got %q want %q", got, want)
	}

	// Two nearby phases must land on neighbouring (not identical) samples.
	near1 := tui.SampleRamp(lut, 0.4)
	near2 := tui.SampleRamp(lut, 0.4+1.0/float64(tui.RampLUTSize-1))
	if near1 == near2 {
		t.Errorf("adjacent phases produced identical colors")
	}
}

func TestSampleRampHandlesEmptyRamp(t *testing.T) {
	got := tui.SampleRamp(nil, 0.5)
	if string(got) == "" {
		t.Errorf("SampleRamp on an empty ramp should fall back to a color, got empty")
	}
	_ = tui.SampleRamp([]lipgloss.Color{}, -3)
}

func TestWrapPhaseAlwaysNormalizes(t *testing.T) {
	cases := map[float64]float64{
		0:     0,
		0.25:  0.25,
		1:     0,
		-0.25: 0.75,
		7.5:   0.5,
		-3.25: 0.75,
	}
	for in, want := range cases {
		if got := tui.WrapPhase(in); got != want {
			t.Errorf("WrapPhase(%v) = %v, want %v", in, got, want)
		}
	}
}

// TestSweepPreservesBannerLayout proves the animation changes only color, never
// geometry: same line count and same visible text at any phase.
func TestSweepPreservesBannerLayout(t *testing.T) {
	for _, width := range []int{120, 80, 60, 45, 38, 36, 28} {
		resting := tui.RenderBannerWithRamp(width, 0, nil)
		moving := tui.RenderBannerWithRamp(width, 0.37, nil)

		restingLines := strings.Split(resting, "\n")
		movingLines := strings.Split(moving, "\n")

		if len(restingLines) != len(movingLines) {
			t.Fatalf("width %d: phase changed line count %d -> %d", width, len(restingLines), len(movingLines))
		}
		if len(restingLines) != tui.BannerHeight(width) {
			t.Errorf("width %d: expected %d banner lines, got %d", width, tui.BannerHeight(width), len(restingLines))
		}

		for i := range restingLines {
			if stripANSI(restingLines[i]) != stripANSI(movingLines[i]) {
				t.Errorf("width %d line %d: sweep altered visible text", width, i)
			}
			if lipgloss.Width(movingLines[i]) != lipgloss.Width(restingLines[i]) {
				t.Errorf("width %d line %d: sweep altered display width", width, i)
			}
		}
	}
}

// TestSweepChangesRenderedColor drives the renderer at truecolor so the color
// payload is visible; the default test profile strips all escape sequences.
func TestSweepChangesRenderedColor(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)

	first := tui.RenderBannerWithRamp(80, 0, nil)
	second := tui.RenderBannerWithRamp(80, 0.35, nil)

	if !strings.Contains(first, "\x1b[") {
		t.Fatalf("expected ANSI color sequences at truecolor, got %q", first)
	}
	if first == second {
		t.Errorf("advancing the phase did not change the rendered banner")
	}
	if stripANSI(first) != stripANSI(second) {
		t.Errorf("phase change must not alter banner text")
	}

	// The sweep colors per column, so the row must carry many truecolor
	// foreground sequences rather than one per 3-cell letter.
	// lipgloss packs bold and color into a single SGR, so the marker to count is
	// the "38;2;" truecolor foreground selector, not the "\x1b[38" prefix.
	if got := strings.Count(first, "38;2;"); got < 10 {
		t.Errorf("per-column sweep produced only %d foreground sequences", got)
	}
}

func TestCompactBannerSweepsToo(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)

	a := tui.CompactBannerWithRamp(30, 0, nil)
	b := tui.CompactBannerWithRamp(30, 0.6, nil)

	if !strings.Contains(stripANSI(a), "SKILLYFIY") {
		t.Errorf("compact banner lost its wordmark: %q", stripANSI(a))
	}
	if a == b {
		t.Errorf("compact banner did not respond to the phase")
	}
}
