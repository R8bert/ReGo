package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/views"
)

type View int

const (
	ViewMainMenu View = iota
	ViewBackup
	ViewRestore
	ViewProfiles
	ViewSettings
	ViewAbout
)

type Model struct {
	currentView View
	mainMenu    views.MainMenuView
	backup      views.BackupView
	restore     views.RestoreView
	profiles    views.ProfilesView
	settings    views.SettingsView
	about       views.AboutView
	width       int
	height      int
}

func NewModel() Model {
	return Model{
		currentView: ViewMainMenu,
		mainMenu:    views.NewMainMenuView(),
		about:       views.NewAboutView(),
		settings:    views.NewSettingsView(),
		profiles:    views.NewProfilesView(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
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
		case "backup":
			m.backup = views.NewBackupView()
			m.currentView = ViewBackup
		case "restore":
			m.restore = views.NewRestoreView()
			m.currentView = ViewRestore
		case "profiles":
			m.currentView = ViewProfiles
		case "settings":
			m.currentView = ViewSettings
		case "about":
			m.currentView = ViewAbout
		case "quit":
			return m, tea.Quit
		}
	case ViewBackup:
		m.backup, cmd, nav = m.backup.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	case ViewRestore:
		m.restore, cmd, nav = m.restore.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	case ViewProfiles:
		m.profiles, cmd, nav = m.profiles.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
		}
	case ViewSettings:
		m.settings, cmd, nav = m.settings.Update(msg)
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
	case ViewBackup:
		return m.backup.View()
	case ViewRestore:
		return m.restore.View()
	case ViewProfiles:
		return m.profiles.View()
	case ViewSettings:
		return m.settings.View()
	case ViewAbout:
		return m.about.View()
	default:
		return m.mainMenu.View()
	}
}
