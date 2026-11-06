/** Mercury speaks. Keep control labels plain; pride lives in headlines and asides. */
export const copy = {
  welcome: {
    title: "You have summoned Mercury",
    lead: "I deliver across your LAN. Encrypted. Without cloud worship.",
    steps: [
      "Pick a passphrase. The same words on every device",
      "Run me on each machine on the same Wi‑Fi or LAN",
      "Copy text or files. I carry them. You paste on the other side.",
    ],
    trust: "End-to-end encrypted. Obviously.",
    cta: "Set up sync",
  },

  status: {
    pausedTitle: "I Rest",
    pausedSub: "Resume me in Settings when you are ready to serve again.",
    issueTitle: "Something Is Wrong",
    syncErrorTitle: "Sync Needs Attention",
    syncStoppedTitle: "Sync Is Stopped",
    syncStoppedSub: "I am not listening right now. Retry to start again.",
    searchingTitle: "Searching for Devices",
    searchingSub:
      "I am listening on this network. Open Mercury on another machine with the same passphrase.",
    retry: "Retry",
    retrying: "Retrying…",
    vpnTitle: "VPN Detected",
    vpnSub:
      "Your VPN may block LAN sync. If copies fail one way, turn the VPN off and try again.",
    connectedTitle: "In Service",
    connectedSub: "Copy here. Paste there.",
    gnomeTip:
      "GNOME: install AppIndicator for tray presence, or speak to me through this window.",
  },

  activity: {
    carrying: "Carrying",
  },

  peers: {
    section: "Your fleet",
    unknown: "Unknown device",
    checking: "Checking connection…",
    reachable: "Connected",
  },

  transfers: {
    section: "Cargo",
    incoming: "Incoming from a peer",
    decline: "Decline",
    accept: "Accept",
    cancel: "Cancel",
    dismiss: "Dismiss",
  },

  settings: {
    pageTitle: "Settings",
    pageLead: "How we share a secret, where files land, and when I wake.",
    about: "About",
    source: "Source",
    issues: "Issues",
    docs: "Docs",
    license: "MIT",
    covenant: "Covenant",
    covenantHint:
      "The same words on every device. Saves when you leave the field.",
    offerings: "Offerings",
    offeringsHint: "Where accepted files land.",
    presence: "Presence",
    passphrase: "Passphrase",
    passphrasePlaceholder: "Enter a passphrase",
    passphraseRequired: "Set a passphrase first.",
    syncEnabled: "Clipboard sync",
    syncEnabledHint: "Off means I rest. No watching, no delivering.",
    saveTo: "Save files to",
    change: "Choose…",
    acceptFiles: "Allow file offers",
    acceptFilesHint: "Let peers send you files, not just text and images.",
    autoAccept: "Accept files automatically",
    autoAcceptHint: "Skip the prompt. I save straight to your folder.",
    startOnLogin: "Open at login",
  },
} as const;
