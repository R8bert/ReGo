package restore

import (
	"github.com/r8bert/rego/internal/backup"
)

type Manager struct {
	restorers  map[RestoreType]Restorer
	backupPath string
}

func NewManager() *Manager {
	m := &Manager{restorers: make(map[RestoreType]Restorer)}
	m.RegisterRestorer(NewFlatpakRestore())
	m.RegisterRestorer(NewRPMRestore())
	m.RegisterRestorer(NewReposRestore())
	m.RegisterRestorer(NewGnomeExtensionsRestore())
	m.RegisterRestorer(NewGnomeSettingsRestore())
	m.RegisterRestorer(NewDotfilesRestore())
	m.RegisterRestorer(NewFontsRestore())
	return m
}

func (m *Manager) RegisterRestorer(r Restorer)                { m.restorers[r.Type()] = r }
func (m *Manager) GetRestorer(t RestoreType) (Restorer, bool) { r, ok := m.restorers[t]; return r, ok }

func (m *Manager) GetAvailableRestorers() []Restorer {
	var available []Restorer
	for _, t := range AllRestoreTypes() {
		if r, ok := m.restorers[t]; ok && r.Available() {
			available = append(available, r)
		}
	}
	return available
}

type RestoreProgress struct {
	CurrentType RestoreType
	CurrentName string
	TotalSteps  int
	CurrentStep int
	Completed   []RestoreResult
	InProgress  bool
	Error       error
}

type ProgressCallback func(progress RestoreProgress)

func (m *Manager) RunRestore(opts RestoreOptions, callback ProgressCallback) ([]RestoreResult, error) {
	m.backupPath = opts.BackupPath

	var typesToRestore []RestoreType
	if opts.IncludeFlatpak {
		typesToRestore = append(typesToRestore, RestoreTypeFlatpak)
	}
	if opts.IncludeRPM {
		typesToRestore = append(typesToRestore, RestoreTypeRPM)
	}
	if opts.IncludeRepos {
		typesToRestore = append(typesToRestore, RestoreTypeRepos)
	}
	if opts.IncludeGnomeExtensions {
		typesToRestore = append(typesToRestore, RestoreTypeGnomeExtensions)
	}
	if opts.IncludeGnomeSettings {
		typesToRestore = append(typesToRestore, RestoreTypeGnomeSettings)
	}
	if opts.IncludeDotfiles {
		typesToRestore = append(typesToRestore, RestoreTypeDotfiles)
	}
	if opts.IncludeFonts {
		typesToRestore = append(typesToRestore, RestoreTypeFonts)
	}

	if dfRestore, ok := m.restorers[RestoreTypeDotfiles].(*DotfilesRestore); ok {
		dfRestore.SetMerge(opts.MergeDotfiles)
	}

	progress := RestoreProgress{TotalSteps: len(typesToRestore), InProgress: true}
	var results []RestoreResult

	for i, restoreType := range typesToRestore {
		restorer, ok := m.restorers[restoreType]
		if !ok || !restorer.Available() {
			continue
		}

		progress.CurrentStep = i + 1
		progress.CurrentType = restoreType
		progress.CurrentName = restorer.Name()
		if callback != nil {
			callback(progress)
		}

		result, _ := restorer.Restore(opts.BackupPath, opts.DryRun)
		results = append(results, result)
		progress.Completed = append(progress.Completed, result)
	}

	progress.InProgress = false
	if callback != nil {
		callback(progress)
	}
	return results, nil
}

func (m *Manager) PreviewRestore(backupPath string) map[RestoreType][]string {
	preview := make(map[RestoreType][]string)
	for t, r := range m.restorers {
		if items, err := r.Preview(backupPath); err == nil {
			preview[t] = items
		}
	}
	return preview
}

func (m *Manager) LoadBackup(backupPath string) (*backup.BackupManifest, error) {
	mgr := backup.NewManager()
	return mgr.LoadBackup(backupPath)
}
