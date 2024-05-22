package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/PeepoFrog/km2UI/gui"
)

func main() {
	a := app.NewWithID("Kira Manager 2.0")
	w := a.NewWindow("Kira Manager 2.0")
	w.SetMaster()
	w.Resize(fyne.NewSize(1024, 768))
	g := gui.Gui{
		Window: w,
	}
	g.WaitDialog = gui.NewWaitDialog(&g)
	content := g.MakeGui()
	g.Window.SetContent(content)
	a.Lifecycle().SetOnStarted(func() {
		g.ShowConnect()
	})
	w.ShowAndRun()
}
