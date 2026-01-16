package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/views"
)

type View int

const (
	ViewMainMenu View = iota
	ViewSave
	ViewLoad
	ViewAbout
)

type Model struct {
	currentView View
	mainMenu    views.MainMenuView
	save        views.LightBackupView
	load        views.LightRestoreView
	about       views.AboutView
}

func NewModel() Model {
	return Model{
		currentView: ViewMainMenu,
		mainMenu:    views.NewMainMenuView(),
		about:       views.NewAboutView(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	var nav string

	switch m.currentView {
	case ViewMainMenu:
		m.mainMenu, cmd, nav = m.mainMenu.Update(msg)
		switch nav {
		case "save":
			m.save = views.NewLightBackupView()
			m.currentView = ViewSave
		case "load":
			m.load = views.NewLightRestoreView()
			m.currentView = ViewLoad
		case "about":
			m.currentView = ViewAbout
		case "quit":
			return m, tea.Quit
		}
	case ViewSave:
		m.save, cmd, nav = m.save.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	case ViewLoad:
		m.load, cmd, nav = m.load.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	case ViewAbout:
		m.about, cmd, nav = m.about.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.currentView {
	case ViewSave:
		return m.save.View()
	case ViewLoad:
		return m.load.View()
	case ViewAbout:
		return m.about.View()
	default:
		return m.mainMenu.View()
	}
}
