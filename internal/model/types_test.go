package model_test

import (
	"testing"

	"skillyfiy/internal/model"
)

func TestSortModeString(t *testing.T) {
	tests := []struct {
		mode     model.SortMode
		expected string
	}{
		{model.SortAlphabetical, "[Sort: A-Z]"},
		{model.SortTokenWeight, "[Sort: Tokens ↓]"},
		{model.SortItemType, "[Sort: Type]"},
	}

	for _, tc := range tests {
		if tc.mode.String() != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, tc.mode.String())
		}
	}
}

func TestAgentItemMethods(t *testing.T) {
	item := model.AgentItem{
		ID:          "skill:test",
		Name:        "test-skill",
		Type:        model.TypeSkill,
		SourcePath:  "/tmp/test.md",
		Description: "A test description",
		Tokens:      42,
	}

	if item.Title() != "test-skill" {
		t.Errorf("Expected Title 'test-skill', got '%s'", item.Title())
	}

	filterVal := item.FilterValue()
	if filterVal != "test-skill A test description /tmp/test.md" {
		t.Errorf("Unexpected FilterValue: %s", filterVal)
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
