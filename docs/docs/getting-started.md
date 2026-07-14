---
sidebar_position: 2
---

# Getting Started

## Install from a release

Download the latest release for your platform from [GitHub Releases](https://github.com/striker561/Mercury/releases).

| Platform                 | Artifact                      | Install                                       |
| ------------------------ | ----------------------------- | --------------------------------------------- |
| macOS                    | `mercury-macos-universal.dmg` | Open the DMG and drag Mercury to Applications |
| Linux (Pop!\_OS, Ubuntu) | `mercury_*_amd64.deb`         | `sudo dpkg -i mercury_*_amd64.deb`            |
| Windows                  | `mercury-installer.exe`       | Run the installer (no console window)         |

Bundle ID: **`com.mercury.app`** (notifications and single-instance).

## Build from source

### Prerequisites

| Platform | Requirements                                                                              |
| -------- | ----------------------------------------------------------------------------------------- |
| All      | [Go 1.25+](https://go.dev/dl/), [Bun](https://bun.sh/), [Wails v3 CLI](https://wails.io/) |
| macOS    | Xcode Command Line Tools                                                                  |
| Linux    | `libgtk-4-dev`, `libwebkitgtk-6.0-dev`                                                    |
| Windows  | WebView2 (usually pre-installed)                                                          |

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
git clone https://github.com/striker561/Mercury.git
cd Mercury
GOTOOLCHAIN=go1.25.12 wails3 dev
```

Production packages:

```bash
wails3 task darwin:package:universal   # macOS .app + DMG
wails3 task linux:create:deb           # Linux .deb
wails3 task windows:package            # Windows NSIS installer
```

## First run

1. Launch Mercury and look for the tray icon (caduceus).
2. **Left-click** the tray icon to open the window (or right-click → Open Mercury).
3. Open **Settings**, enter a **passphrase**, and click **Save & start**.
4. Repeat on your other machines with the **same passphrase** on the **same LAN**.
5. On **Home**, wait for peers to appear.

Copy text on one machine. Paste on the other.

## Troubleshooting discovery

If no peers appear after ~30 seconds:

- Confirm all machines share the **same subnet** and passphrase (case-sensitive).
- Allow TCP port **47821** through the firewall.
- On **Pop!\_OS / GNOME**, install [AppIndicator Support](https://extensions.gnome.org/extension/615/appindicator-support/) if the tray icon is missing.

See [FAQ](./faq) for VPN and passphrase mismatch notes.
