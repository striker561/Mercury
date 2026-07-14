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

### When Mercury is active

Mercury **rests** (no clipboard polling) when:

- No passphrase is saved
- Sync is paused
- No peers are connected

Clipboard watching starts automatically when at least one peer is online.

## File transfer

1. Copy a **file path** or file from your file manager.
2. Mercury detects the file and sends an **offer** to connected peers.
3. The recipient sees the offer on **Home** and can **Accept** or **Decline**.
4. Accepted files save to the folder configured in Settings (default `~/Downloads/Mercury/`).

## Notifications

OS notifications (incoming files) work from **installed** builds (`.app`, `.deb`, installer). `wails3 dev` on macOS skips notifications because there is no bundle ID in dev mode.
