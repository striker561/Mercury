---
sidebar_position: 4
---

# Security

Mercury encrypts clipboard payloads and file chunks with **AES-256-GCM** using a **PBKDF2-derived key** from your passphrase. This page summarizes the cryptography and threat model.

## Key Derivation: PBKDF2

When you set a passphrase, Mercury does not use it directly. Instead:

```
passphrase → PBKDF2-HMAC-SHA256 (100,000 iterations) → 256-bit key
```

- **100,000 iterations** meets common OWASP guidance for PBKDF2-HMAC-SHA256.
- Derivation runs once at startup; the key stays in memory for the process lifetime.
- The **salt** is a fixed string (`"mercury-lan-sync-v1"`), hardcoded in the source. This ensures the same passphrase produces the same key on every device, so all devices with the same passphrase derive the same key. No per-installation salt is stored on disk.

## Encryption: AES-256-GCM

Every clipboard payload and every file chunk is encrypted with **AES-256 in GCM mode**.

| Property                      | What it means                                                                     |
| ----------------------------- | --------------------------------------------------------------------------------- |
| **256-bit key**               | Brute-forcing this would take longer than the universe has been around            |
| **GCM (Galois/Counter Mode)** | Authenticated encryption; tampering is detected before decryption                 |
| **Random nonce**              | A fresh 12-byte nonce per message. Same plaintext encrypts differently every time |
| **Authentication tag**        | 16-byte tag appended to each ciphertext. Wrong key? Tag fails. Message dropped.   |

### What happens with the wrong key

If a peer uses a different passphrase, the GCM authentication tag fails and Mercury **drops the message** without writing to the clipboard.

## mDNS Discovery

Mercury announces itself on the LAN using **mDNS** (Multicast DNS) under the service type `_mercury._tcp`.

| Leaks                        | Does not leak                                |
| ---------------------------- | -------------------------------------------- |
| Your hostname                | Your passphrase or key material              |
| Your IP address              | The contents of your clipboard               |
| That you are running Mercury | File names or transfer metadata              |
| TCP port 47821               | Anything beyond "a Mercury node exists here" |

mDNS is a local-only protocol. Announcements do not leave your subnet unless you have configured mDNS reflection (and if you have, you probably know what you are doing).

**The threat:** A neighbour on the same Wi-Fi can see that someone is running Mercury. They cannot see your data, your key, or what you are copying.

## The 25 MB Cap

Clipboard payloads larger than **25 MB** are ignored. Use **file transfer** for large files.

## Threat Model

### Who this protects against

| Attacker                     | Outcome                                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------------------- |
| **Same-network snoop**       | Can see mDNS announcements and encrypted traffic. Cannot decrypt.                           |
| **Rogue device on your LAN** | Cannot join without the passphrase. Cannot decrypt captured traffic.                        |
| **Casual packet capture**    | Encrypted payloads. No metadata beyond "something was sent on port 47821".                  |
| **Malicious file**           | Files are delivered as-is. Mercury does not scan for malware. That is your antivirus's job. |

### Who this does NOT protect against

| Scenario                              | Why                                                                                                 |
| ------------------------------------- | --------------------------------------------------------------------------------------------------- |
| **Device compromise**                 | If someone has access to your machine, they have the key in memory.                                 |
| **Advanced persistent threat on LAN** | A determined attacker with ARP spoofing or MITM tools could disrupt service (but not decrypt data). |
| **Physical access**                   | The passphrase is saved in plaintext in the SQLite settings database. The derived AES key is in memory only and lost on shutdown.                                 |
| **Malware on a peer device**          | Your trusted peer can exfiltrate decrypted clipboard content. Trust your peers.                     |

**Bottom line:** Mercury protects your clipboard content from being read by anyone who does not have the passphrase. If someone has access to one of your devices or can physically intercept your network at the switch level, all bets are off. But that is true of almost everything.
