package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/components"
	"github.com/r8bert/rego/ui/styles"
)

type MainMenuView struct {
	menu   *components.Menu
	footer *components.Footer
	frame  int
}

func NewMainMenuView() MainMenuView {
	items := []components.MenuItem{
		{ID: "quick", Title: "Quick Save", Description: "Light backup - just package lists (few KB)", Icon: "⚡"},
		{ID: "full", Title: "Full Save", Description: "Complete backup with fonts, dotfiles, themes", Icon: "💾"},
		{ID: "load", Title: "Load Backup", Description: "Restore from a backup file", Icon: "📥"},
		{ID: "about", Title: "About", Description: "About ReGo", Icon: "ℹ️"},
		{ID: "quit", Title: "Quit", Description: "Exit", Icon: "🚪"},
	}
	return MainMenuView{
		menu:   components.NewMenu(items),
		footer: components.NewFooter("↑/↓: Navigate • Enter: Select • q: Quit"),
	}
}

func (m MainMenuView) Init() tea.Cmd { return components.Tick() }

func (m MainMenuView) Update(msg tea.Msg) (MainMenuView, tea.Cmd, string) {
	switch msg := msg.(type) {
	case components.TickMsg:
		m.frame++
		return m, components.Tick(), ""
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

// Animated logo frames
var logoColors = []string{"#7C3AED", "#8B5CF6", "#A78BFA", "#8B5CF6"}

func (m MainMenuView) View() string {
	// Animated logo with subtle pulse effect
	frame := m.frame / 4 % 4
	_ = frame // Could be used for color cycling

	s := styles.RenderLogo() + "\n"

	// Animated tagline
	taglines := []string{
		"Linux Reinstall Helper",
		"Linux Reinstall Helper ✨",
		"Linux Reinstall Helper",
		"Linux Reinstall Helper ⚡",
	}
	s += styles.SubtitleStyle.Render(taglines[m.frame/10%len(taglines)]) + "\n\n"

	// Menu with animated selection indicator
	for i, item := range m.menu.GetItems() {
		if i == m.menu.GetCursor() {
			// Animated cursor
			cursors := []string{"▸ ", "► ", "▹ ", "► "}
			cursor := cursors[m.frame/3%len(cursors)]
			line := cursor + item.Icon + " " + styles.SelectedStyle.Render(item.Title)
			s += line + "\n"
			s += "    " + styles.DescriptionStyle.Render(item.Description) + "\n"
		} else {
			s += "  " + item.Icon + " " + styles.NormalStyle.Render(item.Title) + "\n"
		}
	}

	s += "\n" + m.footer.View()
	return s
}
