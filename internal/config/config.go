package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// Config holds runtime options parsed from flags and default environment paths.
type Config struct {
	SkillsPath string
	MCPPath    string
	ConfigPath string
	DryRun     bool
	Version    string
}

// DefaultSkillsPath returns the standard agent skills directory path (~/.agent/skills).
func DefaultSkillsPath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".agent", "skills")
}

// DefaultMCPPath returns the standard Claude settings.json path (~/.claude/settings.json).
func DefaultMCPPath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// Load constructs a Config struct from given paths and applies defaults.
func Load(skillsPath, mcpPath, configPath string, dryRun bool) *Config {
	cfg := &Config{
		DryRun: dryRun,
	}

	// Resolve Skills Path
	if skillsPath != "" {
		expanded, err := homedir.Expand(skillsPath)
		if err == nil {
			cfg.SkillsPath = expanded
		} else {
			cfg.SkillsPath = skillsPath
		}
	} else {
		cfg.SkillsPath = DefaultSkillsPath()
		// If ~/.agent/skills does not exist, check common fallback agent directories
		if _, err := os.Stat(cfg.SkillsPath); os.IsNotExist(err) {
			home, _ := homedir.Dir()
			candidate1 := filepath.Join(home, ".agents", "skills")
			candidate2 := filepath.Join(home, ".gemini", "config", "skills")
			if _, err := os.Stat(candidate1); err == nil {
				cfg.SkillsPath = candidate1
			} else if _, err := os.Stat(candidate2); err == nil {
				cfg.SkillsPath = candidate2
			}
		}
	}

	// Resolve MCP Path
	if mcpPath != "" {
		expanded, err := homedir.Expand(mcpPath)
		if err == nil {
			cfg.MCPPath = expanded
		} else {
			cfg.MCPPath = mcpPath
		}
	} else {
		cfg.MCPPath = DefaultMCPPath()
		// If ~/.claude/settings.json does not exist, check common fallback paths
		if _, err := os.Stat(cfg.MCPPath); os.IsNotExist(err) {
			home, _ := homedir.Dir()
			candidates := []string{
				filepath.Join(home, ".gemini", "antigravity-cli", "mcp_config.json"),
				filepath.Join(home, ".mcp.json"),
				filepath.Join(home, ".gemini", "config", "mcp_config.json"),
				filepath.Join(home, ".config", "claude", "settings.json"),
			}
			for _, cand := range candidates {
				if _, err := os.Stat(cand); err == nil {
					cfg.MCPPath = cand
					break
				}
			}
		}
	}

	// Resolve Config Path
	if configPath != "" {
		expanded, err := homedir.Expand(configPath)
		if err == nil {
			cfg.ConfigPath = expanded
		} else {
			cfg.ConfigPath = configPath
		}
	}

	return cfg
}

// ParseFlags parses command-line flags and expands user home directories.
func ParseFlags() (*Config, bool) {
	skillsFlag := flag.String("skills", "", "Path to skills directory (default: ~/.agent/skills)")
	mcpFlag := flag.String("mcp", "", "Path to MCP JSON config (default: ~/.claude/settings.json)")
	configFlag := flag.String("config", "", "Optional custom configuration file path")
	dryRunFlag := flag.Bool("dry-run", false, "Simulate purge without writing changes or deleting files")
	versionFlag := flag.Bool("v", false, "Print current version and exit")
	versionLongFlag := flag.Bool("version", false, "Print current version and exit")

	flag.Usage = func() {
		fmt.Printf("Skillyfiy CLI — High-Performance Agent Skill & MCP Pruner\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  skillyfiy [flags]\n\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("  -v, --version        Print version information and exit\n")
		fmt.Printf("  --skills <path>      Custom path to skills directory\n")
		fmt.Printf("  --mcp <path>         Custom path to MCP configuration file\n")
		fmt.Printf("  --config <path>      Custom path to generic configuration\n")
		fmt.Printf("  --dry-run            Simulate purge without writing changes or deleting files\n")
		fmt.Printf("  -h, --help           Show help documentation\n\n")
		fmt.Printf("Keybindings:\n")
		fmt.Printf("  Space / x            Toggle single item\n")
		fmt.Printf("  s / v                Toggle Skillyfiy Span (Visual Range mode)\n")
		fmt.Printf("  Ctrl+A               Toggle select all visible items\n")
		fmt.Printf("  Tab                  Switch between List and Inspector pane\n")
		fmt.Printf("  Ctrl+S / F1          Cycle sort mode (A-Z, Tokens, Type)\n")
		fmt.Printf("  /                    Focus agent search prompt\n")
		fmt.Printf("  Enter                Review and confirm purge\n")
		fmt.Printf("  Ctrl+Q / Ctrl+C      Exit without making changes\n")
	}

	flag.Parse()

	if *versionFlag || *versionLongFlag {
		return nil, true
	}

	cfg := Load(*skillsFlag, *mcpFlag, *configFlag, *dryRunFlag)
	return cfg, false
}
