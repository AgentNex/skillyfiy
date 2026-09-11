package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderHelpModal builds a centered, responsive modal listing all keybindings and commands.
func RenderHelpModal(width, height int) string {
	var b strings.Builder

	title := "⚡ SKILLYFIY KEYBINDINGS & CONTROLS"
	b.WriteString(BannerStyle.Render(title) + "\n")

	modalW := width - 4
	if modalW > 68 {
		modalW = 68
	}
	if modalW < 36 {
		modalW = 36
	}

	keyWidth := 15
	if modalW < 50 {
		keyWidth = 9
	}
	keyStyle := HelpKeyStyle.Copy().Width(keyWidth)

	row := func(k, desc string) string {
		return fmt.Sprintf("  %s %s\n", keyStyle.Render(k), HelpDescStyle.Render(desc))
	}

	isCompact := height < 24
	if !isCompact {
		b.WriteString("\n" + HelpSectionStyle.Render("▶ NAVIGATION") + "\n")
	}
	b.WriteString(row("↑ / ↓ / k / j", "Move cursor / scroll inspector"))
	b.WriteString(row("Tab", "Switch focus between List & Inspector"))

	if !isCompact {
		b.WriteString(HelpSectionStyle.Render("▶ SELECTION & SPAN") + "\n")
	}
	b.WriteString(row("Space / x", "Toggle selection of single item"))
	b.WriteString(row("v", "Turn Multi-Select mode ON / OFF"))
	b.WriteString(row("s", "Pin start item; scroll; press 's' to lock span"))
	b.WriteString(row("Ctrl+A", "Select or deselect all visible items"))

	if !isCompact {
		b.WriteString(HelpSectionStyle.Render("▶ SEARCH & FILTER") + "\n")
	}
	b.WriteString(row("/", "Focus search bar (press '/' again to exit)"))
	b.WriteString(row("Ctrl+S / F1", "Cycle 6 sort modes (Newest, Size, A-Z...)"))
	b.WriteString(row("Ctrl+T / F2", "Cycle 3 resource filters (All, Skills, MCP)"))

	if !isCompact {
		b.WriteString(HelpSectionStyle.Render("▶ ACTIONS") + "\n")
	}
	b.WriteString(row("Enter", "Review and confirm purge of selected items"))
	b.WriteString(row("i / Esc", "Close this help overlay"))
	b.WriteString(row("Ctrl+Q / Ctrl+C", "Exit Skillyfiy immediately"))

	b.WriteString("\n" + SubtitleStyle.Render("Press [ i ] or [ Esc ] to return to dashboard"))

	modalBox := HelpModalStyle.Copy().Width(modalW).Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modalBox)
}
