# ReGo ⚡

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Platform-Linux-orange" alt="Platform">
</p>

**Super simple Linux backup** - Save your entire system configuration to a single tiny file.

## ⚡ Quick Start

```bash
# Build
go build -o rego .

# Run
./rego

# Select "Quick Save" → Done!
```

Your backup is saved to: `~/rego-hostname.json` (usually just a few KB!)

## 📦 What Gets Saved

| Component | Saved As |
|-----------|----------|
| Flatpak apps | List of app IDs |
| RPM packages | List of package names |
| GNOME extensions | List of extension UUIDs |
| GNOME settings | dconf database dump |
| Repositories | List of third-party repos |

**No files are copied** - just the data needed to reinstall everything.

## 🔄 Workflow

### Before Reinstalling

1. Run `./rego`
2. Press Enter on "Quick Save"
3. Copy `~/rego-hostname.json` to USB/cloud/email

### After Fresh Install

1. Copy `rego-hostname.json` to new system
2. Build and run `./rego`
3. Select "Load Backup"
4. Choose what to restore

## 🎹 Controls

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Enter` | Select |
| `d` | Toggle dry-run |
| `Esc` | Back |
| `q` | Quit |

## 📄 Example Backup File

```json
{
  "version": "1.0",
  "hostname": "my-laptop",
  "flatpaks": ["org.mozilla.firefox", "com.spotify.Client"],
  "rpm_packages": ["vim", "htop", "nodejs"],
  "gnome_extensions": ["dash-to-dock@micxgx.gmail.com"],
  "dconf_settings": "[org/gnome/desktop/interface]\ncolor-scheme='prefer-dark'"
}
```

## 📋 Requirements

- Go 1.21+
- Linux with GNOME (optional - for GNOME features)
- `flatpak`, `dnf`, `dconf` (optional - only used if available)

## 📄 License

MIT
