package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillyfiy/internal/config"
)

func TestDefaultPaths(t *testing.T) {
	skillsPath := config.DefaultSkillsPath()
	if !strings.HasSuffix(skillsPath, filepath.Join(".agent", "skills")) {
		t.Errorf("Unexpected DefaultSkillsPath: %s", skillsPath)
	}

	mcpPath := config.DefaultMCPPath()
	if !strings.HasSuffix(mcpPath, filepath.Join(".claude", "settings.json")) {
		t.Errorf("Unexpected DefaultMCPPath: %s", mcpPath)
	}
}

func TestKeyMapBindings(t *testing.T) {
	km := config.DefaultKeyMap

	if len(km.CursorUp.Keys()) == 0 {
		t.Errorf("CursorUp keys empty")
	}
	if len(km.CursorDown.Keys()) == 0 {
		t.Errorf("CursorDown keys empty")
	}
	if len(km.ToggleSelect.Keys()) == 0 {
		t.Errorf("ToggleSelect keys empty")
	}
	if len(km.VisualRange.Keys()) == 0 {
		t.Errorf("VisualRange keys empty")
	}
	if len(km.SelectAll.Keys()) == 0 {
		t.Errorf("SelectAll keys empty")
	}
	if len(km.FocusSearch.Keys()) == 0 {
		t.Errorf("FocusSearch keys empty")
	}
	if len(km.SwitchPane.Keys()) == 0 {
		t.Errorf("SwitchPane keys empty")
	}
	if len(km.CycleSort.Keys()) == 0 {
		t.Errorf("CycleSort keys empty")
	}
	if len(km.ConfirmPurge.Keys()) == 0 {
		t.Errorf("ConfirmPurge keys empty")
	}
	if len(km.Quit.Keys()) == 0 {
		t.Errorf("Quit keys empty")
	}
	if len(km.ForceQuit.Keys()) == 0 {
		t.Errorf("ForceQuit keys empty")
	}
}

func TestParseFlagsDefaults(t *testing.T) {
	// Reset CommandLine flags for testing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"skillyfiy"}

	cfg, isVersion := config.ParseFlags()
	if isVersion {
		t.Fatalf("Expected isVersion to be false")
	}

	if cfg.SkillsPath == "" {
		t.Errorf("Expected non-empty default SkillsPath")
	}
	if cfg.MCPPath == "" {
		t.Errorf("Expected non-empty default MCPPath")
	}
	if cfg.DryRun != false {
		t.Errorf("Expected default DryRun to be false")
	}
}
