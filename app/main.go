package app

import (
	"embed"
	"log"
	"os"
	"strings"

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

	// Create notification service for OS-level alerts.
	notifySvc := notifications.New()
	mercuryApp.SetNotifier(notifySvc)

	app := application.New(application.Options{
		Name:        "Mercury",
		Description: "LAN Clipboard & File Sharing",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			application.NewService(mercuryApp),
			application.NewService(notifySvc),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Create the settings window (hidden by default, shown on tray click).
	settingsWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "settings",
		Title:            "Mercury Settings",
		Width:            480,
		Height:           640,
		AlwaysOnTop:      true,
		Frameless:        false,
		Hidden:           true,
		MinWidth:         400,
		MinHeight:        500,
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
	menu := system.BuildMenu(app, func() {
		settingsWindow.Show()
		settingsWindow.Focus()
	})
	tray.SetMenu(menu)

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

// detectGNOME checks if we're running under the GNOME desktop environment.
func detectGNOME() bool {
	// Check XDG_CURRENT_DESKTOP and GDMSESSION env vars.
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	session := strings.ToLower(os.Getenv("GDMSESSION"))
	return strings.Contains(desktop, "gnome") || strings.Contains(session, "gnome")
}
