package app

import (
	"embed"
	"log"
	"os"
	"strings"

	"mercury/app/system"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Run is the real Mercury entry point, called from root main.go.
// assets is the embedded frontend/dist from the root module.
func Run(assets embed.FS) error {
	// Detect if running under GNOME (tray may not work without AppIndicator).
	isGNOME := detectGNOME()

	app := application.New(application.Options{
		Name:        "Mercury",
		Description: "LAN Clipboard & File Sharing",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			application.NewService(NewMercuryApp()),
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
		BackgroundColour: application.NewRGB(18, 18, 18),
		URL:              "/",
	})

	if isGNOME {
		log.Println("[mercury] GNOME detected — tray may need AppIndicator extension")
		// On GNOME without AppIndicator, show settings window on startup.
		settingsWindow.Show()
	}

	// Create the system tray.
	tray := app.SystemTray.New()

	// Build the right-click context menu.
	menu := system.BuildMenu(app)
	tray.SetMenu(menu)

	// Attach settings window to tray — left click toggles it.
	tray.AttachWindow(settingsWindow)

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
