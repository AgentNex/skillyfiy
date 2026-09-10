package tui

import (
	"strings"
)

// ASCIIBanner contains the stylized compact block letters for "SKILLYFIY".
const ASCIIBanner = `  ___ _  _____ _    _   __  _____ ___ __   __
 / __| |/ /_ _| |  | |  \ \/ / __|_ _|\ \ / /
 \__ \ ' < | || |__| |__ \  /| _| | |  \ V / 
 |___/_|\_\___|____|____|/_/ |_| |___|  |_|  `

// SubtitleText provides context for the tool.
const SubtitleText = "Terminal AI Agent Skill & MCP Context Overhead Optimizer"

// RenderBanner renders the stylized header with Mint Green accent styling.
func RenderBanner(width int) string {
	banner := BannerStyle.Render(ASCIIBanner)
	subtitle := SubtitleStyle.Render(SubtitleText)

	return banner + "\n" + subtitle
}

// CompactBanner returns a single-line compact representation for very narrow terminals.
func CompactBanner() string {
	return BannerStyle.Render("⚡ SKILLYFIY") + " " + SubtitleStyle.Render("— Context Overhead Optimizer")
}

// BannerHeight returns the number of vertical terminal lines consumed by the banner.
func BannerHeight() int {
	return strings.Count(ASCIIBanner, "\n") + 2
}
