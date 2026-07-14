import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const config: Config = {
  title: "Mercury",
  tagline:
    "Cross-platform LAN clipboard and file sharing. Encrypted, tray-native, no cloud.",
  favicon: "img/mercury-logo.png",

  url: "https://striker561.github.io",
  baseUrl: "/Mercury/",

  organizationName: "striker561",
  projectName: "Mercury",

  onBrokenLinks: "throw",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          sidebarPath: "./sidebars.ts",
          editUrl: "https://github.com/striker561/Mercury/tree/main/docs/",
        },
        blog: false,
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: "img/mercury-logo.png",
    colorMode: {
      defaultMode: "dark",
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: "Mercury",
      logo: {
        alt: "Mercury",
        src: "img/mercury-logo.png",
      },
      items: [
        { to: "/docs/intro", label: "Documentation", position: "left" },
        {
          href: "https://github.com/striker561/Mercury",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      copyright: `Mercury · MPL 2.0 · Built with Go and Wails`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.vsDark,
      additionalLanguages: ["bash", "go"],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
