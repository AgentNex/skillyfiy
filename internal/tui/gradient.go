package tui

import (
	"math"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// RampLUTSize is the number of samples in the interpolated gradient lookup table.
// A 256-entry table keeps per-frame work to a single array lookup per column.
const RampLUTSize = 256

// BadgePairOffset separates the SKILL and MCP badge colors on the ramp so the two
// resource types stay visually distinguishable while both sweep with the gradient.
const BadgePairOffset = 0.5

// hslTriple is an HSL color whose hue is intentionally allowed to run outside
// [0,360). Anchors are "unwrapped" so that neighbouring anchors interpolate along
// their shortest path instead of blending the long way around the color wheel.
//
// HSL is used rather than HCL/Lab on purpose: HSL to RGB is always in the sRGB
// gamut, so a sampled ramp can never be clipped. Interpolating in HCL walks
// through out-of-gamut chroma values and clipping then produces \~250-point
// channel jumps between neighbouring samples, which reads as hard banding.
type hslTriple struct {
	H, S, L float64
}

var (
	rampOnce       sync.Once
	rampAnchors    []hslTriple
	baseRampLUT    []lipgloss.Color
	rampTextCache  sync.Map // lipgloss.Color -> lipgloss.Style
	rampBadgeCache sync.Map // lipgloss.Color -> lipgloss.Style
)

// NormalizeHue folds any hue into [0,360).
func NormalizeHue(h float64) float64 {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return h
}

// WrapPhase folds any phase value into [0,1).
func WrapPhase(t float64) float64 {
	t = math.Mod(t, 1)
	if t < 0 {
		t += 1
	}
	return t
}

// terminalColorHex extracts the hex payload from a lipgloss terminal color.
func terminalColorHex(tc lipgloss.TerminalColor) (string, bool) {
	if c, ok := tc.(lipgloss.Color); ok {
		return string(c), true
	}
	return "", false
}

// buildAnchors parses BannerGradient into unwrapped HSL anchors.
func buildAnchors() []hslTriple {
	colors := make([]colorful.Color, 0, len(BannerGradient))
	for _, tc := range BannerGradient {
		hex, ok := terminalColorHex(tc)
		if !ok {
			continue
		}
		c, err := colorful.Hex(hex)
		if err != nil {
			continue
		}
		colors = append(colors, c)
	}

	if len(colors) == 0 {
		// Defensive fallback: a plausible mint green so rendering never crashes.
		return []hslTriple{{H: 160, S: 0.84, L: 0.39}}
	}

	out := make([]hslTriple, len(colors))
	prevH := 0.0
	for i, c := range colors {
		h, s, l := c.Hsl()
		if i > 0 {
			// Pick the hue representative nearest the previous anchor so the
			// interpolation never takes the long way around the wheel.
			for h < prevH-180 {
				h += 360
			}
			for h > prevH+180 {
				h -= 360
			}
		}
		out[i] = hslTriple{H: h, S: s, L: l}
		prevH = h
	}
	return out
}

// loadRamp builds the shared anchor set and the un-rotated base ramp exactly once.
func loadRamp() {
	rampOnce.Do(func() {
		rampAnchors = buildAnchors()
		baseRampLUT = buildRampLUTFrom(rampAnchors, 0)
	})
}

// BaseRampLUT returns the cached, un-rotated gradient ramp. It is safe for
// concurrent use and must not be mutated by callers.
func BaseRampLUT() []lipgloss.Color {
	loadRamp()
	return baseRampLUT
}

// BuildRampLUT interpolates the banner palette into RampLUTSize samples, rotating
// every color's hue by hueShiftDeg. It is a pure function of hueShiftDeg: the same
// input always yields an identical ramp. A zero shift returns the cached base ramp.
func BuildRampLUT(hueShiftDeg float64) []lipgloss.Color {
	loadRamp()
	if hueShiftDeg == 0 {
		return baseRampLUT
	}
	return buildRampLUTFrom(rampAnchors, hueShiftDeg)
}

// buildRampLUTFrom expands unwrapped HSL anchors into a lookup table.
func buildRampLUTFrom(anchors []hslTriple, hueShiftDeg float64) []lipgloss.Color {
	out := make([]lipgloss.Color, RampLUTSize)

	if len(anchors) == 0 {
		return out
	}
	if len(anchors) == 1 {
		a := anchors[0]
		c := lipgloss.Color(colorful.Hsl(NormalizeHue(a.H+hueShiftDeg), a.S, a.L).Hex())
		for i := range out {
			out[i] = c
		}
		return out
	}

	last := float64(len(anchors) - 1)
	for i := 0; i < RampLUTSize; i++ {
		// u walks the anchor list continuously: 0 -> first anchor, last -> final anchor.
		u := float64(i) / float64(RampLUTSize-1) * last
		seg := int(u)
		if seg > len(anchors)-2 {
			seg = len(anchors) - 2
		}
		if seg < 0 {
			seg = 0
		}

		a, b := anchors[seg], anchors[seg+1]
		frac := u - float64(seg)

		h := a.H + (b.H-a.H)*frac + hueShiftDeg
		s := a.S + (b.S-a.S)*frac
		l := a.L + (b.L-a.L)*frac
		out[i] = lipgloss.Color(colorful.Hsl(NormalizeHue(h), s, l).Hex())
	}
	return out
}

// SampleRamp samples a ramp at t, wrapping t into [0,1) first. Adjacent inputs
// produce neighbouring ramp entries, which is what makes the sweep continuous.
func SampleRamp(lut []lipgloss.Color, t float64) lipgloss.Color {
	if len(lut) == 0 {
		return lipgloss.Color("#10B981")
	}
	idx := int(math.Round(WrapPhase(t) * float64(len(lut)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(lut) {
		idx = len(lut) - 1
	}
	return lut[idx]
}

// RampTextStyle returns a cached bold foreground style for a ramp color.
func RampTextStyle(c lipgloss.Color) lipgloss.Style {
	if v, ok := rampTextCache.Load(c); ok {
		return v.(lipgloss.Style)
	}
	s := lipgloss.NewStyle().Foreground(c).Bold(true)
	rampTextCache.Store(c, s)
	return s
}

// RampBadgeStyle returns a cached pill style for a ramp-colored badge background.
// The foreground stays dark because every ramp sample is a light-to-mid tone,
// which keeps the pill legible across the whole sweep.
func RampBadgeStyle(c lipgloss.Color) lipgloss.Style {
	if v, ok := rampBadgeCache.Load(c); ok {
		return v.(lipgloss.Style)
	}
	s := lipgloss.NewStyle().Background(c).Foreground(ColorBase).Bold(true).Padding(0, 1)
	rampBadgeCache.Store(c, s)
	return s
}

// renderSweepLine colors one raw banner row by sampling the ramp once per column.
//
// The ramp period is the row's own display width, so exactly one full palette
// traversal spans the banner and the look is identical at every terminal width.
// Whitespace is emitted unstyled to keep the escape-sequence count down.
func renderSweepLine(raw string, phase float64, lut []lipgloss.Color) string {
	if raw == "" {
		return raw
	}
	if len(lut) == 0 {
		lut = BaseRampLUT()
	}

	period := float64(lipgloss.Width(raw))
	if period <= 0 {
		return raw
	}

	var b strings.Builder
	b.Grow(len(raw) * 4)

	col := 0
	for _, r := range raw {
		width := lipgloss.Width(string(r))
		if r == ' ' || r == '\t' {
			b.WriteRune(r)
			col += width
			continue
		}
		t := phase + float64(col)/period
		b.WriteString(RampTextStyle(SampleRamp(lut, t)).Render(string(r)))
		col += width
	}
	return b.String()
}
