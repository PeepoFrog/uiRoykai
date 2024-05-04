package gui

import "fyne.io/fyne/v2"

type Tab struct {
	Title, Info string
	View        func(w fyne.Window, g *Gui) fyne.CanvasObject
}

var (
	Tabs = map[string]Tab{
		"status": {
			Title: "Status screen",
			Info:  "",
			View:  makeStatusScreen,
		},
	}

	TabsIndex = map[string][]string{
		"": {"status"},
	}
)
