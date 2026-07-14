# Mercury TODO

*Even gods have a to-do list. Ours is shorter than most.*

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

## 🚧 Phase 3 — Settings & Storage

- [x] SQLite settings database (`app/backend/storage/`)
- [x] Persist passphrase between restarts
- [x] Auto-start sync on launch if saved passphrase exists
- [ ] Save received folder path
- [ ] Settings UI pre-fills from saved values
- [ ] Settings UI — Change folder button

## ⏳ Phase 4 — File Transfer

- [ ] File offer protocol (name, size metadata)
- [ ] Accept / reject flow with frontend UI
- [ ] Chunked file stream over a dedicated TCP connection
- [ ] Save received files to configurable directory
- [ ] Transfer progress events to frontend

## ⏳ Phase 5 — Polish

- [ ] Tray icons (connected / idle states)
- [ ] Autostart toggle
- [ ] Edge case handling (disconnect, oversized files, wrong key)
- [ ] Linux GNOME AppIndicator detection message
- [ ] Release builds for all platforms

## Maybe Someday

- [ ] Smite button — disconnect a peer with extreme prejudice
- [ ] Passphrase easter egg: `> /dev/null`
- [ ] Mercury emoji animation on successful sync
- [ ] Phase of the moon. Why not?
