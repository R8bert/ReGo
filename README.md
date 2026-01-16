# ReGo ⚡

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Platform-Linux-orange" alt="Platform">
</p>

**Linux backup made simple** - Save your system configuration before reinstalling.

## ⚡ Quick Start

```bash
go build -o rego .
./rego
```

## 📦 Two Backup Modes

### ⚡ Quick Save (Recommended for most users)
Creates a tiny JSON file (~10KB) with:
- Flatpak app list
- RPM package list
- GNOME extensions
- GNOME settings (dconf)
- Repository list

**Output:** `~/rego-hostname.json`

### 💾 Full Save (For complete backups)
Creates a tar.gz archive with everything:
- All of the above, plus:
- **Dotfiles** (.bashrc, .zshrc, .gitconfig, .vimrc, etc.)
- **User fonts** (~/.local/share/fonts)
- **SSH config** (~/.ssh/config - no keys for security)
- **Autostart apps** (~/.config/autostart)
- **Wallpapers** (custom backgrounds)
- **GTK themes & icons** (~/.themes, ~/.icons)

**Output:** `~/rego-full-hostname-date.tar.gz`

## 🎹 Controls

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Space` | Toggle checkbox |
| `a` | Select/deselect all |
| `Enter` | Confirm |
| `Esc` | Back |
| `q` | Quit |

## 🔄 Workflow

### Before Reinstalling
1. Run `./rego`
2. Choose **Quick Save** or **Full Save**
3. Select what to include
4. Copy the file to USB/cloud

### After Fresh Install
1. Copy your backup file to new system
2. Run `./rego`
3. Select **Load Backup**

## 📋 Requirements

- Go 1.21+
- Linux (Fedora/GNOME recommended)
- Optional: `flatpak`, `dnf`, `dconf`, `gnome-extensions`

## 📄 License

MIT
