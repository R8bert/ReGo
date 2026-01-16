# ReGo 🔄

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Platform-Linux-orange" alt="Platform">
</p>

A beautiful TUI application that helps you seamlessly backup and restore your Linux system configuration during reinstallation.

![ReGo Screenshot](docs/screenshot.png)

## ✨ Features

- **📦 Flatpak Backup** - Backup all Flatpak applications and remotes
- **📦 RPM Packages** - Backup user-installed RPM packages (DNF)
- **📁 Repositories** - Backup third-party DNF/YUM repositories
- **🧩 GNOME Extensions** - Backup extensions and their settings
- **⚙️ GNOME Settings** - Full dconf database backup
- **📄 Dotfiles** - Shell configs, git settings, SSH config, and more
- **🔤 User Fonts** - Backup custom fonts from ~/.local/share/fonts
- **🔒 Dry-Run Mode** - Preview changes before restoring
- **📋 Selective Restore** - Choose exactly what to restore

## 🚀 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/r8bert/rego.git
cd rego

# Build
go build -o rego .

# Run
./rego
```

### Requirements

- Go 1.21 or later
- Linux with GNOME (for GNOME-specific features)
- `flatpak` (for Flatpak backup)
- `dnf` (for RPM/repo backup)
- `dconf` (for GNOME settings backup)

## 📖 Usage

### Running ReGo

```bash
./rego
```

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate up/down |
| `Enter` | Select/confirm |
| `Space` | Toggle checkbox |
| `a` | Select/deselect all |
| `d` | Toggle dry-run mode |
| `Esc` | Go back |
| `q` | Quit |

### Creating a Backup

1. Select **Backup System** from the main menu
2. Choose which components to backup (all selected by default)
3. Press `Enter` to confirm
4. Wait for the backup to complete

Backups are saved to: `~/.config/rego/backups/[timestamp]/`

### Restoring from Backup

1. Select **Restore System** from the main menu
2. Choose a backup to restore from
3. Select which components to restore
4. Press `d` to toggle **dry-run mode** (recommended for first run)
5. Confirm and start the restore

> ⚠️ **Tip**: Always use dry-run mode first to see what will be changed!

## 📁 Project Structure

```
rego/
├── main.go                 # Application entry point
├── internal/
│   ├── backup/             # Backup modules
│   │   ├── flatpak.go      # Flatpak apps & remotes
│   │   ├── rpm.go          # RPM packages
│   │   ├── repos.go        # DNF repositories
│   │   ├── gnome_extensions.go
│   │   ├── gnome_settings.go
│   │   ├── dotfiles.go
│   │   ├── fonts.go
│   │   └── manager.go      # Backup orchestrator
│   ├── restore/            # Restore modules (mirrors backup)
│   └── utils/              # Shared utilities
├── ui/
│   ├── app.go              # Main TUI model
│   ├── styles/             # Lip Gloss styling
│   ├── components/         # Reusable UI components
│   └── views/              # Screen views
└── profiles/               # Saved backup profiles
```

## 🔧 Backup Data

Each backup creates a directory with:

| File | Contents |
|------|----------|
| `manifest.json` | Backup metadata and summary |
| `flatpak.json` | Installed Flatpak apps and remotes |
| `rpm_packages.json` | User-installed RPM packages |
| `repos.json` + `repos.d/` | Third-party repository files |
| `gnome_extensions.json` | Extension list and settings |
| `gnome_settings.dconf` | Full dconf database dump |
| `dotfiles/` | Copied dotfiles preserving structure |
| `fonts/` | User fonts directory copy |

## 🎨 Customization

### Default Dotfiles

Edit `internal/backup/types.go` to customize which dotfiles are backed up:

```go
func DefaultDotfiles() []string {
    return []string{
        ".bashrc",
        ".zshrc",
        ".gitconfig",
        // Add your own...
    }
}
```

## 🤝 Contributing

Contributions are welcome! Feel free to:

- Report bugs
- Suggest features
- Submit pull requests

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
