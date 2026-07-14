# Mercury TODO

_Even gods have a to-do list. Ours is shorter than most._

## ✅ Phase 1 — Tray Shell

- [x] Wails v3 project scaffold
- [x] System tray with right-click menu and hidden settings window
- [x] MercuryApp Wails bindings (TypeScript)
- [x] Cross-platform build config (Linux, macOS, Windows)
- [x] Dev mode working with Vite hot-reload
- [x] Monochrome Settings UI

## ✅ Phase 2 — Clipboard Sync

- [x] Clipboard watcher (150ms polling + debounce, text + image)
- [x] Sync manager wired into the app
- [x] mDNS discovery, TCP transport, AES-256-GCM encryption
- [x] Receive path: network → decrypt → OS clipboard
- [x] Broadcast path: clipboard change → encrypt → TCP to peers
- [x] JSON payload format
- [x] Heartbeat and failure-based eviction
- [x] Pause/Resume support

## ✅ Phase 3 — Settings & Storage

- [x] SQLite settings database
- [x] Constants/defaults system (add new keys anytime, no migrations)
- [x] Settings: passphrase, sync_enabled, paused, allow_files, received_folder, autostart
- [x] Persist passphrase between restarts
- [x] Auto-start sync on launch
- [x] Frontend loads saved values on mount
- [x] Toggle switches for allow_files, autostart
- [x] DB path: ~/.local/share/mercury/mercury.db (XDG compliant)

## ✅ Phase 4 — File Transfer

- [x] File offer protocol (name, size, offer ID)
- [x] Accept / reject flow with frontend UI (offer cards, progress bars)
- [x] Chunked file stream (256 KiB, encrypted, demuxed by type byte)
- [x] file:// URI detection for macOS Finder / Linux Nautilus
- [x] macOS pasteboard reader via CGO (NSFilenamesPboardType)
- [x] Save received files to configurable directory
- [x] Auto-accept setting with received-folder path
- [x] Transfer progress and status tracking
- [x] OS notifications on file offer (macOS bundled, Linux D-Bus)

## 🔄 Phase 5 — Polish (in progress)

- [x] Tray icons (Mercury logo with active state)
- [x] Window close → hide to tray (not quit)
- [x] Settings as full-page overlay
- [x] Fixed window size (no resize)
- [x] Tray menu wired (pause/resume, peer count, tooltip)
- [x] Autostart setting stored (wiring to OS pending)
- [x] macOS notification service skips dev mode gracefully
- [ ] Edge case handling (disconnect, oversized files, wrong key)
- [ ] Linux GNOME AppIndicator detection message
- [ ] Release builds for all platforms
- [ ] Light/dark theme support for UI

## Maybe Someday

- [ ] Smite button — disconnect a peer with extreme prejudice
- [ ] Passphrase easter egg: `> /dev/null`
- [ ] Mercury emoji animation on successful sync
- [ ] Phase of the moon. Why not?
