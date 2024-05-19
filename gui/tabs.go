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
		"nodeInfo": {
			Title: "Node Info",
			Info:  "",
			View:  makeNodeInfoScreen,
		},
		"networkTree": {
			Title: "Network tree",
			Info:  "",
			View:  makeNetworkTreeScreen,
		},
		"test": {},
	}

	TabsIndex = map[string][]string{
		"":     {"status", "nodeInfo", "networkTree"},
		"test": {"a", "b"},
	}
)
