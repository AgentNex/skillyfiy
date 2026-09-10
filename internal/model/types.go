package model

import "fmt"

// ItemType defines whether an agent item is a skill or an MCP server.
type ItemType string

const (
	TypeSkill ItemType = "SKILL"
	TypeMCP   ItemType = "MCP"
)

// SortMode defines the active sorting criteria for the items list.
type SortMode int

const (
	SortAlphabetical SortMode = iota
	SortTokenWeight
	SortItemType
)

// String returns the user-facing badge string for the current sort mode.
func (s SortMode) String() string {
	switch s {
	case SortAlphabetical:
		return "[Sort: A-Z]"
	case SortTokenWeight:
		return "[Sort: Tokens ↓]"
	case SortItemType:
		return "[Sort: Type]"
	default:
		return "[Sort: A-Z]"
	}
}

// AgentItem represents an AI agent skill file or an MCP server configuration entry.
type AgentItem struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        ItemType          `json:"type"`
	SourcePath  string            `json:"source_path"`
	Description string            `json:"description"`
	Tokens      int               `json:"tokens"`
	Size        int64             `json:"size"`
	Details     map[string]string `json:"details"`
	RawPreview  string            `json:"raw_preview"`
	Selected    bool              `json:"-"`
}

// Title returns the display name for the item.
func (i AgentItem) Title() string {
	return i.Name
}

// FilterValue satisfies bubbles/list.Item interface.
func (i AgentItem) FilterValue() string {
	return i.Name + " " + i.Description + " " + i.SourcePath
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
