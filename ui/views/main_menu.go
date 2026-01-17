package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r8bert/rego/internal/backup"
	"github.com/r8bert/rego/ui/components"
	"github.com/r8bert/rego/ui/styles"
)

type MainMenuView struct {
	cursor int
	items  []menuItem
	frame  int
}

type menuItem struct {
	id    string
	icon  string
	title string
	desc  string
}

func NewMainMenuView() MainMenuView {
	return MainMenuView{
		items: []menuItem{
			{id: "quick", icon: ">", title: "Quick Save", desc: "Light backup - package lists only"},
			{id: "full", icon: ">", title: "Full Save", desc: "Complete backup with all files"},
			{id: "load", icon: ">", title: "Load Backup", desc: "Restore from a saved backup"},
			{id: "about", icon: ">", title: "About", desc: "About ReGo"},
			{id: "quit", icon: ">", title: "Quit", desc: "Exit application"},
		},
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
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			return m, nil, m.items[m.cursor].id
		case "q", "ctrl+c":
			return m, tea.Quit, "quit"
		}
	}
	return m, nil, ""
}

func (m MainMenuView) View() string {
	container := lipgloss.NewStyle().Padding(1, 3)

	// Logo
	logo := styles.RenderLogo()

	// Tagline
	tagline := lipgloss.NewStyle().
		Width(styles.AppWidth).
		Align(lipgloss.Center).
		Render(styles.RenderTagline())

	// System info box
	sysInfo := styles.GetSystemInfo()
	desktopInfo := styles.GetDesktopSuggestion()
	pkgMgr := backup.GetPackageManagerName()

	infoContent := styles.DimStyle.Render("System: ") + styles.NormalStyle.Render(sysInfo) + "\n" +
		styles.DimStyle.Render("Package Manager: ") + styles.NormalStyle.Render(pkgMgr) + "\n" +
		styles.DimStyle.Render("Desktop: ") + desktopInfo

	infoBox := styles.CardStyle.Width(styles.AppWidth - 6).Render(infoContent)

	// Menu items
	var menuContent string
	for i, item := range m.items {
		selected := i == m.cursor
		menuContent += styles.RenderMenuItem(item.icon, item.title, item.desc, selected, m.frame)
		if i < len(m.items)-1 {
			menuContent += "\n"
		}
	}

	// Menu card
	menuCard := styles.CardStyle.Width(styles.AppWidth - 6).Render(menuContent)

	// Footer
	footer := lipgloss.NewStyle().
		Width(styles.AppWidth).
		Align(lipgloss.Center).
		Foreground(styles.Muted).
		MarginTop(1).
		Render("[up/down] Navigate  [enter] Select  [q] Quit")

	content := logo + tagline + "\n" + infoBox + "\n" + menuCard + footer

	return container.Render(content)
}
