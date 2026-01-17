# ReGo

<p align="center">
  <img src="imgs/logo.png" alt="ReGo Logo" width="400">
</p>

ReGo is a terminal-based Linux system backup and restore utility designed to help users preserve their system configuration before reinstalling their operating system.

## Overview

ReGo creates lightweight backup files containing lists of installed packages, desktop environment settings, and configuration files. These backups can be used to quickly restore a system to its previous state after a fresh installation.

## Features

### Package Manager Support

ReGo automatically detects and supports multiple package managers:

| Package Manager | Distribution |
|----------------|--------------|
| APT | Debian, Ubuntu, Linux Mint, Pop!_OS |
| DNF | Fedora, RHEL, CentOS, Rocky Linux |
| Pacman | Arch Linux, Manjaro |
| Zypper | openSUSE |

### Desktop Environment Support

| Desktop | Backed Up Components |
|---------|---------------------|
| GNOME | Extensions, dconf settings, keybindings |
| KDE Plasma | Plasma config, KWin settings, widgets, themes |

### Backup Types

#### Quick Save

Creates a lightweight JSON file containing:
- Installed Flatpak applications
- User-installed system packages (apt/dnf/pacman)
- GNOME extensions or KDE widgets
- Desktop settings (dconf dump)
- Third-party repository names

Output: `~/rego-[hostname].json` (typically 10-50 KB)

#### Full Save

Creates a compressed tar.gz archive containing:
- All Quick Save data
- Dotfiles (.bashrc, .zshrc, .gitconfig, .vimrc, etc.)
- User fonts (~/.local/share/fonts)
- SSH configuration (~/.ssh/config, not private keys)
- Autostart applications
- Custom wallpapers
- GTK themes and icons
- Konsole/terminal profiles

Output: `~/rego-full-[hostname]-[date].tar.gz`

## Installation

### Prerequisites

- Go 1.21 or later
- Linux operating system

### Building from Source

```bash
git clone https://github.com/r8bert/rego.git
cd rego
go build -o rego .
```

### Running

```bash
./rego
```

## Usage

### Keyboard Controls

| Key | Action |
|-----|--------|
| Up/Down or j/k | Navigate menu |
| Space | Toggle checkbox |
| a | Select/deselect all |
| Enter | Confirm selection |
| Escape | Go back |
| q | Quit |

### Creating a Backup

1. Launch ReGo
2. Select "Quick Save" or "Full Save"
3. Review and toggle the components you want to backup
4. Press Enter to create the backup
5. Copy the output file to external storage

### Restoring a Backup

1. Copy your backup file to the new system
2. Launch ReGo
3. Select "Load Backup"
4. Select the backup file
5. Choose dry-run mode to preview changes
6. Confirm to restore

## Project Structure

```
rego/
├── main.go                 # Application entry point
├── internal/
│   ├── backup/             # Backup modules
│   │   ├── light.go        # Quick Save implementation
│   │   ├── full.go         # Full Save implementation
│   │   ├── system.go       # Package manager detection
│   │   ├── flatpak.go      # Flatpak backup
│   │   ├── kde.go          # KDE Plasma backup
│   │   └── ...
│   ├── restore/            # Restore modules
│   └── utils/              # Utility functions
└── ui/
    ├── app.go              # Main TUI application
    ├── views/              # TUI views
    ├── components/         # Reusable TUI components
    └── styles/             # Visual styling
```

## Backup File Format

### Quick Save (JSON)

```json
{
  "version": "1.0",
  "created_at": "2024-01-17T12:00:00Z",
  "hostname": "workstation",
  "distro": "Fedora Linux 39",
  "flatpaks": [
    "org.mozilla.firefox",
    "com.spotify.Client"
  ],
  "rpm_packages": [
    "vim",
    "htop",
    "nodejs"
  ],
  "gnome_extensions": [
    "dash-to-dock@micxgx.gmail.com"
  ],
  "dconf_settings": "[org/gnome/desktop/interface]\ncolor-scheme='prefer-dark'",
  "repos": [
    "rpmfusion-free",
    "rpmfusion-nonfree"
  ]
}
```

## Configuration

ReGo stores its configuration and temporary files in the user's home directory. No system-level configuration is required.

## Security Considerations

- SSH private keys are never backed up
- Backup files may contain sensitive configuration data
- Store backup files securely
- Review backup contents before sharing

## Requirements

### Runtime Dependencies (Optional)

- `flatpak` - For Flatpak application backup
- `dconf` - For GNOME settings backup
- `gnome-extensions` - For GNOME extension backup

## Contributing

Contributions are welcome. Please submit issues and pull requests to the GitHub repository.


