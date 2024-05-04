package gui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/PeepoFrog/km2UI/gui/dialogs"
	"github.com/PeepoFrog/km2UI/helper/gssh"
	"golang.org/x/crypto/ssh"
)

func (g *Gui) ShowConnect() {

	var wizard *dialogWizard.Wizard

	//join to new host tab
	joinToNewHost := func() *fyne.Container {
		userEntry := widget.NewEntry()
		ipEntry := widget.NewEntry()
		portEntry := widget.NewEntry()
		passwordEntry := widget.NewPasswordEntry()
		errorLabel := widget.NewLabel("")
		keyPathEntry := widget.NewEntry()
		var privKeyState bool
		portEntry.PlaceHolder = "22"
		addresBoxEntry := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel(":"), portEntry), ipEntry)

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
			widget.NewLabel("Password"),
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
			ip := strings.TrimSpace(ipEntry.Text)
			port := ""
			if portEntry.Text == "" {
				port = "22"
			} else {
				port = strings.TrimSpace(portEntry.Text)
			}
			address := fmt.Sprintf("%v:%v", ip, (port))
			log.Print(address)
			if privKeyState {
				var b []byte
				g.sshClient, err = func() (*ssh.Client, error) {
					b, err = os.ReadFile(keyPathEntry.Text)
					if err != nil {
						return nil, err
					}

					c, err := gssh.MakeSSH_ClientWithPrivKey(address, userEntry.Text, b)
					if err != nil {
						return nil, err
					}
					return c, nil
				}()
			} else {
				g.sshClient, err = gssh.MakeSHH_ClientWithPassword(address, userEntry.Text, passwordEntry.Text)
			}
			if err != nil {
				errorLabel.SetText(fmt.Sprintf("ERROR: %s", err.Error()))
			} else {
				// err = TryToRunSSHSessionForTerminal(g.sshClient)
				// if err != nil {
				// } else {
				g.Host = &Host{
					IP: ip,
				}
				wizard.Hide()
			}

		}
		ipEntry.OnSubmitted = func(s string) { submitFunc() }
		userEntry.OnSubmitted = func(s string) { submitFunc() }
		passwordEntry.OnSubmitted = func(s string) { submitFunc() }
		connectButton := widget.NewButton("Connect to remote host", func() { submitFunc() })

		logging := container.NewVBox(
			widget.NewLabel("Ip and port"),
			// ipEntry,
			addresBoxEntry,
			widget.NewLabel("User"),
			userEntry,
			keyEntryBox,
			connectButton,
			privKeyCheck,
			errorLabel,
		)
		return logging
	}
	// mainDialogScreen := container.NewAppTabs(
	// 	// container.NewTabItem("Existing Node", joinToInitializedNode()),
	// 	container.NewTabItem("New Host", joinToNewHost()),
	// )
	wizard = dialogWizard.NewWizard("Create ssh connection", joinToNewHost())
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(350, 450))
}
