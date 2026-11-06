---
sidebar_position: 3
---

# Usage

## Tray

| Action                    | Result                                               |
| ------------------------- | ---------------------------------------------------- |
| **Left-click** tray icon  | Show or hide the Mercury window                      |
| **Right-click** tray icon | Open menu: peers, pause, quit                        |
| **Second app launch**     | Focuses the existing instance (no duplicate process) |

On macOS, Mercury uses an accessory app (menu bar only, no Dock icon by default).

On Linux with AppIndicator, a left-click may open the tray menu instead of toggling the window; use **Open Mercury**.

## Window

- **Home:** connection status, peer list, incoming file offers, active transfers
- **Settings:** passphrase, sync toggle, file preferences, start on login
- **Close (×):** hides the window; Mercury keeps running in the tray

## Settings

| Section     | Purpose                                          |
| ----------- | ------------------------------------------------ |
| **Sync**    | Shared passphrase and clipboard sync on/off      |
| **Files**   | Download folder, accept file offers, auto-accept |
| **Startup** | Launch Mercury when you log in                   |

Turning **Accept files** off only rejects new incoming offers; clipboard sync keeps running. Auto-accept is disabled while it is off.

### When Mercury is active

Mercury **rests** (no clipboard polling) when:

- No passphrase is saved
- Sync is paused
- No peers are connected

Clipboard watching starts automatically when at least one peer is online.

## File transfer

1. Copy a **file** in Finder (macOS), **File Explorer** (Windows), or your **Linux file manager** — or copy a file path as text.
2. Mercury detects the file and sends an **offer** to connected peers.
3. The recipient sees the offer on **Home** and can **Accept** or **Decline**. A decline removes the offer locally; no rejection frame is sent to the sender.
4. Accepted files save to the folder configured in Settings (default `~/Downloads/Mercury/`).
5. There is no partial final file: the receiver writes a `.part-*` file and renames it only after the full size arrives. A crash can leave a part behind; Mercury cleans old parts when sync starts.

On macOS, allow **Local Network** and the file-folder prompt if you send from Downloads, Documents, or Desktop.

## Notifications

OS notifications (incoming files) work from bundled builds (`.app`, `.deb`, installer). Linux needs a desktop notification daemon. On macOS, `wails3 dev` runs the dev `.app` bundle, so notifications work there too; a bare `./bin/mercury` falls back to `osascript`.
