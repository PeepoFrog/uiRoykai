package gui

import (
	"context"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/PeepoFrog/km2UI/helper/networkparser"
)

func makeNetworkTreeScreen(_ fyne.Window, g *Gui) fyne.CanvasObject {

	var nodes = make(map[string]networkparser.Node)
	nodes["node1"] = networkparser.Node{IP: "node2", ID: "nod2", Peers: []networkparser.Node{networkparser.Node{IP: "node1", ID: "nod1"}}}
	nodes["node2"] = networkparser.Node{IP: "node1", ID: "nod1", Peers: []networkparser.Node{networkparser.Node{IP: "node2", ID: "nod2"}}}
	var err error

	data := make([]networkparser.Node, len(nodes))
	i := 0
	for _, v := range nodes {
		data[i] = v
		i++
	}

	infoData := make([]networkparser.Node, 0)

	infoList := widget.NewList(
		func() int {
			return len(infoData)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template Object"))

		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("%v@%v", infoData[id].ID, infoData[id].IP))
		},
	)

	list := widget.NewList(
		func() int {
			return len(nodes)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("%v@%v", data[id].ID, data[id].IP))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		infoData = data[id].Peers
		infoList.Refresh()
	}

	refreshButton := widget.NewButton("Refresh", func() {
		g.WaitDialog.ShowWaitDialog()
		nodes, err = networkparser.GetAllNodesV3(context.Background(), "148.251.69.56", 3, false)
		if err != nil {
			log.Println(err)
			return
		}

		for k, v := range nodes {
			log.Printf("%v %v  %v", k, v.IP, v.ID)
		}
		data = make([]networkparser.Node, len(nodes))
		i := 0
		for _, v := range nodes {
			data[i] = v
			// fmt.Println(data[i].ID, k)
			i++
		}
		list.Refresh()
		g.WaitDialog.HideWaitDialog()
	})

	return container.NewBorder(
		nil, refreshButton, nil, nil, container.NewHSplit(list, infoList),
	)

}
