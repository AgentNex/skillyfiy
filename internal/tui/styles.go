package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Minimalist Color Palette matching the Master Specification.
var (
	ColorBase        = lipgloss.AdaptiveColor{Light: "#121214", Dark: "#121214"}
	ColorMintGreen   = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#10B981"}
	ColorSeaGreen    = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#059669"}
	ColorMintBlue    = lipgloss.AdaptiveColor{Light: "#38BDF8", Dark: "#38BDF8"}
	ColorDullSilver  = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#94A3B8"}
	ColorBrightSlate = lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#E2E8F0"}
	ColorWarningRed  = lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#EF4444"}
	ColorMutedBg     = lipgloss.AdaptiveColor{Light: "#1E1E24", Dark: "#1E1E24"}
	ColorWhite       = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
)

// Lipgloss Styles
var (
	// Header & Banner
	BannerStyle = lipgloss.NewStyle().
			Foreground(ColorMintGreen).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorDullSilver)

	// Search Prompt
	SearchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMintBlue).
			Padding(0, 1)

	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(ColorMintBlue).
				Bold(true)

	SearchTextStyle = lipgloss.NewStyle().
			Foreground(ColorBrightSlate)

	SortBadgeStyle = lipgloss.NewStyle().
			Foreground(ColorMintBlue).
			Background(ColorMutedBg).
			Padding(0, 1).
			Bold(true)

	FilterBadgeStyle = lipgloss.NewStyle().
			Foreground(ColorMintGreen).
			Background(ColorMutedBg).
			Padding(0, 1).
			Bold(true)

	// Split View Panes
	LeftPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDullSilver).
			Padding(0)

	LeftPaneFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMintBlue).
				Padding(0)

	RightPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDullSilver).
			Padding(0, 1)

	RightPaneFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMintBlue).
				Padding(0, 1)

	// Badges
	SkillBadgeStyle = lipgloss.NewStyle().
			Background(ColorSeaGreen).
			Foreground(ColorWhite).
			Padding(0, 1).
			Bold(true)

	MCPBadgeStyle = lipgloss.NewStyle().
			Background(ColorMintBlue).
			Foreground(ColorBase).
			Padding(0, 1).
			Bold(true)

	TokenWeightStyle = lipgloss.NewStyle().
				Foreground(ColorMintGreen).
				Bold(true)

	// Item List Rendering
	NormalTitleStyle = lipgloss.NewStyle().
				Foreground(ColorBrightSlate)

	SelectedTitleStyle = lipgloss.NewStyle().
				Foreground(ColorMintGreen).
				Bold(true)

	CursorTitleStyle = lipgloss.NewStyle().
				Foreground(ColorMintBlue).
				Bold(true)

	NormalDescStyle = lipgloss.NewStyle().
			Foreground(ColorDullSilver)

	VisualHighlightStyle = lipgloss.NewStyle().
				Background(ColorMutedBg).
				Foreground(ColorMintGreen)

	// Checkboxes
	CheckedBoxStyle = lipgloss.NewStyle().
			Foreground(ColorMintGreen).
			Bold(true)

	UncheckedBoxStyle = lipgloss.NewStyle().
				Foreground(ColorDullSilver)

	// Footer & Status Bar
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorDullSilver).
			Padding(0, 1)

	FooterVisualModeStyle = lipgloss.NewStyle().
				Background(ColorSeaGreen).
				Foreground(ColorWhite).
				Padding(0, 1).
				Bold(true)

	FooterReclaimedStyle = lipgloss.NewStyle().
				Foreground(ColorMintGreen).
				Bold(true)

	FooterSelectedCountStyle = lipgloss.NewStyle().
					Foreground(ColorBrightSlate).
					Bold(true)

	// Inspector Details
	InspectorHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorMintBlue).
				Bold(true).
				MarginBottom(1)

	InspectorKeyStyle = lipgloss.NewStyle().
				Foreground(ColorDullSilver).
				Width(14)

	InspectorValStyle = lipgloss.NewStyle().
				Foreground(ColorBrightSlate)

	InspectorCodeBlock = lipgloss.NewStyle().
				Foreground(ColorBrightSlate).
				Background(ColorMutedBg).
				Padding(1).
				MarginTop(1)

	// Confirmation Modal Overlay
	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorWarningRed).
			Background(ColorBase).
			Padding(1, 3).
			Width(64)

	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(ColorWarningRed).
			Bold(true).
			MarginBottom(1)

	ModalActionStyle = lipgloss.NewStyle().
				Foreground(ColorMintGreen).
				Bold(true)
)
