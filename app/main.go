package app

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mercury/app/system"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// Run is the real Mercury entry point, called from root main.go.
// assets is the embedded frontend/dist from the root module.
func Run(assets embed.FS) error {
	// Detect if running under GNOME (tray may not work without AppIndicator).
	isGNOME := detectGNOME()

	mercuryApp := NewMercuryApp()

	// Create notification service for OS-level file offer alerts.
	// On macOS this requires a bundled .app with a valid CFBundleIdentifier
	// (only available in production builds).  In dev mode we skip it.
	notifySvc := notifications.New()
	mercuryApp.SetNotifier(notifySvc)

	svcs := []application.Service{
		application.NewService(mercuryApp),
	}
	// Only register the notification service when running from a proper
	// macOS bundle (or on Linux/Windows where it always works).
	if !isDarwinDev() {
		svcs = append(svcs, application.NewService(notifySvc))
	} else {
		log.Println("[mercury] skipping notification service (dev mode — no bundle ID)")
	}

	app := application.New(application.Options{
		Name:        "Mercury",
		Description: "LAN Clipboard & File Sharing",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: svcs,
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Wire the autostart manager so the "Start on login" setting works.
	mercuryApp.SetAutostartManager(app.Autostart)

	// Create the settings window (hidden by default, shown on tray click).
	settingsWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "settings",
		Title:            "Mercury Settings",
		Width:            480,
		Height:           640,
		AlwaysOnTop:      true,
		Frameless:        false,
		Hidden:           true,
		DisableResize:    true,
		HideOnEscape:     true,
		BackgroundColour: application.NewRGB(18, 18, 18),
		URL:              "/",
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: true,
		},
	})

	// Hide instead of close — we're a tray app, closing should keep us running.
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("[mercury] window closing → hiding to tray")
		settingsWindow.Hide()
		e.Cancel()
	})

	if isGNOME {
		log.Println("[mercury] GNOME detected — tray may need AppIndicator extension")
		// On GNOME without AppIndicator, show settings window on startup.
		settingsWindow.Show()
	}

	// Wire window show so file offers can pop open the settings.
	mercuryApp.SetShowWindow(func() { settingsWindow.Show() })

	// Create the system tray with an icon (required on macOS — without one,
	// the tray item is invisible in the menu bar).
	tray := app.SystemTray.New()
	// On macOS, SetTemplateIcon auto-inverts for light/dark mode.
	// On Linux, use the white version so it's visible on dark trays.
	tray.SetTemplateIcon(trayIconWhite)
	tray.SetDarkModeIcon(trayIconWhite)
	tray.SetIcon(trayIconWhite)

	// Build the right-click context menu with an "Open Mercury" item.
	menu, refs := system.BuildMenu(app, func() {
		settingsWindow.Show()
		settingsWindow.Focus()
	})
	tray.SetMenu(menu)

	// Wire the pause/resume menu item.
	refs.Pause.OnClick(func(ctx *application.Context) {
		paused := mercuryApp.TogglePause()
		if paused {
			refs.Pause.SetLabel("Resume Sync")
			tray.SetTooltip("Mercury — Paused")
		} else {
			refs.Pause.SetLabel("Pause Sync")
			tray.SetTooltip("Mercury — Running")
		}
	})

	// Update tray status periodically.
	go func() {
		for {
			time.Sleep(2 * time.Second)
			n := mercuryApp.GetPeerCount()
			paused := mercuryApp.IsPaused()

			// Check for active transfers — switch to active icon.
			active := hasActiveTransfers(mercuryApp)
			if active {
				tray.SetTemplateIcon(trayIconActiveWhite)
				tray.SetDarkModeIcon(trayIconActiveWhite)
				tray.SetIcon(trayIconActiveWhite)
			} else {
				tray.SetTemplateIcon(trayIconWhite)
				tray.SetDarkModeIcon(trayIconWhite)
				tray.SetIcon(trayIconWhite)
			}

			var status string
			if paused {
				status = "⏸ Paused"
			} else if n > 0 {
				status = fmt.Sprintf("● Connected (%d peer%s)", n, map[bool]string{true: "", false: "s"}[n == 1])
			} else {
				status = "○ Idle (0 peers)"
			}
			refs.Status.SetLabel(status)
		}
	}()

	// No automatic left-click toggle — the user opens the window from the
	// context menu.  On Linux, right-click opens the menu via the native
	// SecondaryActivate fallback (menu is set, so the menu appears).

	// Register application event listeners.
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(event *application.ApplicationEvent) {
		log.Println("[mercury] application started")
	})

	// Run the application (blocks until exit).  Wails handles tray cleanup
	// internally — calling tray.Destroy() here would double-close a channel.
	err := app.Run()

	return err
}

// isDarwinDev returns true on macOS when NOT running from a bundled .app.
// The notification service needs CFBundleIdentifier which only exists
// inside a proper macOS bundle (production builds).
func isDarwinDev() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	// A bundled .app has a path like .../Mercury.app/Contents/MacOS/mercury
	return !strings.Contains(filepath.ToSlash(exe), ".app/Contents/MacOS/")
}

// detectGNOME checks if we're running under the GNOME desktop environment.
func detectGNOME() bool {
	// Check XDG_CURRENT_DESKTOP and GDMSESSION env vars.
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	session := strings.ToLower(os.Getenv("GDMSESSION"))
	return strings.Contains(desktop, "gnome") || strings.Contains(session, "gnome")
}

// hasActiveTransfers returns true when any transfer is in progress.
func hasActiveTransfers(app *MercuryApp) bool {
	for _, p := range app.GetTransferProgress() {
		if p.Status == "sending" || p.Status == "receiving" {
			return true
		}
	}
	return false
}
