package components

import (
	"strings"

	"github.com/r8bert/rego/ui/styles"
)

type CheckboxItem struct {
	ID          string
	Title       string
	Description string
	Checked     bool
	Disabled    bool
}

type CheckboxList struct {
	items  []CheckboxItem
	cursor int
}

func NewCheckboxList(items []CheckboxItem) *CheckboxList {
	return &CheckboxList{items: items}
}

func (c *CheckboxList) Up() {
	if c.cursor > 0 {
		c.cursor--
	}
}
func (c *CheckboxList) Down() {
	if c.cursor < len(c.items)-1 {
		c.cursor++
	}
}

func (c *CheckboxList) Toggle() {
	if !c.items[c.cursor].Disabled {
		c.items[c.cursor].Checked = !c.items[c.cursor].Checked
	}
}

func (c *CheckboxList) ToggleAll() {
	allChecked := true
	for _, item := range c.items {
		if !item.Disabled && !item.Checked {
			allChecked = false
			break
		}
	}
	for i := range c.items {
		if !c.items[i].Disabled {
			c.items[i].Checked = !allChecked
		}
	}
}

func (c *CheckboxList) GetSelected() []CheckboxItem {
	var selected []CheckboxItem
	for _, item := range c.items {
		if item.Checked {
			selected = append(selected, item)
		}
	}
	return selected
}

func (c *CheckboxList) Items() []CheckboxItem { return c.items }

func (c *CheckboxList) View() string {
	var b strings.Builder
	for i, item := range c.items {
		cursor := "  "
		style := styles.NormalStyle
		if i == c.cursor {
			cursor = styles.SelectedStyle.Render("▸ ")
			style = styles.SelectedStyle
		}
		if item.Disabled {
			style = styles.DimStyle
		}

		checkbox := "[ ]"
		if item.Checked {
			checkbox = styles.SuccessStyle.Render("[✓]")
		}
		if item.Disabled {
			checkbox = styles.DimStyle.Render("[-]")
		}

		line := cursor + checkbox + " " + style.Render(item.Title)
		if item.Description != "" && i == c.cursor {
			line += "\n      " + styles.DimStyle.Render(item.Description)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}
