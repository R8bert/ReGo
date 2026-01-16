package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/styles"
	"github.com/r8bert/rego/ui/components"
)

type MainMenuView struct {
	menu   *components.Menu
	footer *components.Footer
}

func NewMainMenuView() MainMenuView {
	items := []components.MenuItem{
		{ID: "backup", Title: "Backup System", Description: "Create a backup of your current configuration", Icon: "💾"},
		{ID: "restore", Title: "Restore System", Description: "Restore from a previous backup", Icon: "🔄"},
		{ID: "profiles", Title: "Manage Profiles", Description: "View and manage saved backup profiles", Icon: "📁"},
		{ID: "settings", Title: "Settings", Description: "Configure backup options", Icon: "⚙️"},
		{ID: "about", Title: "About", Description: "About ReGo", Icon: "ℹ️"},
		{ID: "quit", Title: "Quit", Description: "Exit the application", Icon: "🚪"},
	}
	return MainMenuView{
		menu:   components.NewMenu(items),
		footer: components.NewFooter("↑/↓: Navigate • Enter: Select • q: Quit"),
	}
}

func (m MainMenuView) Init() tea.Cmd { return nil }

func (m MainMenuView) Update(msg tea.Msg) (MainMenuView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.menu.Up()
		case "down", "j":
			m.menu.Down()
		case "enter":
			return m, nil, m.menu.Selected().ID
		case "q", "ctrl+c":
			return m, tea.Quit, "quit"
		}
	}
	return m, nil, ""
}

func (m MainMenuView) View() string {
	s := styles.RenderLogo() + "\n"
	s += styles.RenderTagline() + "\n\n"
	s += styles.TitleStyle.Render("Main Menu") + "\n\n"
	s += m.menu.View() + "\n"
	s += m.footer.View()
	return s
}
