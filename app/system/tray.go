package system

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// BuildMenu creates the right-click context menu for the system tray.
// The menu items with dynamic labels (peer count, pause state) are returned
// so the caller can update them later.
func BuildMenu(app *application.App, showFn func()) *application.Menu {
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
	statusItem := menu.Add("● Connected (0 peers)").SetEnabled(false)

	menu.AddSeparator()

	// Pause/Resume toggle
	pauseItem := menu.Add("Pause Sync")
	pauseItem.OnClick(func(ctx *application.Context) {
		// Will be wired in Phase 2
	})

	menu.AddSeparator()

	// Quit
	quitItem := menu.Add("Quit")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// Store references for dynamic updates.
	// TODO: Wire status updates via event listeners in Phase 2.
	_ = statusItem
	_ = pauseItem

	return menu
}
