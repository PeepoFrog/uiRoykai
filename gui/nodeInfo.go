package gui

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/PeepoFrog/km2UI/helper/httph"
	"github.com/PeepoFrog/km2UI/types"
)

type InfoScreen struct {
	NodeIP string
}

func makeNodeInfoScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {
	var i InfoScreen = InfoScreen{
		// NodeIP: g.Host.IP,
		NodeIP: "148.251.69.56",
	}
	latestBlockData := binding.NewString()
	latestBlockLabel := widget.NewLabelWithData(latestBlockData)

	latestBlockBox := container.NewVBox(
		widget.NewLabel("latestBlock"), latestBlockLabel,
	)

	refreshScreen := func() {
		wd := NewWaitDialog(g)
		wd.ShowWaitDialog()
		// donechan := make(chan bool)
		// showWaitDialog(g, donechan)
		i, err := i.getInterxStatus()
		if err != nil {

		}
		err = latestBlockData.Set(i.InterxInfo.LatestBlockHeight)
		if err != nil {

		}
		// donechan <- true
		wd.HideWaitDialog()

	}

	refreshButton := widget.NewButton("Refresh", refreshScreen)
	mainInfo := container.NewVScroll(
		container.NewHBox(
			latestBlockBox,
		),
	)
	return container.NewBorder(nil, container.NewGridWithColumns(3, container.NewStack(), refreshButton, container.NewStack()), nil, nil, mainInfo)
}

func (i InfoScreen) getInterxStatus() (*types.Info, error) {
	url := fmt.Sprintf("http://%v:11000/api/status", i.NodeIP)
	b, err := httph.MakeHttpRequest(url, "GET")
	if err != nil {
		return nil, err
	}
	var info types.Info
	err = json.Unmarshal(b, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// func (i InfoScreen) getValoperInfo() (int, error) {}
