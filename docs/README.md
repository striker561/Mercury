# Mercury Documentation

Static documentation site for Mercury, built with [Docusaurus](https://docusaurus.io/).

## Local development

```bash
cd docs
bun install
bun start
```

Open [http://localhost:3000/Mercury/](http://localhost:3000/Mercury/).

## Build

```bash
bun run build
```

Output is written to `docs/build/`.

## Deploy to GitHub Pages

```bash
GIT_USER=<username> bun run deploy
```

Or publish the `docs/build` folder from CI.

## Design

The site uses the same dark/light tokens as the Mercury app: Inter typography, green accent, grouped cards on the homepage, and Phosphor icons.
