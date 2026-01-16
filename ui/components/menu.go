package components

import (
	"strings"

	"github.com/r8bert/rego/ui/styles"
)

type MenuItem struct {
	ID          string
	Title       string
	Description string
	Icon        string
}

type Menu struct {
	items  []MenuItem
	cursor int
	width  int
}

func NewMenu(items []MenuItem) *Menu {
	return &Menu{items: items, width: 50}
}

func (m *Menu) SetWidth(w int) { m.width = w }
func (m *Menu) Up() {
	if m.cursor > 0 {
		m.cursor--
	}
}
func (m *Menu) Down() {
	if m.cursor < len(m.items)-1 {
		m.cursor++
	}
}
func (m *Menu) Selected() MenuItem   { return m.items[m.cursor] }
func (m *Menu) SelectedIndex() int   { return m.cursor }
func (m *Menu) GetItems() []MenuItem { return m.items }
func (m *Menu) GetCursor() int       { return m.cursor }

func (m *Menu) View() string {
	var b strings.Builder
	for i, item := range m.items {
		cursor := "  "
		style := styles.NormalStyle
		if i == m.cursor {
			cursor = styles.SelectedStyle.Render("▸ ")
			style = styles.SelectedStyle
		}
		icon := item.Icon
		if icon == "" {
			icon = "•"
		}
		line := cursor + style.Render(icon+" "+item.Title)
		if item.Description != "" && i == m.cursor {
			line += "\n    " + styles.DimStyle.Render(item.Description)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}
