package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Select    key.Binding
	Toggle    key.Binding
	Back      key.Binding
	Quit      key.Binding
	Help      key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
	SelectAll key.Binding
}

var Keys = KeyMap{
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Left:      key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
	Right:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
	Select:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Toggle:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	Back:      key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Confirm:   key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
	Cancel:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
	SelectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Toggle, k.SelectAll},
		{k.Back, k.Quit, k.Help},
	}
}
