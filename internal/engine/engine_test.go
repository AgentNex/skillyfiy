package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillyfiy/internal/engine"
	"skillyfiy/internal/model"
)

func TestScanSkills(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create a markdown skill with YAML frontmatter
	mdPath := filepath.Join(tempDir, "web-search.md")
	mdContent := `---
name: web-search
description: Powerful web searching skill for AI agents
---
# Web Search
This tool searches Google.
`
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatalf("Failed to write md skill: %v", err)
	}

	// 2. Create a python skill in a subdirectory
	subDir := filepath.Join(tempDir, "python-tools")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	pyPath := filepath.Join(subDir, "analyzer.py")
	pyContent := `"""Code analysis tool for inspecting python AST."""
import os
import sys

def analyze():
    pass
`
	if err := os.WriteFile(pyPath, []byte(pyContent), 0644); err != nil {
		t.Fatalf("Failed to write py skill: %v", err)
	}

	// 3. Create a shell skill without frontmatter
	shPath := filepath.Join(tempDir, "deploy.sh")
	shContent := `#!/bin/bash
# Deploy script for services
echo "deploying"
`
	if err := os.WriteFile(shPath, []byte(shContent), 0644); err != nil {
		t.Fatalf("Failed to write sh skill: %v", err)
	}

	// 4. Create an ignored file (.txt) and a hidden file
	if err := os.WriteFile(filepath.Join(tempDir, "ignore.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".hidden.md"), []byte("hidden"), 0644); err != nil {
		t.Fatalf("Failed to write hidden file: %v", err)
	}

	items, err := engine.ScanSkills(tempDir)
	if err != nil {
		t.Fatalf("ScanSkills returned error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items (md, py, sh), got %d", len(items))
	}

	var foundMD, foundPY, foundSH bool
	for _, it := range items {
		if strings.HasSuffix(it.SourcePath, "web-search.md") {
			foundMD = true
			if it.Description != "Powerful web searching skill for AI agents" {
				t.Errorf("Unexpected description: %s", it.Description)
			}
			if it.Type != model.TypeSkill {
				t.Errorf("Expected TypeSkill, got %v", it.Type)
			}
			if it.Tokens <= 0 {
				t.Errorf("Expected positive token count, got %d", it.Tokens)
			}
		}
		if strings.HasSuffix(it.SourcePath, "analyzer.py") {
			foundPY = true
			if it.Tokens <= 0 {
				t.Errorf("Expected positive token count for py, got %d", it.Tokens)
			}
		}
		if strings.HasSuffix(it.SourcePath, "deploy.sh") {
			foundSH = true
			if it.Type != model.TypeSkill {
				t.Errorf("Expected TypeSkill for sh, got %v", it.Type)
			}
		}
	}

	if !foundMD || !foundPY || !foundSH {
		t.Errorf("Did not find all expected skills (md: %v, py: %v, sh: %v)", foundMD, foundPY, foundSH)
	}
}

func TestScanSkillsEdgeCases(t *testing.T) {
	// Non-existent path returns empty items, nil error
	items, err := engine.ScanSkills("/non/existent/path/for/skills")
	if err != nil {
		t.Errorf("Expected nil error for non-existent path, got: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}

	// File path instead of directory returns error
	tempFile := filepath.Join(t.TempDir(), "dummy.txt")
	os.WriteFile(tempFile, []byte("test"), 0644)
	_, err = engine.ScanSkills(tempFile)
	if err == nil {
		t.Errorf("Expected error when scanning a file as directory, got nil")
	}
}

func TestMCPManifestLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "settings.json")

	initialJSON := `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "env": {
        "DEBUG": "1"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "secret"
      }
    }
  }
}`

	if err := os.WriteFile(manifestPath, []byte(initialJSON), 0644); err != nil {
		t.Fatalf("Failed to write initial MCP manifest: %v", err)
	}

	// 1. Parse MCP
	items, err := engine.ParseMCP(manifestPath)
	if err != nil {
		t.Fatalf("ParseMCP returned error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 MCP items, got %d", len(items))
	}

	// 2. Empty server deletion returns nil
	if err := engine.MutateAndSaveMCP(manifestPath, []string{}); err != nil {
		t.Errorf("Expected nil error on empty deletion, got: %v", err)
	}

	// 3. Purge 1 server (github)
	err = engine.MutateAndSaveMCP(manifestPath, []string{"github"})
	if err != nil {
		t.Fatalf("MutateAndSaveMCP returned error: %v", err)
	}

	// Check that backup file was created
	bakPath := manifestPath + ".bak"
	bakData, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("Backup file was not created: %v", err)
	}
	if !strings.Contains(string(bakData), "github") {
		t.Errorf("Backup file missing original github server")
	}

	// Re-parse mutated file
	mutatedItems, err := engine.ParseMCP(manifestPath)
	if err != nil {
		t.Fatalf("Re-parsing mutated manifest returned error: %v", err)
	}
	if len(mutatedItems) != 1 {
		t.Fatalf("Expected 1 MCP item after deletion, got %d", len(mutatedItems))
	}
	if mutatedItems[0].Name != "filesystem" {
		t.Errorf("Expected remaining server to be filesystem, got %s", mutatedItems[0].Name)
	}
}

func TestMCPEdgeCases(t *testing.T) {
	// Non-existent file returns empty items, nil error
	items, err := engine.ParseMCP("/non/existent/settings.json")
	if err != nil {
		t.Errorf("Expected nil error for non-existent manifest, got: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}

	// Directory path instead of file returns error
	tempDir := t.TempDir()
	_, err = engine.ParseMCP(tempDir)
	if err == nil {
		t.Errorf("Expected error when parsing directory as MCP manifest, got nil")
	}

	// Alternative 'servers' key format
	altPath := filepath.Join(tempDir, "alt.json")
	altJSON := `{
  "servers": {
    "local-test": {
      "command": "node",
      "args": ["server.js"]
    }
  }
}`
	os.WriteFile(altPath, []byte(altJSON), 0644)
	items, err = engine.ParseMCP(altPath)
	if err != nil {
		t.Fatalf("ParseMCP failed on alt format: %v", err)
	}
	if len(items) != 1 || items[0].Name != "local-test" {
		t.Errorf("Expected 1 item named local-test, got: %v", items)
	}
}

func TestExecutePurgeDryRunAndLive(t *testing.T) {
	tempDir := t.TempDir()

	skillFile := filepath.Join(tempDir, "test-skill.md")
	if err := os.WriteFile(skillFile, []byte("some skill content here"), 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}

	mcpFile := filepath.Join(tempDir, "claude.json")
	mcpJSON := `{
  "mcpServers": {
    "to-remove": {
      "command": "echo",
      "args": ["hello"]
    }
  }
}`
	if err := os.WriteFile(mcpFile, []byte(mcpJSON), 0644); err != nil {
		t.Fatalf("Failed to create mcp file: %v", err)
	}

	items := []model.AgentItem{
		{
			ID:         "skill:test-skill.md",
			Name:       "test-skill.md",
			Type:       model.TypeSkill,
			SourcePath: skillFile,
			Tokens:     20,
			Size:       80,
			Selected:   true,
		},
		{
			ID:         "mcp:to-remove",
			Name:       "to-remove",
			Type:       model.TypeMCP,
			SourcePath: mcpFile,
			Tokens:     50,
			Size:       120,
			Selected:   true,
		},
	}

	// 1. Dry Run
	dryReport, err := engine.ExecutePurge(items, true)
	if err != nil {
		t.Fatalf("ExecutePurge dry-run error: %v", err)
	}
	if dryReport.SkillsUnlinked != 1 {
		t.Errorf("Expected 1 dry-run unlinked skill, got %d", dryReport.SkillsUnlinked)
	}
	if dryReport.MCPServersPurged != 1 {
		t.Errorf("Expected 1 dry-run purged MCP, got %d", dryReport.MCPServersPurged)
	}
	if dryReport.ReclaimedTokens != 70 {
		t.Errorf("Expected 70 reclaimed tokens, got %d", dryReport.ReclaimedTokens)
	}

	// Check files still exist on disk
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Fatalf("Skill file was removed during dry-run!")
	}

	// 2. Live Purge
	liveReport, err := engine.ExecutePurge(items, false)
	if err != nil {
		t.Fatalf("ExecutePurge live error: %v", err)
	}
	if liveReport.SkillsUnlinked != 1 {
		t.Errorf("Expected 1 unlinked skill, got %d", liveReport.SkillsUnlinked)
	}
	if liveReport.MCPServersPurged != 1 {
		t.Errorf("Expected 1 unlinked MCP, got %d", liveReport.MCPServersPurged)
	}

	// Check skill file is now removed
	if _, err := os.Stat(skillFile); !os.IsNotExist(err) {
		t.Fatalf("Skill file was NOT removed after live purge")
	}

	// Check MCP server was deleted from json
	data, _ := os.ReadFile(mcpFile)
	if strings.Contains(string(data), "to-remove") {
		t.Errorf("MCP server 'to-remove' still present in manifest")
	}
}
