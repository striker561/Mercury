package app

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	configureLinuxRenderer()

	// Detect if running under GNOME (tray may not work without AppIndicator).
	isGNOME := detectGNOME()

	mercuryApp := NewMercuryApp()
	if isGNOME {
		mercuryApp.setGNOMETrayTip(true)
	}

	// Create notification service for OS-level file offer alerts. On macOS it
	// needs a bundled .app with a CFBundleIdentifier; bare dev binaries fall
	// back to osascript.
	notifySvc := notifications.New()

	svcs := []application.Service{
		application.NewService(mercuryApp),
	}
	if !isDarwinDev() {
		svcs = append(svcs, application.NewService(notifySvc))
		mercuryApp.setNotifier(notifySvc)
	} else {
		log.Println("[mercury] notification service unavailable (dev mode — using osascript)")
	}

	var settingsWindow *application.WebviewWindow
	var showSettings func()

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
				if showSettings != nil {
					showSettings()
				}
			},
			ExitCode: 0,
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		Linux: application.LinuxOptions{
			DisableQuitOnLastWindowClosed: true,
		},
	})

	// Wire the autostart manager so the "Start on login" setting works.
	mercuryApp.setAutostartManager(app.Autostart)

	// Wire the in-app updater against GitHub Releases.
	if err := configureUpdater(app); err != nil {
		return err
	}
	startSilentUpdateCheck(app, mercuryApp)

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

	showSettings = func() {
		if settingsWindow == nil {
			return
		}
		// app.Show is the one call that does not marshal itself to the main
		// thread; the window methods below already do.
		application.InvokeSync(app.Show)
		settingsWindow.Restore()
		settingsWindow.Show()
		settingsWindow.Focus()
	}

	toggleWindow := func() {
		if !settingsWindow.IsVisible() {
			showSettings()
			return
		}
		if !settingsWindow.IsFocused() {
			settingsWindow.Restore()
			settingsWindow.Focus()
			return
		}
		settingsWindow.Hide()
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
	mercuryApp.setShowWindow(showSettings)
	mercuryApp.setHideWindow(func() {
		settingsWindow.Hide()
	})

	// Create the system tray with an icon (required on macOS — without one,
	// the tray item is invisible in the menu bar).
	tray := app.SystemTray.New()

	menu, refs := system.BuildMenu(app, showSettings, mercuryApp.markUpdateChecked)
	tray.SetMenu(menu)

	// Left-click toggles the window; right-click opens the tray menu.
	tray.OnClick(toggleWindow)

	mercuryApp.setEmitChange(func() {
		app.Event.Emit("dashboard:changed")
		updateTray(tray, refs, mercuryApp)
	})
	mercuryApp.setFileOfferUI(func() {
		app.Event.Emit("dashboard:file-offer")
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
	return app.Run()
}

// isDarwinDev returns true on macOS when NOT running from a bundled .app.
// The notification service needs a CFBundleIdentifier, which only exists
// inside a proper macOS bundle.
func isDarwinDev() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	// A bundled .app has a path like .../Mercury.app/Contents/MacOS/mercury
	return !strings.Contains(filepath.ToSlash(exe), ".app/Contents/MacOS/")
}

// Temporary: Wails v3.1.0 stopped applying WebKitGTK's NVIDIA DMA-BUF workaround
// on hybrid Wayland, and GTK4's GL renderer aborts when EGL is unavailable.
// Remove when upstream handles both.
const linuxRenderFallback = true

func configureLinuxRenderer() {
	if !linuxRenderFallback || runtime.GOOS != "linux" {
		return
	}
	hasNvidia := hasNVIDIA()
	if hasNvidia && os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER") == "" {
		_ = os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	sessionType := os.Getenv("XDG_SESSION_TYPE")
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	gskRenderer := os.Getenv("GSK_RENDERER")
	force := os.Getenv("MERCURY_FORCE_CAIRO") != ""

	if waylandDisplay != "" || sessionType == "wayland" || force {
		log.Printf("[mercury] renderer check: session=%q desktop=%q gdk=%q gsk=%q wayland=%q nvidia=%v force=%v",
			sessionType, desktop, os.Getenv("GDK_BACKEND"), gskRenderer, waylandDisplay, hasNvidia, force)
	}
	if !needsCairoRenderer(runtime.GOOS, gskRenderer, sessionType, waylandDisplay, desktop, hasNvidia, force) {
		return
	}
	_ = os.Setenv("GSK_RENDERER", "cairo")
	if force {
		log.Println("[mercury] MERCURY_FORCE_CAIRO set; using GTK Cairo renderer")
	} else {
		log.Println("[mercury] COSMIC NVIDIA Wayland detected; using GTK Cairo renderer")
	}
}

func needsCairoRenderer(goos, gskRenderer, sessionType, waylandDisplay, desktop string, hasNVIDIA, force bool) bool {
	if goos != "linux" {
		return false
	}
	if force {
		return true
	}
	wayland := sessionType == "wayland" || waylandDisplay != ""
	desktop = strings.ToUpper(desktop)
	return gskRenderer == "" && wayland && hasNVIDIA &&
		(strings.Contains(desktop, "COSMIC") || strings.Contains(desktop, "POP"))
}

func hasNVIDIA() bool {
	for _, path := range []string{"/sys/module/nvidia", "/proc/driver/nvidia/version", "/dev/nvidia0"} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// detectGNOME checks if we're running under the GNOME desktop environment.
func detectGNOME() bool {
	// Check XDG_CURRENT_DESKTOP and GDMSESSION env vars.
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	session := strings.ToLower(os.Getenv("GDMSESSION"))
	return strings.Contains(desktop, "gnome") || strings.Contains(session, "gnome")
}

func setTrayIcon(tray *application.SystemTray, active bool) {
	icon := trayIconWhite
	if active {
		icon = trayIconActiveWhite
	}
	switch runtime.GOOS {
	case "darwin":
		tray.SetTemplateIcon(icon)
	case "windows":
		tray.SetIcon(icon)
		tray.SetDarkModeIcon(icon)
	default:
		tray.SetIcon(icon)
	}
}

// updateTray refreshes the tray icon and status label from current app state.
func updateTray(tray *application.SystemTray, refs *system.MenuRefs, mercuryApp *MercuryApp) {
	active := mercuryApp.trayActive()
	n := mercuryApp.GetPeerCount()
	paused := mercuryApp.IsPaused()
	var status string
	if paused {
		status = "⏸ Resting"
	} else if n > 0 {
		peerSuffix := "s"
		if n == 1 {
			peerSuffix = ""
		}
		status = fmt.Sprintf("● In service (%d peer%s)", n, peerSuffix)
	} else {
		status = "○ Awaiting fleet"
	}

	setTrayIcon(tray, active)
	refs.Status.SetLabel(status)
}
