package model

import (
	"fmt"
	"strings"
	"time"
)

// ItemType defines whether an agent item is a skill or an MCP server.
type ItemType string

const (
	TypeSkill ItemType = "SKILL"
	TypeMCP   ItemType = "MCP"
)

// SortMode defines the active sorting criteria for the items list.
type SortMode int

const (
	SortNewest SortMode = iota
	SortOldest
	SortLargest
	SortSmallest
	SortAZ
	SortZA
)

// Legacy aliases for backward compatibility.
const (
	SortAlphabetical = SortAZ
	SortTokenWeight  = SortLargest
	SortItemType     = SortNewest
)

// String returns the user-facing badge string for the current sort mode.
func (s SortMode) String() string {
	switch s {
	case SortNewest:
		return "[Sort: Newest]"
	case SortOldest:
		return "[Sort: Oldest]"
	case SortLargest:
		return "[Sort: Largest]"
	case SortSmallest:
		return "[Sort: Smallest]"
	case SortAZ:
		return "[Sort: A-Z]"
	case SortZA:
		return "[Sort: Z-A]"
	default:
		return "[Sort: Newest]"
	}
}

// Compact returns a short badge string for narrow terminal displays.
func (s SortMode) Compact() string {
	switch s {
	case SortNewest:
		return "[Newest]"
	case SortOldest:
		return "[Oldest]"
	case SortLargest:
		return "[Largest]"
	case SortSmallest:
		return "[Smallest]"
	case SortAZ:
		return "[A-Z]"
	case SortZA:
		return "[Z-A]"
	default:
		return "[Newest]"
	}
}

// FilterType defines the resource type filtering criteria.
type FilterType int

const (
	FilterAll FilterType = iota
	FilterSkills
	FilterMCP
)

// String returns the user-facing badge string for the active type filter.
func (f FilterType) String() string {
	switch f {
	case FilterAll:
		return "[Filter: All]"
	case FilterSkills:
		return "[Filter: Skills]"
	case FilterMCP:
		return "[Filter: MCP]"
	default:
		return "[Filter: All]"
	}
}

// Compact returns a short badge string for narrow terminal displays.
func (f FilterType) Compact() string {
	switch f {
	case FilterAll:
		return "[All]"
	case FilterSkills:
		return "[Skills]"
	case FilterMCP:
		return "[MCP]"
	default:
		return "[All]"
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
	ModTime     time.Time         `json:"mod_time"`
	Details     map[string]string `json:"details"`
	RawPreview  string            `json:"raw_preview"`
	Selected    bool              `json:"-"`

	// Pre-cached lowercase strings for zero-alloc tiered search
	NameLower string `json:"-"`
	DescLower string `json:"-"`
	PathLower string `json:"-"`
}

// EnsureCache populates lowercase cached strings if empty.
func (i *AgentItem) EnsureCache() {
	if i.NameLower == "" {
		i.NameLower = strings.ToLower(i.Name)
	}
	if i.DescLower == "" {
		i.DescLower = strings.ToLower(i.Description)
	}
	if i.PathLower == "" {
		i.PathLower = strings.ToLower(i.SourcePath)
	}
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
