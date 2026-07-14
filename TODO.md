# Mercury TODO

_Even gods have a to-do list. Ours is shorter than most._

## Phase 1 - Tray Shell (done)

- [x] Wails v3 project scaffold
- [x] System tray with right-click menu and hidden settings window
- [x] MercuryApp Wails bindings (TypeScript)
- [x] Cross-platform build config (Linux, macOS, Windows)
- [x] Dev mode working with Vite hot-reload

## Phase 2 - Clipboard Sync (done)

- [x] Clipboard watcher (150ms polling + debounce, text + image)
- [x] mDNS discovery, TCP transport, AES-256-GCM encryption
- [x] Pause/Resume support

## Phase 3 - Settings & Storage (done)

- [x] SQLite settings database
- [x] Passphrase, sync, files, autostart settings

## Phase 4 - File Transfer (done)

- [x] Offer/accept flow, chunked transfer, progress, notifications

## Phase 5 - Polish v0.1.0 (done)

- [x] Native-feeling UI (Home/Settings tabs, system light/dark theme)
- [x] Event-driven dashboard (no frontend polling)
- [x] Tray icon active state on sync/transfer
- [x] Welcome screen, save feedback, window close to tray
- [x] GNOME detection + AppIndicator tip in UI
- [x] Wrong-passphrase hint after repeated decrypt failures
- [x] CI (push/PR) + Release workflow (tag `v*`)
- [x] MIT LICENSE
- [x] Linux package deps via nfpm `depends:` + AppIndicator `recommends:`

## Post v0.1.0 (feedback-driven)

- [ ] Oversized clipboard toast (25MB limit, currently log-only)
- [ ] Native folder picker (replace zenity/osascript)
- [ ] Signed release binaries (.deb/.app with notarization)

## Maybe Someday

- [ ] Passphrase easter egg: `> /dev/null`
- [ ] Mercury emoji animation on successful sync
