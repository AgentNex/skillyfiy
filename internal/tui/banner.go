package tui

import (
	"strings"

	"github.com/muesli/reflow/truncate"
)

// ASCIIBanner contains the stylized compact block letters for "SKILLYFIY".
const ASCIIBanner = `  ___ _  _____ _    _   __  _____ ___ __   __
 / __| |/ /_ _| |  | |  \ \/ / __|_ _|\ \ / /
 \__ \ ' < | || |__| |__ \  /| _| | |  \ V / 
 |___/_|\_\___|____|____|/_/ |_| |___|  |_|  `

// SubtitleText provides context for the tool.
const SubtitleText = "Terminal AI Agent Skill & MCP Context Overhead Optimizer"

// RenderBanner renders the stylized header, adapting dynamically to terminal width.
func RenderBanner(width int) string {
	if width >= 70 {
		banner := BannerStyle.Render(ASCIIBanner)
		sub := SubtitleText
		avail := width - 2
		if avail > 0 && len(sub) > avail {
			sub = truncate.StringWithTail(sub, uint(avail), "…")
		}
		subtitle := SubtitleStyle.Render(sub)
		return banner + "\n" + subtitle
	}

	return CompactBanner(width)
}

// CompactBanner returns a single-line compact representation for narrow terminals.
func CompactBanner(width ...int) string {
	w := 80
	if len(width) > 0 && width[0] > 0 {
		w = width[0]
	}

	title := "⚡ SKILLYFIY"
	desc := "Context Overhead Optimizer"
	if w < 50 {
		desc = "Optimizer"
	}
	if w < 28 {
		return BannerStyle.Render(title)
	}

	line := BannerStyle.Render(title) + " " + SubtitleStyle.Render("— "+desc)
	if w > 2 && len(title)+len(desc)+5 > w {
		line = truncate.StringWithTail(line, uint(w-2), "…")
	}
	return line
}

// BannerHeight returns the number of vertical terminal lines consumed by the banner.
func BannerHeight(width ...int) int {
	w := 80
	if len(width) > 0 && width[0] > 0 {
		w = width[0]
	}
	if w >= 70 {
		return strings.Count(ASCIIBanner, "\n") + 2
	}
	return 1
}
