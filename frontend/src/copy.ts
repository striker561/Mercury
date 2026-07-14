/** Mercury speaks. Keep control labels plain; pride lives in headlines and asides. */
export const copy = {
  welcome: {
    title: "You have summoned Mercury",
    lead: "I deliver across your LAN. Encrypted. Without cloud worship.",
    steps: [
      "Swear a passphrase shared with your devices",
      "Run me on each machine on the same network",
      "Copy anything. I carry it. You paste it. Simple.",
    ],
    trust: "End-to-end encrypted. Obviously.",
    cta: "Set the covenant",
  },

  status: {
    pausedTitle: "I rest",
    pausedSub: "Resume me in Settings when you are ready to serve again.",
    issueTitle: "The covenant is broken",
    connectedTitle: (n: number) =>
      `In service · ${n} device${n === 1 ? "" : "s"}`,
    connectedSub: "Copy anywhere. I deliver everywhere.",
    waitingTitle: "Awaiting your fleet",
    waitingSub:
      "Run Mercury on each machine. Same passphrase. Same network. I will find them.",
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
    emptyHint: "They appear when Mercury runs elsewhere on your LAN.",
    unknown: "Unknown mortal",
  },

  transfers: {
    section: "Deliveries",
    incoming: "Offered to you",
    decline: "Refuse",
    accept: "Accept",
    cancel: "Stop",
  },

  settings: {
    covenant: "The covenant",
    offerings: "Offerings",
    presence: "Presence",
    passphrase: "Passphrase",
    passphrasePlaceholder: "Shared secret",
    syncEnabled: "I am listening",
    save: "Seal covenant",
    saved: "Sealed",
    saveTo: "Save to",
    change: "Change",
    acceptFiles: "Accept offerings",
    autoAccept: "Accept without asking",
    startOnLogin: "Rise on login",
  },
} as const;
