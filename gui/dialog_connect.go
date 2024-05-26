package gui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	dialogWizard "github.com/PeepoFrog/km2UI/gui/dialogs"
	"github.com/PeepoFrog/km2UI/helper/gssh"
	"github.com/fyne-io/terminal"
	"golang.org/x/crypto/ssh"
)

func (g *Gui) ShowConnect() {

	var wizard *dialogWizard.Wizard

	//join to new host tab
	join := func() *fyne.Container {
		userEntry := widget.NewEntry()
		ipEntry := widget.NewEntry()
		portEntry := widget.NewEntry()
		passwordEntry := widget.NewPasswordEntry()
		errorLabel := widget.NewLabel("")
		keyPathEntry := widget.NewEntry()
		passphraseEntry := widget.NewPasswordEntry()
		passphraseEntry.Hide()
		var privKeyState bool
		var passphraseState bool
		portEntry.PlaceHolder = "22"
		passphraseEntry.Validator = func(s string) error {
			if s == "" {
				return fmt.Errorf("enter your passphrase")
			}
			return nil
		}
		addressBoxEntry := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel(":"), portEntry), ipEntry)

		keyPathEntry.PlaceHolder = "path to your private key"
		passphraseEntry.PlaceHolder = "your passphrase"
		passphraseCheck := widget.NewCheck("SSH passphrase key", func(b bool) {
			passphraseState = b
			if passphraseState {
				passphraseEntry.Show()
			} else {
				passphraseEntry.Hide()
			}
		})

		keyPathEntry.OnChanged = func(s string) {
			b, err := os.ReadFile(s)
			if err != nil {
				return
			}
			check, err := gssh.CheckIfPassphraseNeeded(b)
			if err != nil {
				return
			}
			if check {
				passphraseCheck.SetChecked(true)
			} else {
				passphraseCheck.SetChecked(false)
			}
			log.Println(s)
		}

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader == nil {
				return
			}

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

		privKeyBoxEntry := container.NewVBox(
			privKeyEntry,
			passphraseEntry,
		)

		passwordBoxEntry := container.NewVBox(
			widget.NewLabel("Password"),
			passwordEntry,
		)
		keyEntryBox := container.NewStack(passwordBoxEntry)

		privKeyCheck := widget.NewCheck("Join with private key", func(b bool) {
			privKeyState = b
			if b {
				keyEntryBox.Objects = []fyne.CanvasObject{privKeyBoxEntry}
			} else {
				keyEntryBox.Objects = []fyne.CanvasObject{passwordBoxEntry}
			}
		})

		privKeyBoxEntry.Objects = append(privKeyBoxEntry.Objects, passphraseCheck)

		errorLabel.Wrapping = 2

		submitFunc := func() {
			g.WaitDialog.ShowWaitDialog()
			var err error
			ip := strings.TrimSpace(ipEntry.Text)
			port := ""
			if portEntry.Text == "" {
				port = "22"
			} else {
				port = strings.TrimSpace(portEntry.Text)
			}
			address := fmt.Sprintf("%v:%v", ip, (port))

			if privKeyState {
				var b []byte
				var c *ssh.Client

				g.sshClient, err = func() (*ssh.Client, error) {
					b, err = os.ReadFile(keyPathEntry.Text)
					if err != nil {
						return nil, err
					}
					check, err := gssh.CheckIfPassphraseNeeded(b)
					if err != nil {
						return nil, err
					}
					if check {
						if passphraseEntry.Hidden {
							passphraseEntry.Validate()
							passphraseEntry.SetValidationError(fmt.Errorf("passphrase required"))
							passphraseCheck.SetChecked(true)
						}

						c, err = gssh.MakeSSH_ClientWithPrivKeyAndPassphrase(address, userEntry.Text, b, []byte(passphraseEntry.Text))
						if err != nil {

							return nil, err
						}
					} else {
						if !passphraseEntry.Hidden {
							passphraseCheck.SetChecked(false)

						}
						c, err = gssh.MakeSSH_ClientWithPrivKey(address, userEntry.Text, b)
						if err != nil {
							return nil, err
						}
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

				err := TryToRunSSHSessionForTerminal(g)
				if err != nil {
					g.showErrorDialog(fmt.Errorf("unable to create terminal instance, disabling terminal: %v", err.Error()), binding.NewDataListener(func() {}))
					g.Terminal.Term = terminal.New()

				}
				g.Host = &Host{
					IP: ip,
				}
				go g.sshAliveTracker()
				g.ConnectionStatusBinding.Set(true)
				wizard.Hide()
			}
			defer g.WaitDialog.HideWaitDialog()
		}

		// / test ui block
		testButton := widget.NewButton("connect to tested env", func() {
			ipEntry.Text = "192.168.1.101"
			userEntry.Text = "d"
			passwordEntry.Text = "d"
			passphraseCheck.SetChecked(false)

			submitFunc()
		})

		///

		ipEntry.OnSubmitted = func(s string) { submitFunc() }
		userEntry.OnSubmitted = func(s string) { submitFunc() }
		passwordEntry.OnSubmitted = func(s string) { submitFunc() }
		connectButton := widget.NewButton("Connect to remote host", func() { submitFunc() })

		logging := container.NewVBox(
			widget.NewLabel("IP and Port"),
			// ipEntry,
			addressBoxEntry,
			widget.NewLabel("User"),
			userEntry,
			keyEntryBox,
			privKeyCheck,
			connectButton,
			errorLabel,
			testButton,
		)
		return logging
	}

	wizard = dialogWizard.NewWizard("Create ssh connection", join())
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(350, 450))
}

func (g *Gui) sshAliveTracker() {
	errorDoneBinding := binding.NewDataListener(func() {
		// g.ShowConnect()
	})
	g.ConnectionCount++

	err := g.sshClient.Wait()
	if err != nil {
		log.Printf("SSH was interrupted: %v", err.Error())
		g.ConnectionStatusBinding.Set(false)
		// g.Terminal.Term.Exit()
		// g.Terminal.SSHSessionForTerminal.Close()
		// g.Terminal.SSHIn.Close()

		g.showErrorDialog(fmt.Errorf("SSH connection was disconnected, reason: %v", err.Error()), errorDoneBinding)
	}

}
