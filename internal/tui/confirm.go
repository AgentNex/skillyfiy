package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"skillyfiy/internal/model"
)

// RenderConfirmModal creates the centered Dry-Run Summary Modal overlay.
func RenderConfirmModal(items []model.AgentItem, width, height int, dryRun bool) string {
	var skillCount int
	var mcpCount int
	var totalTokens int
	var totalBytes int64

	for _, it := range items {
		if !it.Selected {
			continue
		}
		if it.Type == model.TypeSkill {
			skillCount++
		} else {
			mcpCount++
		}
		totalTokens += it.Tokens
		totalBytes += it.Size
	}

	title := "⚠️  IRREVERSIBLE PURGE CONFIRMATION"
	if dryRun {
		title = "🔍 DRY-RUN PURGE SIMULATION"
	}

	var b strings.Builder
	b.WriteString(ModalTitleStyle.Render(title) + "\n\n")

	b.WriteString(NormalTitleStyle.Render("Review resources scheduled for removal:") + "\n\n")

	b.WriteString(fmt.Sprintf("  • %-30s %s\n",
		InspectorValStyle.Render("Skills to permanently unlink:"),
		ColorMintGreenStyle(fmt.Sprintf("%d files", skillCount))))

	b.WriteString(fmt.Sprintf("  • %-30s %s\n",
		InspectorValStyle.Render("MCP servers to deregister:"),
		ColorMintBlueStyle(fmt.Sprintf("%d servers", mcpCount))))

	b.WriteString(fmt.Sprintf("  • %-30s %s\n",
		InspectorValStyle.Render("Estimated context reclaimed:"),
		FooterReclaimedStyle.Render(fmt.Sprintf("~%d tokens", totalTokens))))

	b.WriteString(fmt.Sprintf("  • %-30s %s\n\n",
		InspectorValStyle.Render("Filesystem size freed:"),
		InspectorValStyle.Render(model.FormatBytes(totalBytes))))

	if dryRun {
		b.WriteString(InspectorKeyStyle.Render("Note: Dry-run active. No files or config keys will be removed.") + "\n\n")
	} else {
		b.WriteString(NormalDescStyle.Render("Note: An automatic '.bak' backup will be saved before modifying MCP manifests.") + "\n\n")
	}

	b.WriteString(ModalTitleStyle.Render("Are you sure you want to purge these resources? (y/N)") + "\n\n")

	actions := fmt.Sprintf("  %s  %s  %s",
		ModalActionStyle.Render("[ y ] Confirm & Purge"),
		SubtitleStyle.Render("│"),
		NormalTitleStyle.Render("[ n / Esc ] Cancel & Return"),
	)
	b.WriteString(actions)

	modalBox := ModalBoxStyle.Render(b.String())

	// Center modal on screen
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modalBox)
}

func ColorMintGreenStyle(s string) string {
	return lipgloss.NewStyle().Foreground(ColorMintGreen).Bold(true).Render(s)
}

func ColorMintBlueStyle(s string) string {
	return lipgloss.NewStyle().Foreground(ColorMintBlue).Bold(true).Render(s)
}
