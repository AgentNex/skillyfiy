package config

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings available across the Skillyfiy TUI.
type KeyMap struct {
	CursorUp     key.Binding
	CursorDown   key.Binding
	ToggleSelect key.Binding
	VisualRange  key.Binding
	SelectAll    key.Binding
	FocusSearch  key.Binding
	SwitchPane   key.Binding
	CycleSort    key.Binding
	ConfirmPurge key.Binding
	Quit         key.Binding
	ForceQuit    key.Binding
}

// DefaultKeyMap provides ergonomic defaults for terminal and mobile/Termux usage.
var DefaultKeyMap = KeyMap{
	CursorUp:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	CursorDown:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	ToggleSelect: key.NewBinding(key.WithKeys(" ", "x"), key.WithHelp("space/x", "toggle")),
	VisualRange:  key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "visual mode")),
	SelectAll:    key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "select all")),
	FocusSearch:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	SwitchPane:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch pane")),
	CycleSort:    key.NewBinding(key.WithKeys("f1", "ctrl+s"), key.WithHelp("ctrl+s", "cycle sort")),
	ConfirmPurge: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "purge")),
	Quit:         key.NewBinding(key.WithKeys("ctrl+q", "esc"), key.WithHelp("ctrl+q/esc", "exit")),
	ForceQuit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
}
