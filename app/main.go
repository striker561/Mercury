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

const singleInstanceID = "com.mercury.app"

// Run is the real Mercury entry point, called from root main.go.
// assets is the embedded frontend/dist from the root module.
func Run(assets embed.FS) error {
	// Detect if running under GNOME (tray may not work without AppIndicator).
	isGNOME := detectGNOME()

	mercuryApp := NewMercuryApp()
	if isGNOME {
		mercuryApp.SetGNOMETrayTip(true)
	}

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

	var settingsWindow *application.WebviewWindow

	app := application.New(application.Options{
		Name:        "Mercury",
		Description: "LAN Clipboard & File Sharing",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: svcs,
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: singleInstanceID,
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				log.Printf("[mercury] second instance launch args=%v", data.Args)
				if settingsWindow != nil {
					settingsWindow.Show()
					settingsWindow.Focus()
				}
			},
			ExitCode: 0,
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Wire the autostart manager so the "Start on login" setting works.
	mercuryApp.SetAutostartManager(app.Autostart)

	// Create the settings window (hidden by default, shown on tray click).
	settingsWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "settings",
		Title:            "Mercury",
		Width:            420,
		Height:           560,
		Frameless:        true,
		Hidden:           true,
		DisableResize:    true,
		HideOnEscape:     true,
		BackgroundColour: application.NewRGB(15, 17, 23),
		URL:              "/",
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: true,
		},
	})

	toggleWindow := func() {
		if settingsWindow.IsVisible() {
			settingsWindow.Hide()
		} else {
			settingsWindow.Show()
			settingsWindow.Focus()
		}
	}

	// Hide instead of close — we're a tray app, closing should keep us running.
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("[mercury] window closing → hiding to tray")
		settingsWindow.Hide()
		e.Cancel()
	})

	if isGNOME {
		log.Println("[mercury] GNOME detected — showing window at startup")
		log.Println("[mercury]   tip: install gnome-shell-extension-appindicator for tray support")
		// On GNOME without AppIndicator, show settings window on startup so
		// the user can still use Mercury even without a visible tray icon.
		settingsWindow.Show()
	}

	// Wire window show/hide so file offers can pop open the settings.
	mercuryApp.SetShowWindow(func() {
		settingsWindow.Show()
		settingsWindow.Focus()
	})
	mercuryApp.SetHideWindow(func() {
		settingsWindow.Hide()
	})

	// Create the system tray with an icon (required on macOS — without one,
	// the tray item is invisible in the menu bar).
	tray := app.SystemTray.New()
	tray.SetTemplateIcon(trayIconWhite)
	tray.SetDarkModeIcon(trayIconWhite)
	tray.SetIcon(trayIconWhite)

	menu, refs := system.BuildMenu(app, func() {
		settingsWindow.Show()
		settingsWindow.Focus()
	})
	tray.SetMenu(menu)

	// Left-click toggles the window; right-click opens the tray menu.
	tray.OnClick(toggleWindow)

	mercuryApp.SetEmitChange(func() {
		app.Event.Emit("dashboard:changed")
		updateTray(tray, refs, mercuryApp)
	})

	go func() {
		var last string
		for {
			interval := 400 * time.Millisecond
			if settingsWindow != nil && !settingsWindow.IsVisible() {
				interval = 3 * time.Second
			}
			time.Sleep(interval)

			fp := mercuryApp.DashboardFingerprint()
			if fp != last {
				last = fp
				mercuryApp.syncClipboardWatch()
				app.Event.Emit("dashboard:changed")
				updateTray(tray, refs, mercuryApp)
			}
		}
	}()

	// Wire the pause/resume menu item.
	refs.Pause.OnClick(func(ctx *application.Context) {
		paused := mercuryApp.TogglePause()
		if paused {
			refs.Pause.SetLabel("Awaken")
			tray.SetTooltip("Mercury - Resting")
		} else {
			refs.Pause.SetLabel("Rest")
			tray.SetTooltip("Mercury - In service")
		}
		updateTray(tray, refs, mercuryApp)
	})

	updateTray(tray, refs, mercuryApp)

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

// updateTray refreshes the tray icon and status label from current app state.
func updateTray(tray *application.SystemTray, refs *system.MenuRefs, mercuryApp *MercuryApp) {
	if mercuryApp.trayActive() {
		tray.SetTemplateIcon(trayIconActiveWhite)
		tray.SetDarkModeIcon(trayIconActiveWhite)
		tray.SetIcon(trayIconActiveWhite)
	} else {
		tray.SetTemplateIcon(trayIconWhite)
		tray.SetDarkModeIcon(trayIconWhite)
		tray.SetIcon(trayIconWhite)
	}

	n := mercuryApp.GetPeerCount()
	paused := mercuryApp.IsPaused()
	var status string
	if paused {
		status = "⏸ Resting"
	} else if n > 0 {
		status = fmt.Sprintf("● In service (%d peer%s)", n, map[bool]string{true: "", false: "s"}[n == 1])
	} else {
		status = "○ Awaiting fleet"
	}
	refs.Status.SetLabel(status)
}
