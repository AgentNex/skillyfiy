package engine

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mitchellh/go-homedir"

	"skillyfiy/internal/config"
	"skillyfiy/internal/model"
)

// SkillDisplayName returns a clean display identifier for a skill item,
// prioritizing the relative path computed by ScanSkills.
func SkillDisplayName(item model.AgentItem) string {
	if item.Name != "" {
		return item.Name
	}
	if rel, ok := item.Details["Relative"]; ok && rel != "" {
		return rel
	}
	if item.SourcePath != "" {
		return item.SourcePath
	}
	if item.ID != "" {
		return strings.TrimPrefix(item.ID, "skill:")
	}
	return "(unnamed skill)"
}

// MCPDisplayName returns the server name for an MCP server item.
func MCPDisplayName(item model.AgentItem) string {
	if item.Name != "" {
		return item.Name
	}
	if srv, ok := item.Details["Server"]; ok && srv != "" {
		return srv
	}
	if item.ID != "" {
		return strings.TrimPrefix(item.ID, "mcp:")
	}
	return "(unnamed server)"
}

// RunList discovers skills and MCP servers non-interactively using the existing
// scanner and parser, and writes the formatted inventory to stdout.
// Warnings and errors are written to stderr.
// It returns exit code 0 on success or valid empty inventory, and 1 on error.
func RunList(cfg *config.Config, stdout io.Writer, stderr io.Writer, extraArgs ...string) int {
	// Parse any trailing flags passed after the "list" argument (e.g. sky list --skills /path)
	for i := 0; i < len(extraArgs); i++ {
		arg := extraArgs[i]
		switch {
		case arg == "-h" || arg == "--help" || arg == "-help":
			fmt.Fprintln(stdout, "Skillyfiy list — Non-interactive skill & MCP server inventory")
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Usage:")
			fmt.Fprintln(stdout, "  sky list [flags]   (or: skillyfiy list [flags])")
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Flags:")
			fmt.Fprintln(stdout, "  --skills <path>      Custom path to skills directory")
			fmt.Fprintln(stdout, "  --mcp <path>         Custom path to MCP configuration file")
			fmt.Fprintln(stdout, "  -h, --help           Show help documentation")
			return 0
		case arg == "--skills" || arg == "-skills":
			if i+1 < len(extraArgs) {
				i++
				expanded, err := homedir.Expand(extraArgs[i])
				if err == nil {
					cfg.SkillsPath = expanded
				} else {
					cfg.SkillsPath = extraArgs[i]
				}
			}
		case strings.HasPrefix(arg, "--skills="):
			val := strings.TrimPrefix(arg, "--skills=")
			if expanded, err := homedir.Expand(val); err == nil {
				cfg.SkillsPath = expanded
			} else {
				cfg.SkillsPath = val
			}
		case strings.HasPrefix(arg, "-skills="):
			val := strings.TrimPrefix(arg, "-skills=")
			if expanded, err := homedir.Expand(val); err == nil {
				cfg.SkillsPath = expanded
			} else {
				cfg.SkillsPath = val
			}
		case arg == "--mcp" || arg == "-mcp":
			if i+1 < len(extraArgs) {
				i++
				expanded, err := homedir.Expand(extraArgs[i])
				if err == nil {
					cfg.MCPPath = expanded
				} else {
					cfg.MCPPath = extraArgs[i]
				}
			}
		case strings.HasPrefix(arg, "--mcp="):
			val := strings.TrimPrefix(arg, "--mcp=")
			if expanded, err := homedir.Expand(val); err == nil {
				cfg.MCPPath = expanded
			} else {
				cfg.MCPPath = val
			}
		case strings.HasPrefix(arg, "-mcp="):
			val := strings.TrimPrefix(arg, "-mcp=")
			if expanded, err := homedir.Expand(val); err == nil {
				cfg.MCPPath = expanded
			} else {
				cfg.MCPPath = val
			}
		}
	}

	hasError := false

	// 1. Discover skills using existing scanner
	skillItems, skillErr := ScanSkills(cfg.SkillsPath)
	if skillErr != nil {
		hasError = true
		fmt.Fprintf(stderr, "Error scanning skills directory (%s): %v\n", cfg.SkillsPath, skillErr)
	}

	// 2. Discover MCP servers using existing parser
	mcpItems, mcpErr := ParseMCP(cfg.MCPPath)
	if mcpErr != nil {
		hasError = true
		fmt.Fprintf(stderr, "Error parsing MCP manifest (%s): %v\n", cfg.MCPPath, mcpErr)
	}

	// 3. Print inventory header
	fmt.Fprintln(stdout, "Skillyfiy — Inventory")
	fmt.Fprintln(stdout)

	// 4. Print SKILLS section
	fmt.Fprintf(stdout, "SKILLS (Total: %d)\n", len(skillItems))
	fmt.Fprintln(stdout, strings.Repeat("─", 32))
	fmt.Fprintln(stdout)

	if len(skillItems) == 0 {
		fmt.Fprintln(stdout, "  (No skill entries found)")
	} else {
		// Sort copy deterministically
		sortedSkills := make([]model.AgentItem, len(skillItems))
		copy(sortedSkills, skillItems)
		sort.Slice(sortedSkills, func(i, j int) bool {
			nameI := strings.ToLower(SkillDisplayName(sortedSkills[i]))
			nameJ := strings.ToLower(SkillDisplayName(sortedSkills[j]))
			if nameI == nameJ {
				return SkillDisplayName(sortedSkills[i]) < SkillDisplayName(sortedSkills[j])
			}
			return nameI < nameJ
		})
		for i, item := range sortedSkills {
			fmt.Fprintf(stdout, "%d. %s\n", i+1, SkillDisplayName(item))
		}
	}
	fmt.Fprintln(stdout)

	// 5. Print MCP SERVERS section
	fmt.Fprintf(stdout, "MCP SERVERS (Total: %d)\n", len(mcpItems))
	fmt.Fprintln(stdout, strings.Repeat("─", 32))
	fmt.Fprintln(stdout)

	if len(mcpItems) == 0 {
		fmt.Fprintln(stdout, "  (No MCP servers found)")
	} else {
		// Sort copy deterministically
		sortedMCP := make([]model.AgentItem, len(mcpItems))
		copy(sortedMCP, mcpItems)
		sort.Slice(sortedMCP, func(i, j int) bool {
			nameI := strings.ToLower(MCPDisplayName(sortedMCP[i]))
			nameJ := strings.ToLower(MCPDisplayName(sortedMCP[j]))
			if nameI == nameJ {
				return MCPDisplayName(sortedMCP[i]) < MCPDisplayName(sortedMCP[j])
			}
			return nameI < nameJ
		})
		for i, item := range sortedMCP {
			fmt.Fprintf(stdout, "%d. %s\n", i+1, MCPDisplayName(item))
		}
	}
	fmt.Fprintln(stdout)

	// 6. Print summary footer
	fmt.Fprintln(stdout, strings.Repeat("─", 32))
	fmt.Fprintf(stdout, "Total skill entries: %d\n", len(skillItems))
	fmt.Fprintf(stdout, "Total MCP servers: %d\n", len(mcpItems))

	if hasError {
		return 1
	}
	return 0
}
