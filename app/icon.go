package app

import _ "embed"

//go:embed icon.png
var trayIcon []byte

//go:embed icon-active.png
var trayIconActive []byte

//go:embed icon-white.png
var trayIconWhite []byte

//go:embed icon-active-white.png
var trayIconActiveWhite []byte
