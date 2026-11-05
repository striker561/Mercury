package system

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
)

// MenuRefs holds live menu items so the caller can update labels/actions.
type MenuRefs struct {
	Status   *application.MenuItem
	Pause    *application.MenuItem
	CheckUpd *application.MenuItem
}

// BuildMenu creates the right-click context menu for the system tray.
// Returns references to dynamic items so the caller can update them.
func BuildMenu(app *application.App, showFn, onUpdateChecked func()) (*application.Menu, *MenuRefs) {
	menu := application.NewMenu()

	// App name header (disabled)
	menu.Add("Mercury").SetEnabled(false)

	menu.AddSeparator()

	// Open the settings window
	menu.Add("Open Mercury").OnClick(func(ctx *application.Context) {
		if showFn != nil {
			showFn()
		}
	})

	menu.AddSeparator()

	// Dynamic status item (disabled, shows peer count)
	statusItem := menu.Add("○ Awaiting fleet").SetEnabled(false)

	menu.AddSeparator()

	// Pause/Resume toggle
	pauseItem := menu.Add("Rest")
	pauseItem.OnClick(func(ctx *application.Context) {
		// Wired via app-level event binding in main.go
	})

	menu.AddSeparator()

	// Check for updates
	checkItem := menu.Add("Check for Updates…")
	checkItem.OnClick(func(ctx *application.Context) {
		checkItem.SetLabel("Checking for updates…")
		go func() {
			checkCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()
			err := app.Updater.CheckAndInstall(checkCtx)
			if err != nil {
				log.Printf("[mercury] update install: %v", err)
			}
			checkItem.SetLabel(updateResultLabel(err, app.Updater.State()))
			if err == nil && onUpdateChecked != nil {
				onUpdateChecked()
			}
		}()
	})

	menu.AddSeparator()

	// Quit
	quitItem := menu.Add("Quit")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	return menu, &MenuRefs{
		Status:   statusItem,
		Pause:    pauseItem,
		CheckUpd: checkItem,
	}
}

func updateResultLabel(err error, state updater.State) string {
	if err != nil {
		msg := strings.SplitN(err.Error(), "\n", 2)[0]
		if len(msg) > 60 {
			msg = msg[:60] + "…"
		}
		return "Update failed: " + msg
	}
	switch state {
	case updater.StateReady:
		return "Update ready — restart Mercury"
	case updater.StateAvailable, updater.StateDownloading, updater.StateVerifying, updater.StateInstalling:
		return "Update in progress…"
	default:
		return "Up to date"
	}
}
