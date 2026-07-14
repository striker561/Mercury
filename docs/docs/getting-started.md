---
sidebar_position: 2
---

# Getting Started

## Download the latest release

**Release page:** [github.com/striker561/Mercury/releases](https://github.com/striker561/Mercury/releases)

| Platform                 | Download                                                                                                                  | Install                                       |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| macOS (Universal)        | [mercury-macos-universal.dmg](https://github.com/striker561/Mercury/releases/latest/download/mercury-macos-universal.dmg) | Open the DMG and drag Mercury to Applications |
| Linux (Pop!\_OS, Ubuntu) | [mercury_linux_amd64.deb](https://github.com/striker561/Mercury/releases/latest/download/mercury_linux_amd64.deb)         | `sudo dpkg -i mercury_linux_amd64.deb`        |
| Windows                  | [mercury-amd64-installer.exe](https://github.com/striker561/Mercury/releases/latest/download/mercury-amd64-installer.exe)             | Run the installer (no console window)         |

Verify files with [`SHA256SUMS`](https://github.com/striker561/Mercury/releases/latest/download/SHA256SUMS) from the same release.

Bundle ID: **`com.mercury.app`** (notifications and single-instance).

### Auto-updates

Installed builds include the [Wails v3 updater](https://v3.wails.io/guides/updater/). Use the tray menu **Check for Updates…** to download, verify (SHA256), and apply a newer release from GitHub. After ~10 minutes of runtime, Mercury may notify you when an update is available.

Manual installers (`.dmg`, `.deb`, NSIS) are for first-time setup. In-app updates use separate bare-binary / zip assets published on each release.

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

Production packages (set `VERSION` to match your tag, without the `v` prefix in plist/nfpm metadata):

```bash
wails3 task darwin:package:universal VERSION=v0.2.0
wails3 task linux:create:deb           VERSION=v0.2.0
wails3 task windows:package            VERSION=v0.2.0
```

Publish a release by pushing a semver tag, e.g. `git tag v0.2.0 && git push origin v0.2.0`. See [`.github/workflows/release.yml`](https://github.com/striker561/Mercury/blob/main/.github/workflows/release.yml).

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
