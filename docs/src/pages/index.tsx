import Link from "@docusaurus/Link";
import useBaseUrl from "@docusaurus/useBaseUrl";
import Layout from "@theme/Layout";
import {
  AppleLogo,
  ArrowRight,
  ClipboardText,
  Devices,
  GithubLogo,
  HardDrives,
  Lock,
  ShieldCheck,
  Tray,
  WifiHigh,
  WindowsLogo,
} from "@phosphor-icons/react";
import styles from "./index.module.css";

const features = [
  {
    icon: ClipboardText,
    title: "Clipboard sync",
    description:
      "Copy text or images on one machine and paste on another. Sub-second delivery on a healthy LAN.",
  },
  {
    icon: HardDrives,
    title: "File transfer",
    description:
      "Send files with accept or decline. Chunked, encrypted transfers with progress in the app.",
  },
  {
    icon: Lock,
    title: "Encrypted locally",
    description:
      "AES-256-GCM with a passphrase-derived key. Your secret never leaves your devices.",
  },
  {
    icon: WifiHigh,
    title: "LAN only",
    description:
      "mDNS discovery on your network. No cloud accounts, no relay servers, no upload queue.",
  },
  {
    icon: Tray,
    title: "Tray-first",
    description:
      "Runs quietly in the system tray. Open the window when you need settings or file offers.",
  },
  {
    icon: ShieldCheck,
    title: "Open source",
    description:
      "Go backend, React UI, Wails v3 shell. Inspect the code, build it yourself, ship it internally.",
  },
];

const steps = [
  {
    title: "Install on each device",
    text: "Download the release for macOS, Linux, or Windows and run Mercury on the same network.",
  },
  {
    title: "Set the same passphrase",
    text: "Open Settings and save a shared secret. Every machine must use identical words.",
  },
  {
    title: "Copy and paste",
    text: "When peers appear on Home, clipboard sync is live. Files can be offered from the clipboard path.",
  },
];

export default function Home(): JSX.Element {
  const logoUrl = useBaseUrl("/img/mercury-logo.png");

  return (
    <Layout
      title="Mercury"
      description="Cross-platform LAN clipboard and file sharing. Encrypted, tray-native, no cloud."
    >
      <div className={styles.page}>
        <header className={styles.hero}>
          <img src={logoUrl} alt="" className={styles.logo} aria-hidden />
          <h1 className={styles.title}>Mercury</h1>
          <p className={styles.lead}>
            Cross-platform clipboard and file sharing for your LAN. Encrypted,
            tray-native, and built to stay out of your way.
          </p>
          <div className={styles.actions}>
            <Link className={styles.primaryBtn} to="/docs/intro">
              Read the docs
              <ArrowRight size={16} weight="bold" />
            </Link>
            <Link
              className={styles.secondaryBtn}
              href="https://github.com/striker561/Mercury"
            >
              <GithubLogo size={18} />
              View on GitHub
            </Link>
          </div>
        </header>

        <main className={styles.section}>
          <p className={styles.sectionTitle}>Features</p>
          <div className={styles.featureGrid}>
            {features.map(({ icon: Icon, title, description }) => (
              <article key={title} className={styles.card}>
                <div className={styles.iconWrap}>
                  <Icon size={22} weight="duotone" />
                </div>
                <h2 className={styles.cardTitle}>{title}</h2>
                <p className={styles.cardText}>{description}</p>
              </article>
            ))}
          </div>

          <p className={styles.sectionTitle}>How it works</p>
          <div className={styles.steps}>
            {steps.map((step, i) => (
              <article key={step.title} className={styles.step}>
                <span className={styles.stepNum}>{i + 1}</span>
                <h2 className={styles.cardTitle}>{step.title}</h2>
                <p className={styles.cardText}>{step.text}</p>
              </article>
            ))}
          </div>

          <p className={styles.sectionTitle}>Platforms</p>
          <div className={styles.platformRow}>
            <div className={styles.platform}>
              <AppleLogo size={20} className={styles.platformIcon} />
              macOS (.dmg)
            </div>
            <div className={styles.platform}>
              <Devices size={20} className={styles.platformIcon} />
              Linux (.deb)
            </div>
            <div className={styles.platform}>
              <WindowsLogo size={20} className={styles.platformIcon} />
              Windows (installer)
            </div>
          </div>
        </main>
      </div>
    </Layout>
  );
}
