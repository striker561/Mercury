# Mercury

<p align="center">
  <img src="frontend/public/mercury-logo.png" alt="Mercury logo" width="200"/>
</p>

> _"I carry messages. Yours, specifically. Across your LAN. Without the clouds, without the drama, and definitely without your data ending up in a database you didn't sign up for."_

Mercury is a cross-platform LAN clipboard and file sharing tray app. It lives in your system tray and syncs whatever you copy on one machine to another — provided they share the same passphrase and are on the same network.

No cloud. No accounts. No history. **It just works.**

## Features

- **Clipboard sync** — Copy text or images on one machine. Paste on another.
- **File transfer** — Send files between peers with accept/decline flow and progress.
- **System tray** — Stays out of the way until you need it. No dock icon on macOS.
- **Encrypted** — AES-256-GCM over the LAN. Passphrase never leaves your devices.

### What it is NOT

- **Not a clipboard manager.** We're a courier, not a hoarder — no history, search, or pinning.
- **Not cloud-based.** LAN only. No accounts. No subscription.
- **Not cross-internet.** If you can't reach each other on the LAN, we can't help.

## Why Mercury?

|                   | Mercury               | Klip                               | Clipp                             |
| ----------------- | --------------------- | ---------------------------------- | --------------------------------- |
| Open source       | Yes                   | Partial (free tier limits devices) | Yes                               |
| Device limit      | None                  | 2 free                             | None                              |
| Clipboard history | No (by design)        | No                                 | Yes                               |
| Trust model       | Shared passphrase     | TLS certificates                   | Group passphrase                  |
| Cross-platform    | macOS, Linux, Windows | macOS, Linux, Windows              | macOS, Windows, iOS (+ Linux CLI) |

Mercury fills the gap for people who want **simple, unlimited, LAN-only sync** without clipboard hoarding or freemium device caps.

## Tech Stack

| Layer      | Choice               |
| ---------- | -------------------- |
| Framework  | Wails v3             |
| Backend    | Go                   |
| Frontend   | React + TypeScript   |
| Bundler    | Bun + Vite           |
| Encryption | AES-256-GCM + PBKDF2 |
| Discovery  | mDNS (zeroconf)      |

## Getting Started

### Prerequisites

- Go 1.25+
- Bun
- Wails v3 CLI
- **Linux:** `libgtk-4-dev`, `libwebkitgtk-6.0-dev`
- **macOS:** Xcode
- **Windows:** WebView2

### Install Wails v3

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Run in Dev Mode

```bash
GOTOOLCHAIN=go1.25.12 wails3 dev
```

### Build for Production

```bash
GOTOOLCHAIN=go1.25.12 go build -o mercury .
```

## Usage

### Tray

- **Right click** — Open Mercury, view peer status, pause/resume sync, quit.
- **Tray icon** — Active state during transfers or recent clipboard sync.

### Window

- **Home** — Connection status, connected devices, incoming file offers.
- **Settings** — Passphrase, sync toggle, file preferences, start on login.
- **Close (×)** — Hides to tray. The app keeps running.

### Settings

| Section | What it does                                                     |
| ------- | ---------------------------------------------------------------- |
| Sync    | Passphrase (shared across your devices), sync on/off             |
| Files   | Save folder (default `~/Downloads/Mercury/`), accept/auto-accept |
| General | Start on login                                                   |

## Security Model

1. **Passphrase** — Never transmitted. Used locally to derive the encryption key.
2. **PBKDF2** — 100,000 iterations of SHA-256 at startup.
3. **AES-256-GCM** — Authenticated encryption. Wrong key → silent drop.
4. **mDNS** — Announces presence only. No secrets on the wire.
5. **25MB max** for clipboard sync payloads.

## Project Structure

```
mercury/
├── main.go              # Entry point
├── app/
│   ├── main.go          # Wails bootstrap, tray, window
│   ├── app.go           # MercuryApp IPC bindings
│   ├── dashboard.go     # Batched dashboard state for UI
│   ├── backend/         # sync, clipboard, transfer, crypto, storage
│   └── system/          # Tray menu
├── frontend/
│   ├── public/          # CSS, fonts, logos
│   └── src/             # React UI
└── build/               # Platform packaging
```

## Roadmap

See [TODO.md](TODO.md). v0.1.0 is feature-complete for real-world use — further work is feedback-driven.

### Linux packages

`.deb` / `.rpm` packages declare runtime libraries in `depends:` (GTK4, WebKitGTK 6). **apt/dnf install those automatically** when you install Mercury. The `preinstall.sh` script does not install packages — it only warns if Mercury is already running during an upgrade. Optional GNOME tray support is listed as `recommends: libayatana-appindicator3-1`.

## Contributing

Issues and PRs welcome. Keep changes focused — this is a small utility, not a platform.

## License

[MIT](LICENSE)
