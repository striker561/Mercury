# Architecture

## Dependency rules

**Backend packages never import each other.** They only consume shared utilities:

```
crypto     → (standalone — AES-256-GCM, PBKDF2)
transport  → (standalone — TCP wire protocol)
clipboard  → (standalone — OS clipboard polling)
storage    → (standalone — SQLite key-value store)
sync       → crypto, transport   (clipboard sync over LAN)
transfer   → crypto, transport   (file streaming over LAN)
```

Zero circular dependencies. Adding a new backend package? It should follow the same rule: import only `crypto` and `transport` if needed, never another backend.

## Package map

```
main.go
└─ app/app.go           ← Wails IPC boundary. Thin: creates Mercury, delegates.
   └─ app/services/      ← Composition layer. Each backend has one service wrapper.
      ├─ sync.go         ← SyncService
      ├─ transfer.go     ← TransferService
      └─ clipboard.go    ← ClipboardService
         ├─ app/backend/sync/       ← mDNS discovery, peer map, clipboard event loop
         ├─ app/backend/transfer/   ← file offers, chunked send/receive
         ├─ app/backend/clipboard/  ← OS clipboard watcher
         ├─ app/backend/crypto/     ← Encrypt / Decrypt / DeriveKey
         ├─ app/backend/transport/  ← TCP Send / SendMsg / Listen, type-byte demux
         └─ app/backend/storage/    ← SQLite settings DB
   └─ app/system/        ← System tray menu
```

## Wire protocol (port 47821)

```
[1 byte type][4 bytes big-endian length][payload bytes]
```

| Type | Content | Encryption |
|------|---------|------------|
| `0` (clipboard) | Encrypted text or image | AES-256-GCM |
| `1` (file chunk) | Encrypted 256 KiB file chunk | AES-256-GCM |

The type byte is **not** encrypted — it's a routing hint. The payload is always
encrypted. The sync Manager owns the TCP listener but delegates non-clipboard
messages to the transfer manager via `OnMessage`.

## Data flow

```
Clipboard sync:
  OS clipboard change → clipboard.Watcher → SyncService.BroadcastText/Image
  → crypto.Encrypt → transport.Send (type 0) → peer's transport.Listen
  → sync.eventLoop decrypts → goclipboard.Write

File transfer:
  User sends file → MercuryApp.SendFile → TransferService.SendFile
  → read 256 KiB → crypto.Encrypt → transport.SendMsg (type 1)
  → peer's transport.Listen → OnMessage → TransferService chunkBuf
  → receiveFile writes to disk
```

## Adding a new setting

1. Add a `KeyXxx` constant in `app/backend/storage/settings.go`
2. Add the default in `defaultValues`
3. Done — no migration needed, missing keys return the default

## Key principles

- **No cloud. No accounts. No history.** Everything is ephemeral LAN sync.
- **Passphrase is the only secret.** The AES key is derived from it via PBKDF2.
- **Single port.** Everything goes through 47821 — one firewall rule, no coordination.
- **Services compose, packages don't couple.** The service layer wires independent backend packages together.
