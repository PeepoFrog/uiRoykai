package gui

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
		addresBoxEntry := container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewLabel(":"), portEntry), ipEntry)

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
				g.Host = &Host{
					IP: ip,
				}
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
			widget.NewLabel("Ip and port"),
			// ipEntry,
			addresBoxEntry,
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

func (g *Gui) showErrorDialog(err error, closeListener binding.DataListener) {
	var wizard *dialogWizard.Wizard
	mainDialogScreen := container.NewVBox(
		widget.NewLabel(err.Error()),
		widget.NewButton("Close", func() { wizard.Hide(); closeListener.DataChanged() }),
	)
	wizard = dialogWizard.NewWizard("Create ssh connection", mainDialogScreen)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(300, 200))

}

func showInfoDialog(g *Gui, infoTitle, infoString string) {
	var wizard *dialogWizard.Wizard
	closeButton := widget.NewButton("Close", func() { wizard.Hide() })
	infoLabel := widget.NewLabel(infoString)
	infoLabel.Wrapping = 2
	content := container.NewBorder(nil, closeButton, nil, nil,
		container.NewHScroll(
			container.NewVScroll(
				infoLabel,
			),
		),
	)

	wizard = dialogWizard.NewWizard(infoTitle, content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 400))
}

func showCmdExecDialogAndRunCmdV4(g *Gui, infoMSG string, cmd string, autoHideCheck bool, errorBinding binding.Bool, errorMessageBinding binding.String) {
	outputChannel := make(chan string)
	errorChannel := make(chan gssh.ResultV2)
	go gssh.ExecuteSSHCommandV2(g.sshClient, cmd, outputChannel, errorChannel)

	var wizard *dialogWizard.Wizard
	outputMsg := binding.NewString()
	statusMsg := binding.NewString()
	statusMsg.Set("Loading...")
	loadingWidget := widget.NewProgressBarInfinite()

	label := widget.NewLabelWithData(outputMsg)
	closeButton := widget.NewButton("Done", func() { wizard.Hide() })
	outputScroll := container.NewVScroll(label)
	loadingDialog := container.NewBorder(
		widget.NewLabelWithData(statusMsg),
		container.NewVBox(loadingWidget, closeButton),
		nil,
		nil,
		container.NewHScroll(outputScroll),
	)
	closeButton.Hide()
	wizard = dialogWizard.NewWizard(infoMSG, loadingDialog)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(300, 400))
	if autoHideCheck {
		defer wizard.Hide()
	}
	var out string
	for line := range outputChannel {
		cleanLine := cleanString(line)
		out = fmt.Sprintf("%s\n%s", out, cleanLine)
		outputMsg.Set(out)
		outputScroll.ScrollToBottom()
	}
	outputScroll.ScrollToBottom()
	loadingWidget.Hide()
	closeButton.Show()
	errcheck := <-errorChannel
	if errcheck.Err != nil {
		log.Printf("Unable to execute executing: <%v>, error: %v", cmd, errcheck.Err.Error())
		errorBinding.Set(true)
		errorMessageBinding.Set(errcheck.Err.Error())
		statusMsg.Set(fmt.Sprintf("Error:\n%s", errcheck.Err))
	} else {
		errorBinding.Set(false)
		wizard.ChangeTitle("Done")
		statusMsg.Set("Successes")
	}

}

func cleanString(s string) string {
	re := regexp.MustCompile("[^\x20-\x7E\n]+")
	return re.ReplaceAllString(s, "")
}

func showSudoEnteringDialog(g *Gui, bindString binding.String, bindCheck binding.Bool) {
	var wizard *dialogWizard.Wizard

	sudoPasswordEntry := widget.NewEntryWithData(bindString)
	errorMessageBinding := binding.NewString()
	checkSudoPassword := func(p string) error {
		cmd := fmt.Sprintf("echo '%v' | sudo -S uname", p)
		// outputChannel := make(chan string)
		// errorChannel := make(chan gssh.ResultV2)
		// go gssh.ExecuteSSHCommandV2(g.sshClient, cmd, outputChannel, errorChannel)
		// for line := range outputChannel {
		// 	log.Println(line)
		// }
		// errcheck := <-errorChannel
		// if errcheck.Err != nil {
		// 	log.Println(errcheck.Err)
		// 	return errcheck.Err
		// }
		errB := binding.NewBool()
		showCmdExecDialogAndRunCmdV4(g, "checking sudo password", cmd, true, errB, errorMessageBinding)
		errExec, _ := errB.Get()
		if errExec {
			errMsg, err := errorMessageBinding.Get()
			if err != nil {
				return err
			}
			return fmt.Errorf("error while checking the sudo password: %v ", errMsg)
		}
		return nil
	}

	okButton := widget.NewButton("Ok", func() {
		err := checkSudoPassword(sudoPasswordEntry.Text)
		if err == nil {
			err = bindCheck.Set(true)
			if err != nil {
				return
			}
			wizard.Hide()
		} else {
			bindCheck.Set(false)
			sudoPasswordEntry.SetValidationError(fmt.Errorf("sudo password is wrong: %w", err))
			showInfoDialog(g, "ERROR", fmt.Sprintf("error when checking sudo password: %v", err.Error()))
		}

	})
	cancelButton := widget.NewButton("Cancel", func() { wizard.Hide() })
	content := container.NewVBox(
		sudoPasswordEntry,
		container.NewHBox(
			okButton, container.NewCenter(), cancelButton,
		),
	)

	wizard = dialogWizard.NewWizard("Enter your sudo password", content)
	wizard.Show(g.Window)

}

func showDeployDialog(g *Gui, doneListener binding.DataListener) {
	var wizard *dialogWizard.Wizard

	ipToJoinEntry := widget.NewEntry()
	interxPortToJoinEntry := widget.NewEntry()
	interxPortToJoinEntry.SetPlaceHolder("11000")

	sekaiRPCPortToJoinEntry := widget.NewEntry()
	sekaiRPCPortToJoinEntry.SetPlaceHolder("26657")

	sekaiP2PPortEntry := widget.NewEntry()
	sekaiP2PPortEntry.SetPlaceHolder("26656")

	sudoPasswordBinding := binding.NewString()
	sudoCheck := binding.NewBool()
	sudoPasswordEntryButton := widget.NewButton("sudo password", func() {
		showSudoEnteringDialog(g, sudoPasswordBinding, sudoCheck)

	})

	deployErrorBinding := binding.NewBool()
	errorMessageBinding := binding.NewString()
	deployButton := widget.NewButton("Deploy", func() {
		sP, err := sudoPasswordBinding.Get()
		cmdForDeploy := fmt.Sprintf(`echo '%v' | sudo -S sh -c "$(curl -s https://raw.githubusercontent.com/KiraCore/sekin/main/scripts/bootstrap.sh 2>&1)"`, sP)
		// cmdForDeploy := fmt.Sprintf(`echo %v`, sP)
		log.Println(interxPortToJoinEntry.PlaceHolder, sekaiRPCPortToJoinEntry.PlaceHolder, sekaiP2PPortEntry.PlaceHolder)
		if err != nil {
			dialog.ShowError(err, g.Window)
		}

		showCmdExecDialogAndRunCmdV4(g, "Deploying", cmdForDeploy, false, deployErrorBinding, errorMessageBinding)

		errB, err := deployErrorBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
		}
		if errB {
			errMsg, err := errorMessageBinding.Get()
			if err != nil {
				g.showErrorDialog(err, binding.NewDataListener(func() {}))
			}

			g.showErrorDialog(fmt.Errorf("error while checking the sudo password: %v ", errMsg), binding.NewDataListener(func() {}))
		} else {
			doneListener.DataChanged()
			wizard.Hide()
		}
		// showInfoDialog(g, "Done", "Finished")

	})

	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})
	sudoCheck.AddListener(binding.NewDataListener(func() {
		check, err := sudoCheck.Get()
		if err != nil {
			return
		}
		if check {
			deployButton.Enable()
		} else {
			if !deployButton.Disabled() {
				deployButton.Disable()
			}
		}

	}))

	deployButton.Disable()

	content := container.NewVBox(
		widget.NewLabel("IP to join"),
		ipToJoinEntry,
		widget.NewLabel("sekai rpc port to join"),
		sekaiRPCPortToJoinEntry,
		widget.NewLabel("sekai P2P port to join"),
		sekaiP2PPortEntry,
		widget.NewLabel("interx port to join"),
		interxPortToJoinEntry,
		sudoPasswordEntryButton,
		deployButton,
		closeButton,
	)

	wizard = dialogWizard.NewWizard("Enter connection info", content)
	wizard.Show(g.Window)
}

type WaitDialog struct {
	wizard *dialogWizard.Wizard
	g      *Gui
}

func NewWaitDialog(g *Gui) *WaitDialog {
	var wizard *dialogWizard.Wizard

	loadingWidget := widget.NewProgressBarInfinite()
	content := container.NewVBox(loadingWidget)

	wizard = dialogWizard.NewWizard("Loading", content)
	return &WaitDialog{
		wizard: wizard,
		g:      g,
	}
}

func (w *WaitDialog) ShowWaitDialog() {
	w.wizard.Show(w.g.Window)
}

func (w *WaitDialog) HideWaitDialog() {
	w.wizard.Hide()
}
