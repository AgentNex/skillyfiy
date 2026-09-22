package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

// BannerGradient defines the smooth 9-color gradient: Mint Green -> Cyan -> Silver Grey.
// These anchors are interpolated into a continuous ramp by gradient.go; the banner
// samples that ramp once per display column so the sweep flows rather than stepping
// letter by letter.
var BannerGradient = []lipgloss.TerminalColor{
	lipgloss.Color("#10B981"), // S: Vibrant Mint Green
	lipgloss.Color("#14B8A6"), // K: Mint Teal
	lipgloss.Color("#06B6D4"), // I: Teal Cyan
	lipgloss.Color("#0EA5E9"), // L: Vivid Cyan
	lipgloss.Color("#38BDF8"), // L: Sky Cyan
	lipgloss.Color("#7DD3FC"), // Y: Light Sky Cyan
	lipgloss.Color("#94A3B8"), // F: Cool Slate Silver
	lipgloss.Color("#CBD5E1"), // I: Silver Grey
	lipgloss.Color("#F1F5F9"), // Y: Bright Crisp Silver
}

// Block letter rows for "SKILLYFIY" (3 cells per letter, 35 columns total).
var (
	bannerLettersRow0 = []string{"█▀▀", "█▄▀", "▀█▀", "█  ", "█  ", "█ █", "█▀▀", "▀█▀", "█ █"}
	bannerLettersRow1 = []string{"▄▄█", "█▀▄", "▄█▄", "█▄▄", "█▄▄", " █ ", "█▀ ", "▄█▄", " █ "}
)

// SubtitleText provides context for the tool.
const SubtitleText = "Terminal AI Agent Skill & MCP Context Overhead Optimizer [sky]"

// RenderBanner renders the banner with the gradient at its resting phase.
func RenderBanner(width int) string {
	return RenderBannerAnimated(width, 0)
}

// RenderBannerAnimated renders the banner with the gradient swept to phase,
// using the un-rotated base ramp.
func RenderBannerAnimated(width int, phase float64) string {
	return RenderBannerWithRamp(width, phase, nil)
}

// RenderBannerWithRamp renders the banner by sampling lut once per column.
// A nil lut falls back to the base ramp. The header stays bold, blocky and slim,
// and adapts to narrow screens exactly as it did before it animated.
func RenderBannerWithRamp(width int, phase float64, lut []lipgloss.Color) string {
	w := width
	if w <= 0 {
		w = 80
	}
	if lut == nil {
		lut = BaseRampLUT()
	}

	if w < 37 {
		return CompactBannerWithRamp(w, phase, lut)
	}

	// Normal and wide screens (>= 37 columns): render 2-line block letters.
	// Period is the row's own display width, so one full palette traversal spans
	// the banner regardless of terminal size.
	raw0 := " " + strings.Join(bannerLettersRow0, " ")
	raw1 := " " + strings.Join(bannerLettersRow1, " ")

	line0 := renderSweepLine(raw0, phase, lut)
	line1 := renderSweepLine(raw1, phase, lut)

	if lipgloss.Width(line0) >= w {
		line0 = truncate.StringWithTail(line0, uint(w-1), "")
	}
	if lipgloss.Width(line1) >= w {
		line1 = truncate.StringWithTail(line1, uint(w-1), "")
	}

	if w >= 45 {
		sub := " — SKILLYFIY Context Optimizer [sky]"
		if w < 60 {
			sub = " — SKILLYFIY Optimizer"
		}
		avail := w - 2
		if avail > 0 && len(sub) > avail {
			sub = truncate.StringWithTail(sub, uint(avail), "")
		}
		line2 := SubtitleStyle.Render(sub)
		return line0 + "\n" + line1 + "\n" + line2
	}

	return line0 + "\n" + line1
}

// CompactBanner returns a single-line compact representation for narrow or zoomed-in
// terminals, with the gradient at its resting phase.
func CompactBanner(width ...int) string {
	return CompactBannerWithRamp(firstWidth(width), 0, nil)
}

// CompactBannerWithRamp is CompactBanner with an explicit phase and ramp.
func CompactBannerWithRamp(width int, phase float64, lut []lipgloss.Color) string {
	w := width
	if w <= 0 {
		w = 80
	}
	if lut == nil {
		lut = BaseRampLUT()
	}

	line := renderSweepLine(" ⚡ SKILLYFIY ", phase, lut) + " "
	if w >= 40 {
		line += SubtitleStyle.Render("— SKILLYFIY Optimizer")
	}

	if lipgloss.Width(line) >= w {
		line = truncate.StringWithTail(line, uint(w-1), "")
	}
	return line
}

// BannerHeight returns the number of vertical terminal lines consumed by the banner.
func BannerHeight(width ...int) int {
	w := firstWidth(width)
	if w >= 45 {
		return 3
	}
	if w >= 37 {
		return 2
	}
	return 1
}

// firstWidth normalizes the optional width argument used by the compact helpers.
func firstWidth(width []int) int {
	if len(width) > 0 && width[0] > 0 {
		return width[0]
	}
	return 80
}
