/** Mercury speaks. Keep control labels plain; pride lives in headlines and asides. */
export const copy = {
  welcome: {
    title: "You have summoned Mercury",
    lead: "I deliver across your LAN. Encrypted. Without cloud worship.",
    steps: [
      "Pick a passphrase — the same words on every device",
      "Run me on each machine on the same Wi‑Fi or LAN",
      "Copy text or files. I carry them. You paste on the other side.",
    ],
    trust: "End-to-end encrypted. Obviously.",
    cta: "Set up sync",
  },

  status: {
    pausedTitle: "I rest",
    pausedSub: "Resume me in Settings when you are ready to serve again.",
    issueTitle: "Something is wrong",
    vpnTitle: "VPN detected",
    vpnSub:
      "Your VPN may block LAN sync. If copies fail one way, turn the VPN off and try again.",
    connectedTitle: (n: number) =>
      `In service · ${n} device${n === 1 ? "" : "s"}`,
    connectedSub: "Copy on one machine. Paste on another. I handle the rest.",
    waitingTitle: "Awaiting your fleet",
    waitingSub:
      "Open Mercury on another device. Same passphrase. Same network. I will find it.",
    gnomeTip:
      "GNOME: install AppIndicator for tray presence, or speak to me through this window.",
  },

  header: {
    paused: "Resting",
    idle: "Alone",
    peers: (n: number) => `${n} peer${n === 1 ? "" : "s"}`,
  },

  peers: {
    section: "Your fleet",
    empty: "No devices nearby",
    emptyHint:
      "Other machines show up here once Mercury is running on the same LAN.",
    unknown: "Unknown device",
  },

  transfers: {
    section: "File deliveries",
    incoming: "Incoming from a peer",
    decline: "Decline",
    accept: "Accept",
    cancel: "Cancel",
  },

  settings: {
    covenant: "Sync",
    covenantHint:
      "Your passphrase is the shared secret. Same words on every device. I use it to encrypt what I carry. It never leaves your machines.",
    offerings: "Files",
    offeringsHint:
      "When someone copies a file for you, it appears on Home as an offer. Choose where saved files land.",
    presence: "Startup",
    presenceHint: "Whether I wake when you log in.",
    passphrase: "Passphrase",
    passphrasePlaceholder: "Same on every device",
    syncEnabled: "Clipboard sync",
    syncEnabledHint: "Off means I rest. No watching, no delivering.",
    save: "Save & start",
    saved: "Saved",
    saveTo: "Save files to",
    change: "Choose…",
    acceptFiles: "Allow file offers",
    acceptFilesHint: "Let peers send you files, not just text and images.",
    autoAccept: "Accept files automatically",
    autoAcceptHint: "Skip the prompt. I save straight to your folder.",
    startOnLogin: "Start on login",
  },
} as const;
