---
sidebar_position: 7
---

# Contributing

Mercury is an open-source learning project. Contributions are welcome: bug fixes, tests, packaging, and documentation improvements.

## Development setup

```bash
git clone https://github.com/striker561/Mercury.git
cd Mercury
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
GOTOOLCHAIN=go1.25.12 wails3 dev
```

## Tests

```bash
# Backend (clipboard integration tests need a live OS clipboard, excluded in CI)
go test $(go list ./... | grep -v '/clipboard$')

# Frontend
cd frontend && bun install && bun run build
```

## Pull request guidelines

- One focused change per PR
- Add tests for crypto, sync, or transfer logic when behavior changes
- Test on two machines for networking or sync changes
- Match existing code style: practical, minimal comments

## Roadmap

See [TODO.md](https://github.com/striker561/Mercury/blob/main/TODO.md) and [GitHub Issues](https://github.com/striker561/Mercury/issues).

## License

Mercury is **MPL 2.0 licensed**. See [LICENSE](https://github.com/striker561/Mercury/blob/main/LICENSE).
