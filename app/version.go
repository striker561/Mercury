// Package app — version.
//
// The app version lives in one place: this var. Both GetVersion() and
// GetAllSettings() reference it. Bump it once and you're done.
//
// ── Dev ──
// The default below matches build/config.yml so wails3 dev just works.
//
// ── Release ──
// Override at build time with -ldflags so the binary carries the exact
// tag without any hardcoded drift:
//
//	GOTOOLCHAIN=go1.25.12 go build  \
//		-ldflags "-X mercury/app.Version=v0.2.0"  \
//		-o bin/mercury .
//
// Why -ldflags -X instead of embedding build/config.yml at runtime?
// Wails v3 parses that YAML only during its CLI commands (dev, build);
// the compiled Go binary never sees it. -ldflags is the standard Go
// mechanism for injecting build-time strings — it's zero-cost at
// runtime, works across all platforms, and every Go developer recognises
// the pattern.
package app

var Version = "0.2.0"
