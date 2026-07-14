# Mercury

<p align="center">
  <img src="frontend/public/mercury-logo.png" alt="Mercury logo" width="200"/>
</p>

> _"I carry messages. Yours, specifically. Across your LAN. Without the clouds, without the drama, and definitely without your data ending up in a database you didn't sign up for."_

Mercury is a cross-platform LAN clipboard and file sharing tray app. It lives in your system tray, judges your networking setup silently, and ensures whatever you copy on one machine appears on another — provided they share the same passphrase and are on the same network.

No cloud. No accounts. No history. **It just works.** Like a messenger god, but for your clipboard.

## Features

- **Clipboard sync** — Copy text or images on one machine. Paste on another. It's that simple.
- **File transfer** — Drag a file onto the tray, your peer gets asked if they want it. They can say yes. They can say no. We judge either way.
- **System tray** — Sits in your tray like a smug little orb. No dock icon. No taskbar presence. You'll forget it exists until you need it. That's the point.

### What it is NOT

- **Not a clipboard manager.** We don't store history. We don't search. We don't pin. We're a courier, not a hoarder.
- **Not cloud-based.** We don't know what "the cloud" is. We use the LAN. It's older, wiser, and doesn't require a monthly subscription.
- **Not cross-internet.** If you can't ping each other, we can't help you. Get closer.

## Tech Stack

| Layer      | Choice               | Why                                                                          |
| ---------- | -------------------- | ---------------------------------------------------------------------------- |
| Framework  | Wails v3             | Because writing web views in bare C++ is barbaric                            |
| Backend    | Go                   | Because I wanted to learn it, and it's fast enough to judge you in real-time |
| Frontend   | React + TypeScript   | Because you have to suffer somewhere                                         |
| Bundler    | Bun                  | It's fast. It's trendy. I like it.                                           |
| Encryption | AES-256-GCM + PBKDF2 | Your cat pictures are safe                                                   |
| Discovery  | mDNS (zeroconf)      | No central server. No config. It just works.                                 |

## Getting Started

### Prerequisites

- Go 1.25+ (we use the toolchain, you heathen)
- Bun (for the frontend)
- Wails v3 CLI
- **Linux:** `libgtk-4-dev`, `libwebkitgtk-6.0-dev`
- **macOS:** Xcode. You know what to do.
- **Windows:** WebView2. It's already there. Probably.

### Install Wails v3

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Install System Dependencies (Linux)

```bash
sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev
```

### Run in Dev Mode

```bash
cd mercury
GOTOOLCHAIN=go1.25.12 wails3 dev
```

The tray appears. The frontend hot-reloads. You change code. It updates. You feel like a god. This is intentional.

### Build for Production

```bash
GOTOOLCHAIN=go1.25.12 go build -o mercury .
```

One binary. Zero dependencies. Divine.

## Project Structure

## Brand Assets

The Mercury logo is available in two variants:

| Variant | File              | Purpose                                 |
| ------- | ----------------- | --------------------------------------- |
| Default | `logo.png`        | Tray icon (idle), favicon, social cards |
| Active  | `logo-active.png` | Tray icon (transfer in progress)        |

On macOS, the logo is set as a **template icon** — the system automatically inverts it (black → white in light menu bars, stays black in dark mode). On Linux, the raw PNG is used.

```
mercury/
├── main.go               # Entry point. Starts the god machine.
├── logo.png              # 🔷 Brand logo (source of truth)
├── logo-active.png       # 🔷 Brand logo — active state
├── app/
│   ├── main.go           # Application bootstrap. Wires everything together.
│   ├── app.go            # MercuryApp — the bindings the frontend talks to.
│   ├── icon.png          # Tray icon (resized from logo.png)
│   ├── icon-active.png   # Tray icon active state
│   ├── backend/
│   │   ├── sync/         # LAN sync engine (crypto, peer mgmt, discovery, transport)
│   │   ├── clipboard/    # Clipboard watcher (CGO for macOS file URLs)
│   │   ├── fileinfo/     # File type detection and path resolution
│   │   └── transfer/     # File transfer (sender/receiver)
│   └── system/
│       └── tray.go       # System tray menu builder. Small. Angry. Effective.
├── build/                # Build configuration. You don't need to be here.
├── frontend/
│   ├── public/
│   │   └── mercury-icon.png  # Favicon
│   └── src/              # React + TypeScript + Vite. The pretty face.
├── go.mod                # Dependencies. Handle with care.
└── Taskfile.yml          # Task runner. Type `wails3 dev` and witness magic.
```

## Usage

### Tray

- **Left click** — Settings window appears. Configure your passphrase. Judge your peers.
- **Right click** — Context menu. Mercury header (gods don't need interaction). Peer count (dynamic, because we're generous). Pause/Resume. Quit.

The tray icon has two moods:

- **Active** — `app/icon-active.png` used when a file transfer is in progress
- **Idle** — `app/icon.png` (default, auto-inverts on macOS for light/dark mode)

The logo (`logo.png` and `logo-active.png`) lives in the project root as the source of truth. Resized copies are placed in `app/` (tray icons) and `frontend/public/` (favicon).

### Settings

| Section | What it does                                                                               |
| ------- | ------------------------------------------------------------------------------------------ |
| Sync    | Passphrase input with show/hide. Enable/disable toggle. Start and stop the god machine.    |
| Peers   | Live list of connected devices. Refreshes every 5 seconds because we care.                 |
| Files   | Where received files land (default `~/Mercury/`). Change folder when you feel adventurous. |

## Security Model

Because even gods respect privacy.

1. **Passphrase** — Never transmitted over the network. It stays on your machine. It's a secret. Keep it that way.
2. **PBKDF2** — Key derived once at startup with 100,000 iterations of SHA-256. Never again. We're not animals.
3. **AES-256-GCM** — Authenticated encryption. Wrong key? Decryption error. Silent drop. No drama.
4. **mDNS** — Announces presence only. No keys. No secrets. Just "hey I exist" — the networking equivalent of a nod.
5. **25MB max** for clipboard sync. If your clipboard is larger than 25MB, you're doing something wrong. We silently skip it.

## Performance

| Metric            | Target                                   |
| ----------------- | ---------------------------------------- |
| Idle RAM          | Under 50MB                               |
| Idle CPU          | Effectively 0% (we're not crypto miners) |
| Copy-to-available | Under 500ms on LAN                       |
| Heartbeat         | UDP (not TCP — we're not savages)        |

## Roadmap

- **Phase 1** — Tray shell with settings window ✅
- **Phase 2** — Clipboard sync (text + images) 🚧
- **Phase 3** — File transfer with accept/reject flow ⏳
- **Phase 4** — Polish, autostart, edge cases, icons ⏳

## Contributing

This is a learning project. I'm figuring Go out as I go. If you see something stupid, laugh, then open an issue. Or a PR. Or both.

## License

[MIT](LICENSE). Do what you want. Don't blame me if you paste something embarrassing across the office LAN.

---

_Built with Go, Wails, and an unreasonable amount of sarcasm._
