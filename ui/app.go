package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r8bert/rego/ui/views"
)

type View int

const (
	ViewMainMenu View = iota
	ViewQuickSave
	ViewFullSave
	ViewLoad
	ViewAbout
)

type Model struct {
	currentView View
	width       int
	height      int
	mainMenu    views.MainMenuView
	quickSave   views.LightBackupView
	fullSave    views.FullSaveView
	load        views.LightRestoreView
	about       views.AboutView
}

func NewModel() Model {
	return Model{
		currentView: ViewMainMenu,
		width:       80,
		height:      24,
		mainMenu:    views.NewMainMenuView(),
		about:       views.NewAboutView(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.mainMenu.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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
		case "quick":
			m.quickSave = views.NewLightBackupView()
			m.currentView = ViewQuickSave
		case "full":
			m.fullSave = views.NewFullSaveView()
			m.currentView = ViewFullSave
		case "load":
			m.load = views.NewLightRestoreView()
			m.currentView = ViewLoad
		case "about":
			m.currentView = ViewAbout
		case "quit":
			return m, tea.Quit
		}
	case ViewQuickSave:
		m.quickSave, cmd, nav = m.quickSave.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	case ViewFullSave:
		m.fullSave, cmd, nav = m.fullSave.Update(msg)
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
	var content string

	switch m.currentView {
	case ViewQuickSave:
		content = m.quickSave.View()
	case ViewFullSave:
		content = m.fullSave.View()
	case ViewLoad:
		content = m.load.View()
	case ViewAbout:
		content = m.about.View()
	default:
		content = m.mainMenu.View()
	}

	// Center content in the window
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content)
}
