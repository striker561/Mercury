---
sidebar_position: 6
---

# FAQ

## Peers do not appear

- Same **passphrase** on every machine (exact match, case-sensitive).
- Same **LAN:** Mercury does not sync over the public internet.
- Firewall allows **TCP 47821**.
- mDNS enabled on your router (default on most home networks).

## Passphrase mismatch

If peers appear but clipboard never syncs, passphrases may differ. After several decrypt failures, Home shows a warning. Set the same secret on all devices and restart sync.

## VPN

Mercury may show a **VPN detected** warning when a named VPN client interface is active (WireGuard, Tailscale, etc.). macOS system `utun` tunnels are ignored.

VPNs can block or skew LAN discovery. If sync fails, disable the VPN temporarily and retry on the same network.

## Linux tray missing (GNOME / Pop!\_OS)

Install AppIndicator support. The `.deb` lists `libayatana-appindicator3-1` as a dependency. If the icon is still missing, install the [GNOME extension](https://extensions.gnome.org/extension/615/appindicator-support/) and log out/in.

## Clipboard not syncing despite peers

- Confirm sync is **not paused** in Settings.
- Mercury only watches the clipboard when **peers are connected**.
- Payloads over **25 MB** are ignored for clipboard sync. Use file transfer instead.

## Is the logo medical?

No. The **caduceus** (two snakes, wings) is the symbol of Mercury/Hermes. The medical **Rod of Asclepius** has one snake and no wings.

## How is this different from a clipboard manager?

Clipboard managers store history. Mercury **forwards** the current clipboard to peers and does not keep an archive.
