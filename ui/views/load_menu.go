package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/components"
	"github.com/r8bert/rego/ui/styles"
)

// LoadMenuView lets user choose between Quick Save and Full Save restore
type LoadMenuView struct {
	menu  *components.Menu
	frame int
}

func NewLoadMenuView() LoadMenuView {
	items := []components.MenuItem{
		{ID: "quick", Title: "⚡ Quick Save", Description: "Restore from .json file (package lists)"},
		{ID: "full", Title: "💾 Full Save", Description: "Restore from .tar.gz archive (with files)"},
	}
	return LoadMenuView{menu: components.NewMenu(items)}
}

func (v LoadMenuView) Init() tea.Cmd { return components.Tick() }

func (v LoadMenuView) Update(msg tea.Msg) (LoadMenuView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case components.TickMsg:
		v.frame++
		return v, components.Tick(), ""
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			v.menu.Up()
		case "down", "j":
			v.menu.Down()
		case "enter":
			sel := v.menu.Selected()
			if sel.ID == "quick" {
				return v, nil, "load_quick"
			} else if sel.ID == "full" {
				return v, nil, "load_full"
			}
		case "esc", "q":
			return v, nil, "back"
		}
	}
	return v, nil, ""
}

func (v LoadMenuView) View() string {
	s := styles.TitleStyle.Render("📥 Load Backup") + "\n\n"
	s += styles.DescriptionStyle.Render("Choose backup type to restore:") + "\n\n"
	s += v.menu.View() + "\n"
	s += styles.FooterStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")
	return s
}
