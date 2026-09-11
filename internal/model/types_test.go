package model_test

import (
	"testing"
	"time"

	"skillyfiy/internal/model"
)

func TestSortModeString(t *testing.T) {
	tests := []struct {
		mode            model.SortMode
		expected        string
		expectedCompact string
	}{
		{model.SortNewest, "[Sort: Newest]", "[Newest]"},
		{model.SortOldest, "[Sort: Oldest]", "[Oldest]"},
		{model.SortLargest, "[Sort: Largest]", "[Largest]"},
		{model.SortSmallest, "[Sort: Smallest]", "[Smallest]"},
		{model.SortAZ, "[Sort: A-Z]", "[A-Z]"},
		{model.SortZA, "[Sort: Z-A]", "[Z-A]"},
		{model.SortMode(999), "[Sort: Newest]", "[Newest]"},
	}

	for _, tc := range tests {
		if tc.mode.String() != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, tc.mode.String())
		}
		if tc.mode.Compact() != tc.expectedCompact {
			t.Errorf("Expected compact %s, got %s", tc.expectedCompact, tc.mode.Compact())
		}
	}
}

func TestFilterTypeString(t *testing.T) {
	tests := []struct {
		filter          model.FilterType
		expected        string
		expectedCompact string
	}{
		{model.FilterAll, "[Filter: All]", "[All]"},
		{model.FilterSkills, "[Filter: Skills]", "[Skills]"},
		{model.FilterMCP, "[Filter: MCP]", "[MCP]"},
		{model.FilterType(999), "[Filter: All]", "[All]"},
	}

	for _, tc := range tests {
		if tc.filter.String() != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, tc.filter.String())
		}
		if tc.filter.Compact() != tc.expectedCompact {
			t.Errorf("Expected compact %s, got %s", tc.expectedCompact, tc.filter.Compact())
		}
	}
}

func TestAgentItemMethods(t *testing.T) {
	now := time.Now()
	item := model.AgentItem{
		ID:          "skill:test",
		Name:        "Test-Skill",
		Type:        model.TypeSkill,
		SourcePath:  "/tmp/Test.md",
		Description: "A Test Description",
		Tokens:      42,
		ModTime:     now,
	}

	if item.Title() != "Test-Skill" {
		t.Errorf("Expected Title 'Test-Skill', got '%s'", item.Title())
	}

	filterVal := item.FilterValue()
	if filterVal != "Test-Skill A Test Description /tmp/Test.md" {
		t.Errorf("Unexpected FilterValue: %s", filterVal)
	}

	item.EnsureCache()
	if item.NameLower != "test-skill" {
		t.Errorf("Expected NameLower 'test-skill', got '%s'", item.NameLower)
	}
	if item.DescLower != "a test description" {
		t.Errorf("Expected DescLower 'a test description', got '%s'", item.DescLower)
	}
	if item.PathLower != "/tmp/test.md" {
		t.Errorf("Expected PathLower '/tmp/test.md', got '%s'", item.PathLower)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{2048, "2.0 KB"},
		{1048576, "1.0 MB"},
	}

	for _, tc := range tests {
		res := model.FormatBytes(tc.bytes)
		if res != tc.expected {
			t.Errorf("For %d bytes, expected %s, got %s", tc.bytes, tc.expected, res)
		}
	}
}
