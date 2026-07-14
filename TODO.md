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

## ✅ Phase 5 — Polish

- [x] Tray icons (Mercury logo with idle + active state, dark/light theme)
- [x] Window close → hide to tray (not quit)
- [x] Full-page settings view with gear toggle
- [x] Fixed window size 380×520, frameless, always-on-top
- [x] Tray menu wired (pause/resume, peer count, tooltip)
- [x] Autostart wired to OS via Wails AutostartManager
- [x] macOS notification service skips dev mode gracefully
- [x] Cancel transfers + partial file cleanup
- [x] Transfer speed (MB/s) + live progress bar (200ms tick)
- [x] Optimistic UI updates (accept/reject/cancel immediate)
- [x] Active tray icon on clipboard sync (flashes 2s)
- [x] Welcome/intro screen for first-run
- [x] Folder picker dialog (zenity/osascript)
- [x] Phosphor icons, redesigned dark UI
- [x] Right-click disabled in webview
- [x] CI workflow (push/PR) + Release workflow (tagged)
- [x] GNOME detection with actionable tip
- [ ] Edge case handling (disconnect, oversized files, wrong key)
- [ ] Light/dark theme support for UI

## Maybe Someday

- [ ] Smite button — disconnect a peer with extreme prejudice
- [ ] Passphrase easter egg: `> /dev/null`
- [ ] Mercury emoji animation on successful sync
- [ ] Phase of the moon. Why not?
