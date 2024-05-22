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
	var dataListenerForSuccesses binding.DataListener
	deployButton := widget.NewButton("Deploy", func() {
		showDeployDialog(g, dataListenerForSuccesses)
	})
	deployButton.Disable()

	interxStatusBinding := binding.NewBool()
	interxStatusInfo := widget.NewLabel("")
	interxInfoBox := container.NewHBox(
		widget.NewLabel("Interx:"),
		interxStatusInfo,
	)

	shidaiStatusBinding := binding.NewBool()
	shidaiStatusInfo := widget.NewLabel("")
	shidaiInfoBox := container.NewHBox(
		widget.NewLabel("Shidai:"),
		shidaiStatusInfo,
	)

	checkInterxStatus := func() {
		_, err := httph.GetInterxStatus(g.Host.IP)
		if err != nil {
			interxStatusInfo.SetText("interx unavailable")
			log.Printf("ERROR getting interx status: %v", err)
			err = interxStatusBinding.Set(false)
			if err != nil {
				log.Printf("ERROR setting binding: %v", err)
				return
			}
		} else {
			err = interxStatusBinding.Set(true)
			interxStatusInfo.SetText("interx is running")
			if err != nil {
				log.Printf("%v", err)
				return
			}
		}

	}

	checkShidaiStatus := func() {
		_, err := httph.MakeHttpRequest(fmt.Sprintf("http://%v:%v/status", g.Host.IP, 8282), "GET")
		if err != nil {
			log.Printf("ERROR: %v", err)
			shidaiStatusInfo.SetText("shidai unavailable")
			err = shidaiStatusBinding.Set(false)
			if err != nil {
				log.Printf("ERROR: %v", err)
				return
			}
		} else {
			shidaiStatusInfo.SetText("shidai is running")
			err = shidaiStatusBinding.Set(true)
			if err != nil {
				return
			}
		}
	}

	refresh := func() {
		g.WaitDialog.ShowWaitDialog()
		checkInterxStatus()
		checkShidaiStatus()
		shidaiCheck, err := shidaiStatusBinding.Get()
		if err != nil {
			log.Println(err)
		}
		interxCheck, err := interxStatusBinding.Get()
		if err != nil {
			log.Println(err)
		}
		if !shidaiCheck && !interxCheck {
			deployButton.Enable()
		}

		g.WaitDialog.HideWaitDialog()
	}

	refreshButton := widget.NewButton("Refresh", func() {
		refresh()
	})

	dataListenerForSuccesses = binding.NewDataListener(func() {
		deployButton.Disable()
		refresh()
	})
	return container.NewVBox(
		deployButton,
		refreshButton,
		interxInfoBox,
		shidaiInfoBox,
	)

}
