package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Add      key.Binding
	Presets  key.Binding
	Export   key.Binding
	Remove   key.Binding
	Help     key.Binding
	Quit     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Up       key.Binding
	Down     key.Binding
	Save     key.Binding
	Delete   key.Binding
	Tab      key.Binding
	Format   key.Binding
	ClearAll key.Binding
}

var keys = keyMap{
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add files"),
	),
	Presets: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "presets"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	Remove: key.NewBinding(
		key.WithKeys("d", "x"),
		key.WithHelp("d/x", "remove"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "left"),
		key.WithHelp("esc/←", "back"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Save: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "save preset"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle dir/file"),
	),
	Format: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "toggle format"),
	),
	ClearAll: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "clear all"),
	),
}
