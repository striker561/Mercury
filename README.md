# Mercury

<p align="center">
  <img src="frontend/public/mercury-logo.png" alt="Mercury logo" width="200"/>
</p>

> _"I carry messages. Yours, specifically. Across your LAN. Without the clouds, without the drama, and definitely without your data ending up in a database you didn't sign up for."_

I am Mercury. Cross-platform LAN clipboard and file sharing. I live in your system tray, judge your networking setup in silence, and deliver whatever you copy on one machine to another. Same passphrase. Same network. That is the covenant. Break it and I will simply stop working. I have no patience for ambiguity.

No cloud. No accounts. No history. **I just work.** Like a messenger god should. For your clipboard. With fewer lightning bolts and zero venture capital.

## What I Do

- **Clipboard sync:** You copy on one machine. You paste on another. I make it so. You're welcome.
- **File transfer:** You send a file. Your peer accepts or declines. I deliver either way. I judge both outcomes equally.
- **System tray:** I sit in your tray like a smug little orb. No dock icon on macOS. I am not here to clutter your life. I am here to run it.
- **Encryption:** AES-256-GCM over your LAN. Your passphrase never leaves your devices. I do not gossip. Gods have standards.

### What I Am NOT

- **Not a clipboard manager.** I am a courier, not a hoarder. No history. No search. No pinning. Copy and move on. I have better things to do than archive your memes.
- **Not cloud-based.** I do not know what "the cloud" is. I use the LAN. It is older, wiser, and does not require a monthly subscription.
- **Not cross-internet.** If you cannot ping each other, I cannot help you. Get closer. I deliver messages, not miracles across continents.

## Why Me?

|                   | Mercury               | Klip                             | Clipp                             |
| ----------------- | --------------------- | -------------------------------- | --------------------------------- |
| Open source       | Yes                   | Partial (free tier caps devices) | Yes                               |
| Device limit      | None                  | 2 free, then pay up              | None                              |
| Clipboard history | No (by design)        | No                               | Yes                               |
| Trust model       | Shared passphrase     | TLS certificates                 | Group passphrase                  |
| Cross-platform    | macOS, Linux, Windows | macOS, Linux, Windows            | macOS, Windows, iOS (+ Linux CLI) |

I exist for mortals who want **simple, unlimited, LAN-only sync** without clipboard hoarding or some startup counting their laptops.

## How I Am Built

| Layer      | Choice               | Why                                                                          |
| ---------- | -------------------- | ---------------------------------------------------------------------------- |
| Framework  | Wails v3             | Writing web views in bare C++ is barbaric                                    |
| Backend    | Go                   | I was built to learn Go. It turned out fast enough to judge you in real time |
| Frontend   | React + TypeScript   | Someone has to suffer. It is not me                                          |
| Bundler    | Bun + Vite           | Fast. Trendy. Acceptable                                                     |
| Encryption | AES-256-GCM + PBKDF2 | Your secrets stay yours                                                      |
| Discovery  | mDNS                 | Zero config. I find your other machines. They find me. As it should be       |

## Getting Started

### Prerequisites

- Go 1.25+ (we use the toolchain, you heathen)
- Bun
- Wails v3 CLI
- **Linux:** `libgtk-4-dev`, `libwebkitgtk-6.0-dev`
- **macOS:** Xcode. You know what to do.
- **Windows:** WebView2. It is probably already there. I do not care either way.

### Install Wails v3

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Run in Dev Mode

```bash
GOTOOLCHAIN=go1.25.12 wails3 dev
```

The tray appears. The frontend hot-reloads. You change code. It updates. For a moment you feel like a god. That is intentional. Do not get used to it.

### Build for Production

```bash
GOTOOLCHAIN=go1.25.12 go build -o mercury .
```

One binary. Zero cloud accounts. Divine.

## Usage

### Tray

- **Right click:** Open me, see your peers, pause, resume, quit. Left click does nothing. I have standards.
- **Tray icon:** I glow when I am working. Otherwise I lurk. You will forget I exist until you need me. That is the point.

### Window

- **Home:** Status, connected devices, incoming file offers.
- **Settings:** Passphrase, sync toggle, file preferences, start on login.
- **Close (×):** Hides to tray. I keep running. Quitting is a choice. Hiding is wisdom.

### Settings

| Section | What it does                                                     |
| ------- | ---------------------------------------------------------------- |
| Sync    | Passphrase shared across your devices, sync on/off               |
| Files   | Save folder (default `~/Downloads/Mercury/`), accept/auto-accept |
| General | Start on login                                                   |

## Security

Even gods respect privacy. Mostly.

1. **Passphrase:** Never transmitted. Stays on your machine. A secret. Keep it that way.
2. **PBKDF2:** 100,000 iterations of SHA-256 at startup. I am not an animal.
3. **AES-256-GCM:** Authenticated encryption. Wrong key? Silent drop. No drama. No second chances.
4. **mDNS:** I announce that I exist on the LAN. No keys. No secrets. Just presence.
5. **25MB max** for clipboard sync. If your clipboard exceeds this, you are doing something unholy and I will ignore it without comment.

## Project Structure

```
mercury/
├── main.go              # Entry point. Where I begin.
├── app/
│   ├── main.go          # Wails bootstrap, tray, window
│   ├── app.go           # MercuryApp. How you speak to me.
│   ├── dashboard.go     # Batched state. I do not do unnecessary round trips.
│   ├── backend/         # sync, clipboard, transfer, crypto, storage
│   └── system/          # Tray menu. Small. Angry. Effective.
├── frontend/
│   ├── public/          # CSS, fonts, logos
│   └── src/             # React UI. The face I show mortals.
└── build/               # Platform packaging. Touch rarely.
```

## Roadmap

See [TODO.md](TODO.md). v0.1.0 is ready. What comes next, mortals will tell me through use. I do not guess.

### Linux packages

`.deb` / `.rpm` packages declare runtime libraries in `depends:` (GTK4, WebKitGTK 6). **apt/dnf install those automatically.** The `preinstall.sh` script does not install packages. It only warns if I am already running during an upgrade. Optional GNOME tray support: `recommends: libayatana-appindicator3-1`. Install it for a tray icon on GNOME. Or do not. The window still works. I am not petty about icons.

## Contributing

This project exists because I wanted to learn Go. That is the truth. I am still learning. The code reflects that.

If you see something stupid in here, **laugh**. Really laugh. Out loud. Then open a PR and fix it. Or open an issue if you are feeling merciful. Keep changes focused. I am a small utility, not a platform, not a lifestyle brand, and not your entire personality.

Pull requests that make me faster, cleaner, or more divine are welcome. Pull requests that add seventeen config options because someone on Hacker News had an opinion are not.

## License

[MIT](LICENSE). Do what you want. Do not blame me if you paste something embarrassing across the office LAN. I delivered it faithfully. That is my job.

---

_Built with Go, Wails, and the arrogance of a god who knows exactly what he is for._
