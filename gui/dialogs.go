package gui

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/PeepoFrog/km2UI/gui/dialogs"
	"github.com/PeepoFrog/km2UI/helper/gssh"
	"golang.org/x/crypto/ssh"
)

func (g *Gui) showConnect() {

	var wizard *dialogWizard.Wizard

	//join to new host tab
	joinToNewHost := func() *fyne.Container {
		userEntry := widget.NewEntry()
		ipEntry := widget.NewEntry()
		passwordEntry := widget.NewPasswordEntry()
		errorLabel := widget.NewLabel("")
		var privKeyState bool

		keyPathEntry := widget.NewEntry()
		keyPathEntry.PlaceHolder = "path to your private key"

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			uri := reader.URI().Path()
			keyPathEntry.SetText(uri)
			log.Println("Opened file: ", uri)
		}, g.Window)

		openFileDialogButton := widget.NewButtonWithIcon("", theme.FileIcon(), func() { fileDialog.Show() })
		privKeyEntry := container.NewBorder(
			widget.NewLabel("Select private key"),
			nil, nil,
			openFileDialogButton,
			keyPathEntry,
		)
		passwordBoxEntry := container.NewVBox(
			widget.NewLabel("password"),
			passwordEntry,
		)
		keyEntryBox := container.NewStack(passwordBoxEntry)

		privKeyCheck := widget.NewCheck("Join with private key", func(b bool) {
			privKeyState = b
			if b {
				keyEntryBox.Objects[0] = privKeyEntry
			} else {
				keyEntryBox.Objects[0] = passwordBoxEntry
			}
		})
		errorLabel.Wrapping = 2
		submitFunc := func() {
			var err error
			if privKeyState {
				var b []byte
				g.sshClient, err = func() (*ssh.Client, error) {
					b, err = os.ReadFile(keyPathEntry.Text)
					if err != nil {
						return nil, err
					}
					c, err := gssh.MakeSSH_ClientWithPrivKey(ipEntry.Text, userEntry.Text, b)
					if err != nil {
						return nil, err
					}
					return c, nil
				}()
			} else {
				g.sshClient, err = gssh.MakeSHH_ClientWithPassword(ipEntry.Text, userEntry.Text, passwordEntry.Text)
			}
			if err != nil {
				errorLabel.SetText(fmt.Sprintf("ERROR: %s", err.Error()))
			} else {
				// err = TryToRunSSHSessionForTerminal(g.sshClient)
				// if err != nil {
				// } else {
				wizard.Hide()
			}

		}
		ipEntry.OnSubmitted = func(s string) { submitFunc() }
		userEntry.OnSubmitted = func(s string) { submitFunc() }
		passwordEntry.OnSubmitted = func(s string) { submitFunc() }
		connectButton := widget.NewButton("connect to remote host", func() { submitFunc() })

		logging := container.NewVBox(
			widget.NewLabel("ip and port"),
			ipEntry,
			widget.NewLabel("user"),
			userEntry,
			keyEntryBox,
			connectButton,
			privKeyCheck,
			errorLabel,
		)
		return logging
	}
	mainDialogScreen := container.NewAppTabs(
		// container.NewTabItem("Existing Node", joinToInitializedNode()),
		container.NewTabItem("New Host", joinToNewHost()),
	)
	wizard = dialogWizard.NewWizard("Create ssh connection", mainDialogScreen)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(300, 200))
}
