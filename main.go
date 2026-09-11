package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"skillyfiy/internal/config"
	"skillyfiy/internal/engine"
	"skillyfiy/internal/model"
	"skillyfiy/internal/tui"
)

// Version is injected at build time via -ldflags="-X main.Version=v0.2.4"
var Version = "v0.2.4"

func main() {
	cfg, isVersion := config.ParseFlags()
	if isVersion {
		fmt.Printf("sky (skillyfiy) %s (%s/%s, pure Go, zero-cgo)\n", Version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	cfg.Version = Version

	// 1. Discover Skills
	var allItems []model.AgentItem
	skillItems, err := engine.ScanSkills(cfg.SkillsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning scanning skills directory (%s): %v\n", cfg.SkillsPath, err)
	} else {
		allItems = append(allItems, skillItems...)
	}

	// 2. Discover MCP Servers
	mcpItems, err := engine.ParseMCP(cfg.MCPPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning parsing MCP manifest (%s): %v\n", cfg.MCPPath, err)
	} else {
		allItems = append(allItems, mcpItems...)
	}

	if len(allItems) == 0 {
		fmt.Println(tui.CompactBanner())
		fmt.Printf("\nNo agent skills or MCP servers were discovered in the configured paths:\n")
		fmt.Printf("  • Skills Directory : %s\n", cfg.SkillsPath)
		fmt.Printf("  • MCP Settings     : %s\n\n", cfg.MCPPath)
		fmt.Printf("You can point to custom locations using flags:\n")
		fmt.Printf("  sky --skills /path/to/skills --mcp /path/to/settings.json\n\n")
		return
	}

	// 3. Initialize Bubble Tea Program in AltScreen mode
	app := tui.NewAppModel(cfg, allItems)
	program := tea.NewProgram(app, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running Skillyfiy TUI: %v\n", err)
		os.Exit(1)
	}

	// 4. Clean Shell Restoration & Minimalist Summary Card Output
	if m, ok := finalModel.(tui.AppModel); ok && m.PurgeReport != nil {
		printSummaryCard(m.PurgeReport, cfg.DryRun)
	}
}

func printSummaryCard(report *engine.PurgeReport, dryRun bool) {
	title := "⚡ SKILLYFIY PURGE REPORT"
	if dryRun {
		title = "🔍 SKILLYFIY DRY-RUN REPORT (SIMULATION)"
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorMintGreen).
		Padding(1, 2).
		Margin(1, 0).
		Width(64)

	var b strings.Builder
	b.WriteString(tui.BannerStyle.Render(title) + "\n\n")

	b.WriteString(fmt.Sprintf("  • %-32s %s\n",
		"Skills unlinked:",
		tui.ColorMintGreenStyle(fmt.Sprintf("%d files", report.SkillsUnlinked))))

	b.WriteString(fmt.Sprintf("  • %-32s %s\n",
		"MCP servers deregistered:",
		tui.ColorMintBlueStyle(fmt.Sprintf("%d servers", report.MCPServersPurged))))

	b.WriteString(fmt.Sprintf("  • %-32s %s\n",
		"Total prompt context reclaimed:",
		tui.FooterReclaimedStyle.Render(fmt.Sprintf("~%d tokens", report.ReclaimedTokens))))

	b.WriteString(fmt.Sprintf("  • %-32s %s\n",
		"Storage space reclaimed:",
		tui.InspectorValStyle.Render(fmt.Sprintf("%d bytes", report.ReclaimedBytes))))

	if len(report.Errors) > 0 {
		b.WriteString("\n" + tui.ModalTitleStyle.Render("Encountered Warnings/Errors:") + "\n")
		for _, e := range report.Errors {
			b.WriteString("  " + tui.ModalTitleStyle.Render("✖ "+e) + "\n")
		}
	} else {
		b.WriteString("\n" + tui.NormalDescStyle.Render("✔ All operations completed successfully without errors.") + "\n")
	}

	fmt.Println(cardStyle.Render(b.String()))
}
