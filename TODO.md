# Mercury TODO

*Even gods have a to-do list. Ours is shorter than most.*

## Phase 1 — Tray Shell

- [x] Wails v3 project scaffold
- [x] System tray with right-click menu and hidden settings window
- [x] MercuryApp Wails bindings for frontend communication
- [x] Cross-platform build config (Linux, macOS, Windows)
- [x] Dev mode working with Vite hot-reload
- [x] Monochrome Settings UI

## Phase 2 — Clipboard Sync

### Core
- [ ] Clipboard watcher — detect text and image changes on the OS clipboard
- [ ] Wire sync manager into the app (passphrase in, encryption out)
- [ ] TCP broadcast to all known peers on the LAN
- [ ] Receive and decrypt inbound clips, write to local clipboard
- [ ] JSON payload format (text + image types)
- [ ] Source tracking — never re-broadcast a clip that arrived from the network

### Peers
- [ ] Dynamic peer count in tray menu
- [ ] Connected / idle tray state
- [ ] Heartbeat and failure-based eviction

### Testing
- [ ] Text sync between two machines
- [ ] Image sync between two machines
- [ ] Wrong passphrase handling (silent drop)
- [ ] 25MB clipboard limit enforcement

## Phase 3 — File Transfer

- [ ] File offer protocol (name, size metadata)
- [ ] Accept / reject flow with frontend UI
- [ ] Chunked file stream over a dedicated TCP connection
- [ ] Save received files to configurable directory
- [ ] Transfer progress events to frontend
- [ ] Test send/receive between two machines

## Phase 4 — Polish

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
