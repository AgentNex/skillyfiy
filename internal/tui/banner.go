package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

// BannerGradient defines the smooth 9-color gradient: Mint Green -> Cyan -> Silver Grey.
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
	bannerLettersRow0 = []string{"█▀▀", "█▀▄", "▀█▀", "█  ", "█  ", "█ █", "█▀▀", "▀█▀", "█ █"}
	bannerLettersRow1 = []string{"▄▄█", "█ █", "▄█▄", "█▄▄", "█▄▄", " █ ", "█  ", "▄█▄", " █ "}
)

// SubtitleText provides context for the tool.
const SubtitleText = "Terminal AI Agent Skill & MCP Context Overhead Optimizer [sky]"

// RenderBanner renders the bold, blocky, yet simple and slim header for "SKILLYFIY"
// with a mint green -> cyan -> silver grey gradient that dynamically adapts to screen size.
func RenderBanner(width int) string {
	w := width
	if w <= 0 {
		w = 80
	}

	if w < 37 {
		return CompactBanner(w)
	}

	// Normal and wide screens (>= 37 columns): render 2-line block letters
	var r0, r1 strings.Builder
	r0.WriteString(" ")
	r1.WriteString(" ")
	for i := 0; i < 9; i++ {
		style := lipgloss.NewStyle().Foreground(BannerGradient[i]).Bold(true)
		r0.WriteString(style.Render(bannerLettersRow0[i]))
		r1.WriteString(style.Render(bannerLettersRow1[i]))
		if i < 8 {
			r0.WriteString(" ")
			r1.WriteString(" ")
		}
	}

	line0 := r0.String()
	line1 := r1.String()

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

// CompactBanner returns a single-line compact representation for narrow or zoomed-in terminals.
func CompactBanner(width ...int) string {
	w := 80
	if len(width) > 0 && width[0] > 0 {
		w = width[0]
	}

	var b strings.Builder
	b.WriteString(" ⚡ ")
	rawChars := []rune("SKILLYFIY")
	for i, ch := range rawChars {
		style := lipgloss.NewStyle().Foreground(BannerGradient[i]).Bold(true)
		b.WriteString(style.Render(string(ch)))
	}
	b.WriteString(" ")

	if w >= 40 {
		b.WriteString(SubtitleStyle.Render("— SKILLYFIY Optimizer"))
	}

	line := b.String()
	if lipgloss.Width(line) >= w {
		line = truncate.StringWithTail(line, uint(w-1), "")
	}
	return line
}

// BannerHeight returns the number of vertical terminal lines consumed by the banner.
func BannerHeight(width ...int) int {
	w := 80
	if len(width) > 0 && width[0] > 0 {
		w = width[0]
	}
	if w >= 45 {
		return 3
	}
	if w >= 37 {
		return 2
	}
	return 1
}
