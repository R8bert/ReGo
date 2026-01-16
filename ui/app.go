package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/views"
)

type View int

const (
	ViewMainMenu View = iota
	ViewExport
	ViewBackup
	ViewRestore
	ViewImport
	ViewAbout
)

type Model struct {
	currentView View
	mainMenu    views.MainMenuView
	export      views.ExportView
	backup      views.BackupView
	restore     views.RestoreView
	importView  views.ImportView
	about       views.AboutView
	width       int
	height      int
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
		case "export":
			m.export = views.NewExportView()
			m.currentView = ViewExport
		case "backup":
			m.backup = views.NewBackupView()
			m.currentView = ViewBackup
		case "restore":
			m.restore = views.NewRestoreView()
			m.currentView = ViewRestore
		case "import":
			m.importView = views.NewImportView()
			m.currentView = ViewImport
		case "about":
			m.currentView = ViewAbout
		case "quit":
			return m, tea.Quit
		}
	case ViewExport:
		m.export, cmd, nav = m.export.Update(msg)
		if nav == "back" {
			m.currentView = ViewMainMenu
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
	case ViewImport:
		m.importView, cmd, nav = m.importView.Update(msg)
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
	case ViewExport:
		return m.export.View()
	case ViewBackup:
		return m.backup.View()
	case ViewRestore:
		return m.restore.View()
	case ViewImport:
		return m.importView.View()
	case ViewAbout:
		return m.about.View()
	default:
		return m.mainMenu.View()
	}
}
