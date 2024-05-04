package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/PeepoFrog/km2UI/gui"
)

func main() {
	a := app.NewWithID("UI for shidai")
	w := a.NewWindow("UI for shidai")
	w.SetMaster()
	w.Resize(fyne.NewSize(1024, 768))
	g := gui.Gui{
		Window: w,
	}
	content := g.MakeGui()
	g.Window.SetContent(content)
	a.Lifecycle().SetOnStarted(func() {
		g.ShowConnect()
	})
	w.ShowAndRun()
}
