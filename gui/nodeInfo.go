package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"

	"github.com/PeepoFrog/km2UI/helper/httph"
)

// type InfoScreen struct {
// 	NodeIP string
// }

func makeNodeInfoScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {
	// TODO: only for testing, delete later
	// g.Host.IP = "148.251.69.56"
	//

	// latest block box
	latestBlockData := binding.NewString()
	latestBlockLabel := widget.NewLabelWithData(latestBlockData)
	latestBlockBox := container.NewHBox(
		widget.NewLabel("latest Block"), latestBlockLabel,
	)

	// validator address box
	validatorAddressData := binding.NewString()
	validatorAddressLabel := widget.NewLabelWithData(validatorAddressData)
	validatorAddressBox := container.NewHBox(
		widget.NewLabel("validator address: "), validatorAddressLabel,
		widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
			data, err := validatorAddressData.Get()
			if err != nil {
				log.Println(err)
				return
			}

			err = clipboard.WriteAll(data)
			if err != nil {
				return
			}
		}),
	)
	// public ip box
	publicIpData := binding.NewString()
	publicIpLabel := widget.NewLabelWithData(publicIpData)
	publicIpBox := container.NewHBox(
		widget.NewLabel("public ip address: "), publicIpLabel,
	)
	// miss chance box
	missChanceData := binding.NewString()
	missChanceLabel := widget.NewLabelWithData(missChanceData)
	missChanceBox := container.NewHBox(
		widget.NewLabel("miss chance: "), missChanceLabel,
	)

	refreshScreen := func() {
		g.WaitDialog.ShowWaitDialog()
		defer g.WaitDialog.HideWaitDialog()
		i, err := httph.GetInterxStatus(g.Host.IP)
		if err != nil {
			return
		}
		err = latestBlockData.Set(i.InterxInfo.LatestBlockHeight)
		if err != nil {
			return
		}

	}

	refreshButton := widget.NewButton("Refresh", refreshScreen)
	mainInfo := container.NewVScroll(
		container.NewVBox(
			latestBlockBox,
			validatorAddressBox,
			publicIpBox,
			missChanceBox,
		),
	)
	return container.NewBorder(nil, container.NewGridWithColumns(3, container.NewStack(), refreshButton, container.NewStack()), nil, nil, mainInfo)
}

// func getInterxStatus() (*types.Info, error) {
// 	url := fmt.Sprintf("http://%v:11000/api/status", i.NodeIP)
// 	b, err := httph.MakeHttpRequest(url, "GET")
// 	if err != nil {
// 		return nil, err
// 	}
// 	var info types.Info
// 	err = json.Unmarshal(b, &info)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &info, nil
// }

// func (i InfoScreen) getValoperInfo() (int, error) {}
