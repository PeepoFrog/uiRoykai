package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/PeepoFrog/km2UI/helper/httph"
)

func makeStatusScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {

	deployButton := widget.NewButton("DEPLOY NODE", func() {})
	deployButton.Disable()

	interxStatusBinding := binding.NewBool()
	interxStatusInfo := widget.NewLabel("")
	interxInfoBox := container.NewHBox(
		widget.NewLabel("Interx:"),
		interxStatusInfo,
	)

	shidaiStatusBinding := binding.NewBool()
	shidaiStatusInfo := widget.NewLabel("")
	shidaiInfoBox := container.NewVBox(
		widget.NewLabel("Shidai:"),
		shidaiStatusInfo,
	)

	refreshButton := widget.NewButton("Refresh", func() {
		_, err := httph.MakeHttpRequest(fmt.Sprintf("http://%v:%v/api/status", g.Host.IP, 11000), "GET")
		if err != nil {
			log.Printf("ERROR: %v", err)
			err = interxStatusBinding.Set(false)
			if err != nil {
				log.Printf("ERROR: %v", err)
				return
			}
		} else {
			interxStatusBinding.Set(true)
		}

		_, err = httph.MakeHttpRequest(fmt.Sprintf("http://%v:%v/status", g.Host.IP, 8282), "GET")
		if err != nil {
			log.Printf("ERROR: %v", err)
			interxStatusInfo.SetText("interx unavailable")
			err = shidaiStatusBinding.Set(false)
			if err != nil {
				log.Printf("ERROR: %v", err)
				return
			}
		} else {
			err = shidaiStatusBinding.Set(true)
			log.Printf("ERROR: %v", err)
			if err != nil {
				interxStatusInfo.SetText("shidai unavailable")
				log.Printf("ERROR: %v", err)
				return
			}
		}
		shidaiCheck, _ := shidaiStatusBinding.Get()
		interxCheck, _ := shidaiStatusBinding.Get()
		if !shidaiCheck || !interxCheck {
			deployButton.Enable()
		}

	})

	return container.NewVBox(
		deployButton,
		refreshButton,
		interxInfoBox,
		shidaiInfoBox,
	)

}
