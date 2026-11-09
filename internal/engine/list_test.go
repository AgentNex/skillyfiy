package engine_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillyfiy/internal/config"
	"skillyfiy/internal/engine"
	"skillyfiy/internal/model"
)

func TestRunList_SuccessWithSkillsAndMCP(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create nested skills
	skillsDir := filepath.Join(tempDir, "skills")
	subDir1 := filepath.Join(skillsDir, "frontend-design")
	subDir2 := filepath.Join(skillsDir, "python-automation", "deep")
	if err := os.MkdirAll(subDir1, 0755); err != nil {
		t.Fatalf("Failed to create subDir1: %v", err)
	}
	if err := os.MkdirAll(subDir2, 0755); err != nil {
		t.Fatalf("Failed to create subDir2: %v", err)
	}

	skill1 := filepath.Join(subDir1, "SKILL.md")
	skill2 := filepath.Join(subDir1, "helper.py")
	skill3 := filepath.Join(subDir2, "script.sh")
	skill4 := filepath.Join(skillsDir, "root-tool.json")

	if err := os.WriteFile(skill1, []byte("---\ndescription: Frontend skill\n---\n# Skill"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill2, []byte("print('hello')"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill3, []byte("#!/bin/bash\necho 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill4, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Create MCP manifest with 3 servers
	mcpPath := filepath.Join(tempDir, "settings.json")
	mcpJSON := `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "SECRET_TOKEN": "SUPER_SECRET_12345"
      }
    },
    "context7": {
      "command": "node",
      "args": ["index.js"]
    }
  }
}`
	if err := os.WriteFile(mcpPath, []byte(mcpJSON), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		SkillsPath: skillsDir,
		MCPPath:    mcpPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := engine.RunList(cfg, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected empty stderr, got: %s", stderr.String())
	}

	out := stdout.String()

	// Verify Header
	if !strings.Contains(out, "Skillyfiy — Inventory") {
		t.Errorf("Expected inventory header in output, got:\n%s", out)
	}

	// Verify SKILLS Section & Total
	if !strings.Contains(out, "SKILLS (Total: 4)") {
		t.Errorf("Expected SKILLS (Total: 4) in output, got:\n%s", out)
	}
	expectedSkills := []string{
		"frontend-design/helper.py",
		"frontend-design/SKILL.md",
		"python-automation/deep/script.sh",
		"root-tool.json",
	}
	for _, expected := range expectedSkills {
		if !strings.Contains(out, expected) {
			t.Errorf("Missing expected skill '%s' in output:\n%s", expected, out)
		}
	}

	// Verify MCP Section & Total
	if !strings.Contains(out, "MCP SERVERS (Total: 3)") {
		t.Errorf("Expected MCP SERVERS (Total: 3) in output, got:\n%s", out)
	}
	expectedServers := []string{"context7", "filesystem", "github"}
	for _, expected := range expectedServers {
		if !strings.Contains(out, expected) {
			t.Errorf("Missing expected MCP server '%s' in output:\n%s", expected, out)
		}
	}

	// Verify Summary Footer
	if !strings.Contains(out, "Total skill entries: 4") {
		t.Errorf("Missing 'Total skill entries: 4' in output:\n%s", out)
	}
	if !strings.Contains(out, "Total MCP servers: 3") {
		t.Errorf("Missing 'Total MCP servers: 3' in output:\n%s", out)
	}

	// Verify sensitive values are NEVER printed
	if strings.Contains(out, "SUPER_SECRET_12345") {
		t.Errorf("Leaked secret environment variable value in output!")
	}
}

func TestRunList_EmptyInventory(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "empty-skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	mcpPath := filepath.Join(tempDir, "settings.json")
	if err := os.WriteFile(mcpPath, []byte(`{"mcpServers": {}}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		SkillsPath: skillsDir,
		MCPPath:    mcpPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := engine.RunList(cfg, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for empty inventory, got %d. Stderr: %s", exitCode, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "SKILLS (Total: 0)") {
		t.Errorf("Expected SKILLS (Total: 0), got:\n%s", out)
	}
	if !strings.Contains(out, "(No skill entries found)") {
		t.Errorf("Expected empty skill message, got:\n%s", out)
	}
	if !strings.Contains(out, "MCP SERVERS (Total: 0)") {
		t.Errorf("Expected MCP SERVERS (Total: 0), got:\n%s", out)
	}
	if !strings.Contains(out, "(No MCP servers found)") {
		t.Errorf("Expected empty MCP message, got:\n%s", out)
	}
	if !strings.Contains(out, "Total skill entries: 0") {
		t.Errorf("Expected 'Total skill entries: 0', got:\n%s", out)
	}
	if !strings.Contains(out, "Total MCP servers: 0") {
		t.Errorf("Expected 'Total MCP servers: 0', got:\n%s", out)
	}
}

func TestRunList_NonExistentPaths(t *testing.T) {
	cfg := &config.Config{
		SkillsPath: "/non/existent/path/for/skills",
		MCPPath:    "/non/existent/path/settings.json",
	}

	var stdout, stderr bytes.Buffer
	exitCode := engine.RunList(cfg, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for non-existent default paths, got %d", exitCode)
	}

	out := stdout.String()
	if !strings.Contains(out, "Total skill entries: 0") || !strings.Contains(out, "Total MCP servers: 0") {
		t.Errorf("Expected 0 totals for non-existent paths, got:\n%s", out)
	}
}

func TestRunList_InvalidMCPJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Valid skill
	skillsDir := filepath.Join(tempDir, "skills")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "test.md"), []byte("# Test"), 0644)

	// Malformed JSON MCP file
	mcpPath := filepath.Join(tempDir, "malformed.json")
	os.WriteFile(mcpPath, []byte(`{ "mcpServers": { broken json `), 0644)

	cfg := &config.Config{
		SkillsPath: skillsDir,
		MCPPath:    mcpPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := engine.RunList(cfg, &stdout, &stderr)

	// Malformed JSON should return non-zero exit code
	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code for invalid MCP JSON, got 0")
	}

	// Stderr must report the error
	if !strings.Contains(stderr.String(), "invalid JSON in MCP manifest") {
		t.Errorf("Expected invalid JSON error in stderr, got: %s", stderr.String())
	}

	// Partial success: skills must still be listed in stdout
	out := stdout.String()
	if !strings.Contains(out, "SKILLS (Total: 1)") {
		t.Errorf("Expected SKILLS (Total: 1) despite MCP failure, got:\n%s", out)
	}
	if !strings.Contains(out, "test.md") {
		t.Errorf("Expected test.md in stdout, got:\n%s", out)
	}
	if !strings.Contains(out, "Total skill entries: 1") {
		t.Errorf("Expected Total skill entries: 1, got:\n%s", out)
	}
}

func TestRunList_SkillDirectoryError(t *testing.T) {
	tempDir := t.TempDir()

	// Skill path pointing to a regular file instead of a directory
	badSkillPath := filepath.Join(tempDir, "not-a-dir.txt")
	os.WriteFile(badSkillPath, []byte("regular file"), 0644)

	// Valid MCP
	mcpPath := filepath.Join(tempDir, "mcp.json")
	os.WriteFile(mcpPath, []byte(`{"mcpServers": {"demo-server": {"command": "echo"}}}`), 0644)

	cfg := &config.Config{
		SkillsPath: badSkillPath,
		MCPPath:    mcpPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := engine.RunList(cfg, &stdout, &stderr)

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code on skill dir error, got 0")
	}

	if !strings.Contains(stderr.String(), "skills path is not a directory") {
		t.Errorf("Expected skill error in stderr, got: %s", stderr.String())
	}

	// MCP servers should still be listed in stdout
	out := stdout.String()
	if !strings.Contains(out, "MCP SERVERS (Total: 1)") {
		t.Errorf("Expected MCP SERVERS (Total: 1) despite skill scan failure, got:\n%s", out)
	}
	if !strings.Contains(out, "demo-server") {
		t.Errorf("Expected demo-server in output, got:\n%s", out)
	}
}

func TestRunList_DeterministicSorting(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")
	os.MkdirAll(skillsDir, 0755)

	names := []string{"zebra.py", "apple.md", "mango.sh", "banana.json"}
	for _, n := range names {
		os.WriteFile(filepath.Join(skillsDir, n), []byte("# content"), 0644)
	}

	mcpPath := filepath.Join(tempDir, "mcp.json")
	mcpJSON := `{
  "mcpServers": {
    "z-server": {"command": "echo"},
    "a-server": {"command": "echo"},
    "m-server": {"command": "echo"}
  }
}`
	os.WriteFile(mcpPath, []byte(mcpJSON), 0644)

	cfg := &config.Config{
		SkillsPath: skillsDir,
		MCPPath:    mcpPath,
	}

	var out1, err1 bytes.Buffer
	var out2, err2 bytes.Buffer

	c1 := engine.RunList(cfg, &out1, &err1)
	c2 := engine.RunList(cfg, &out2, &err2)

	if c1 != 0 || c2 != 0 {
		t.Fatalf("Expected exit code 0, got %d, %d", c1, c2)
	}

	if out1.String() != out2.String() {
		t.Errorf("Outputs are not deterministic!\nRun 1:\n%s\nRun 2:\n%s", out1.String(), out2.String())
	}

	// Verify alphabetical order for skills: apple, banana, mango, zebra
	outStr := out1.String()
	posA := strings.Index(outStr, "apple.md")
	posB := strings.Index(outStr, "banana.json")
	posM := strings.Index(outStr, "mango.sh")
	posZ := strings.Index(outStr, "zebra.py")

	if !(posA < posB && posB < posM && posM < posZ) {
		t.Errorf("Skills not sorted alphabetically: posA=%d, posB=%d, posM=%d, posZ=%d", posA, posB, posM, posZ)
	}

	// Verify alphabetical order for MCP: a-server, m-server, z-server
	posSA := strings.Index(outStr, "a-server")
	posSM := strings.Index(outStr, "m-server")
	posSZ := strings.Index(outStr, "z-server")

	if !(posSA < posSM && posSM < posSZ) {
		t.Errorf("MCP servers not sorted alphabetically: posSA=%d, posSM=%d, posSZ=%d", posSA, posSM, posSZ)
	}
}

func TestRunList_ExtraArgsOverrides(t *testing.T) {
	tempDir := t.TempDir()

	customSkills := filepath.Join(tempDir, "custom-skills")
	os.MkdirAll(customSkills, 0755)
	os.WriteFile(filepath.Join(customSkills, "override.md"), []byte("# override"), 0644)

	customMCP := filepath.Join(tempDir, "custom-mcp.json")
	os.WriteFile(customMCP, []byte(`{"mcpServers": {"custom-srv": {"command": "echo"}}}`), 0644)

	cfg := &config.Config{
		SkillsPath: "/default/skills",
		MCPPath:    "/default/mcp.json",
	}

	var stdout, stderr bytes.Buffer
	code := engine.RunList(cfg, &stdout, &stderr, "--skills", customSkills, "--mcp", customMCP)

	if code != 0 {
		t.Fatalf("Expected 0, got %d. Stderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "override.md") {
		t.Errorf("Expected override.md from custom skills path, got:\n%s", out)
	}
	if !strings.Contains(out, "custom-srv") {
		t.Errorf("Expected custom-srv from custom MCP path, got:\n%s", out)
	}
}

func TestDisplayNameFallbacks(t *testing.T) {
	// Skill with empty Name falls back to Details["Relative"]
	itemSkill := model.AgentItem{
		Details: map[string]string{"Relative": "fallback/path.md"},
	}
	if name := engine.SkillDisplayName(itemSkill); name != "fallback/path.md" {
		t.Errorf("Expected fallback/path.md, got %s", name)
	}

	// MCP with empty Name falls back to Details["Server"]
	itemMCP := model.AgentItem{
		Details: map[string]string{"Server": "fallback-server"},
	}
	if name := engine.MCPDisplayName(itemMCP); name != "fallback-server" {
		t.Errorf("Expected fallback-server, got %s", name)
	}
}
