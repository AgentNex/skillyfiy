package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"skillyfiy/internal/config"
	"skillyfiy/internal/engine"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintSummaryCard(t *testing.T) {
	// Test Live Purge Summary
	report := &engine.PurgeReport{
		SkillsUnlinked:   2,
		MCPServersPurged: 1,
		ReclaimedTokens:  450,
		ReclaimedBytes:   1800,
		Errors:           nil,
	}

	out := captureStdout(func() {
		printSummaryCard(report, false)
	})

	if !strings.Contains(out, "SKILLYFIY PURGE REPORT") {
		t.Errorf("Expected live report title in output, got: %s", out)
	}
	if !strings.Contains(out, "2 files") {
		t.Errorf("Expected 2 files unlinked in output, got: %s", out)
	}
	if !strings.Contains(out, "1 servers") {
		t.Errorf("Expected 1 server purged in output, got: %s", out)
	}
	if !strings.Contains(out, "~450 tokens") {
		t.Errorf("Expected ~450 tokens in output, got: %s", out)
	}

	// Test Dry-Run Summary with Errors
	reportDry := &engine.PurgeReport{
		SkillsUnlinked:   0,
		MCPServersPurged: 0,
		ReclaimedTokens:  0,
		ReclaimedBytes:   0,
		Errors:           []string{"Permission denied on /tmp/protected.md"},
	}

	outDry := captureStdout(func() {
		printSummaryCard(reportDry, true)
	})

	if !strings.Contains(outDry, "DRY-RUN REPORT") {
		t.Errorf("Expected dry-run report title in output, got: %s", outDry)
	}
	if !strings.Contains(outDry, "Permission denied") {
		t.Errorf("Expected error message in output, got: %s", outDry)
	}
}

func TestListCommandOutput(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := tempDir + "/skills"
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(skillsDir+"/test.md", []byte("# Skill"), 0644)

	mcpPath := tempDir + "/settings.json"
	os.WriteFile(mcpPath, []byte(`{"mcpServers": {"test-mcp": {"command": "echo"}}}`), 0644)

	var stdout, stderr bytes.Buffer
	cfg := &config.Config{
		SkillsPath: skillsDir,
		MCPPath:    mcpPath,
	}

	code := engine.RunList(cfg, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Expected 0, got %d. Stderr: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Skillyfiy — Inventory") {
		t.Errorf("Expected header in output, got: %s", out)
	}
	if !strings.Contains(out, "1. test.md") {
		t.Errorf("Expected skill test.md in output, got: %s", out)
	}
	if !strings.Contains(out, "1. test-mcp") {
		t.Errorf("Expected mcp test-mcp in output, got: %s", out)
	}
}
