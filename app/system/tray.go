package system

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MenuRefs holds live menu items so the caller can update labels/actions.
type MenuRefs struct {
	Status *application.MenuItem
	Pause  *application.MenuItem
}

// BuildMenu creates the right-click context menu for the system tray.
// Returns references to dynamic items so the caller can update them.
func BuildMenu(app *application.App, showFn func()) (*application.Menu, *MenuRefs) {
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

	// Quit
	quitItem := menu.Add("Quit")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	return menu, &MenuRefs{
		Status: statusItem,
		Pause:  pauseItem,
	}
}
