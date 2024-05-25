package gui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/PeepoFrog/km2UI/gui/dialogs"
	mnemonicHelper "github.com/PeepoFrog/km2UI/helper/mnemonicHelper"

	"github.com/atotto/clipboard"
)

func showMnemonicManagerDialog(g *Gui, mnemonicBinding binding.String, doneAction binding.DataListener) {
	var wizard *dialogWizard.Wizard
	mnemonicDisplay := container.NewGridWithColumns(2)
	localMnemonicBinding := binding.NewString()
	warningConfirmDataListener := binding.NewDataListener(func() {
		lMnemonic, err := localMnemonicBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		err = mnemonicBinding.Set(lMnemonic)
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		doneAction.DataChanged()
		wizard.Hide()
	})
	doneButton := widget.NewButton("Done", func() {
		showMnemonicWarningMessage(g, warningConfirmDataListener)
	})
	doneButton.Disable()
	var content *fyne.Container

	// doing this to display mnemonic if it was already generated
	m, err := mnemonicBinding.Get()
	if err != nil {
		g.showErrorDialog(err, binding.NewDataListener(func() {}))
		return
	}
	if m != "" {
		mnemonicWords := strings.Split(m, " ")
		mnemonicDisplay.RemoveAll()
		for i, w := range mnemonicWords {
			mnemonicDisplay.Add(widget.NewLabel(fmt.Sprintf("%v. %v", i+1, w)))
		}
	}
	//

	mnemonicChanged := binding.NewDataListener(func() {
		m, err := localMnemonicBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		err = mnemonicHelper.ValidateMnemonic(m)
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		doneButton.Enable()
		mnemonicWords := strings.Split(m, " ")
		mnemonicDisplay.RemoveAll()
		for i, w := range mnemonicWords {
			mnemonicDisplay.Add(widget.NewLabel(fmt.Sprintf("%v. %v", i+1, w)))
		}
		content.Refresh()
	})

	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})

	enterMnemonicManuallyButton := widget.NewButton("Enter your mnemonic", func() {
		doneEnteringMnemonicListener := binding.NewDataListener(func() {
			mnemonicChanged.DataChanged()
		})
		showMnemonicEntryDialog(g, mnemonicBinding, doneEnteringMnemonicListener)
	})

	copyButton := widget.NewButtonWithIcon("Copy", theme.FileIcon(), func() {
		data, err := localMnemonicBinding.Get()
		if err != nil {
			log.Println(err)
			return
		}
		err = clipboard.WriteAll(data)
		if err != nil {
			log.Println(err)
			return
		}
	})

	generateButton := widget.NewButton("Generate", func() {
		masterMnemonic, err := mnemonicHelper.GenerateMnemonic()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		err = mnemonicHelper.ValidateMnemonic(masterMnemonic.String())
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		err = localMnemonicBinding.Set(masterMnemonic.String())
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		mnemonicChanged.DataChanged()
		log.Println(localMnemonicBinding.Get())
	})

	content = container.NewBorder(
		nil,
		container.NewVBox(enterMnemonicManuallyButton, container.NewVBox(container.NewGridWithColumns(2, generateButton, copyButton)), closeButton, doneButton),
		nil,
		nil,
		mnemonicDisplay,
	)

	wizard = dialogWizard.NewWizard("Mnemonic setup", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 700))
}

func showMnemonicEntryDialog(g *Gui, mnemonicBinding binding.String, doneAction binding.DataListener) {
	var wizard *dialogWizard.Wizard
	infoLabel := widget.NewLabel("Enter your mnemonic")
	infoLabel.Wrapping = fyne.TextWrapWord
	mnemonicEntry := widget.NewEntry()
	mnemonicEntry.Wrapping = fyne.TextWrapWord
	mnemonicEntry.MultiLine = true
	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})

	doneButton := widget.NewButton("Done", func() {
		mnemonicBinding.Set(mnemonicEntry.Text)
		doneAction.DataChanged()
		wizard.Hide()
	})
	doneButton.Disable()

	mnemonicEntry.OnChanged = func(s string) {
		err := mnemonicHelper.ValidateMnemonic(mnemonicEntry.Text)
		if err != nil {
			infoLabel.SetText("Mnemonic is not valid")
			doneButton.Disable()
		} else {
			infoLabel.SetText("Mnemonic is valid")
			doneButton.Enable()
		}
	}

	content := container.NewBorder(
		infoLabel,
		container.NewVBox(closeButton, doneButton),
		nil,
		nil,
		(mnemonicEntry),
	)

	wizard = dialogWizard.NewWizard("Mnemonic setup", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(900, 200))
}

func showMnemonicWarningMessage(g *Gui, confirmAction binding.DataListener) {
	var wizard *dialogWizard.Wizard

	warningInfoLabel := widget.NewLabel(`By clicking "Proceed," you confirm that you have saved your mnemonic. You will no longer be able to see your mnemonic a second time. Make sure you have securely stored it before proceeding.
If you have not please press "Return" and save your mnemonic.`)

	warningInfoLabel.Wrapping = fyne.TextWrapWord
	warningInfoLabel.Importance = widget.DangerImportance
	proceedButton := widget.NewButtonWithIcon("Proceed", theme.ConfirmIcon(), func() {
		wizard.Hide()
		confirmAction.DataChanged()
	})
	returnButton := widget.NewButton("Return", func() { wizard.Hide() })

	content := container.NewBorder(
		nil,
		container.NewGridWithColumns(2, proceedButton, returnButton),
		nil,
		nil,
		warningInfoLabel,
	)
	wizard = dialogWizard.NewWizard("WARNING!", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(500, 400))
}
