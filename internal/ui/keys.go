package ui

import key "charm.land/bubbles/v2/key"

type keyMap struct {
	Quit      key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Space     key.Binding
	New       key.Binding
	Delete    key.Binding
	MoveDown  key.Binding
	MoveUp    key.Binding
	MarkDone  key.Binding
	Backspace key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "switch back"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "stop editing"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "previous item"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "next item"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "cursor left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "cursor right"),
		),
		Space: key.NewBinding(
			key.WithKeys("space", " "),
			key.WithHelp("space", "start/pause"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new item"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete item"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "move item down"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("=", "+"),
			key.WithHelp("=/+", "move item up"),
		),
		MarkDone: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "toggle done"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "delete char"),
		),
	}
}
