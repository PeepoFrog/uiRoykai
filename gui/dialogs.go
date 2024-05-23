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
	"github.com/PeepoFrog/km2UI/helper/httph"
	mnemonicHelper "github.com/PeepoFrog/km2UI/helper/mnemonicHelper"
	"github.com/atotto/clipboard"
	"golang.org/x/crypto/ssh"
)

type WaitDialog struct {
	wizard *dialogWizard.Wizard
	g      *Gui
}

// Wait dialog init
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

	errorLabel := widget.NewLabel(err.Error())
	errorLabel.Wrapping = fyne.TextWrapWord

	mainDialogScreen := container.NewVBox(
		errorLabel,
		widget.NewButton("Copy", func() {
			err = clipboard.WriteAll(errorLabel.Text)
			if err != nil {
				return
			}
		}),
		widget.NewButton("Close", func() { wizard.Hide(); closeListener.DataChanged() }),
	)
	wizard = dialogWizard.NewWizard("Error", mainDialogScreen)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 200))

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
		log.Printf("Unable to execute executing: <%v>, error: %v, %v ", cmd, errcheck.Err.Error(), string(out))
		errorBinding.Set(true)
		errorMessageBinding.Set(fmt.Sprintf("Out: %v, Error: %v", string(out), errcheck.Err.Error()))
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

	localCheckBinding := binding.NewBool()
	localCheck := widget.NewCheckWithData("local", localCheckBinding)

	sudoPasswordBinding := binding.NewString()
	mnemonicBinding := binding.NewString()
	sudoCheck := binding.NewBool()
	mnemonicCheck := binding.NewBool()

	sudoPasswordEntryButton := widget.NewButtonWithIcon("sudo password", theme.CancelIcon(), func() {
		showSudoEnteringDialog(g, sudoPasswordBinding, sudoCheck)
	})

	doneMnemonicDataListener := binding.NewDataListener(func() {
		err := mnemonicCheck.Set(true)
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		confirmedMnemonic, err := mnemonicBinding.Get()
		log.Println("Confirmed mnemonic:", confirmedMnemonic, err)
	})
	mnemonicManagerDialogButton := widget.NewButtonWithIcon("mnemonic", theme.CancelIcon(), func() {
		showMnemonicManagerDialog(g, mnemonicBinding, doneMnemonicDataListener)
	})

	constructJoinCmd := func() (string, error) {
		rpcPort := sekaiRPCPortToJoinEntry.Text
		if rpcPort == "" {
			rpcPort = sekaiRPCPortToJoinEntry.PlaceHolder
		} else {
			validate := httph.ValidatePortRange(rpcPort)
			if !validate {
				sekaiP2PPortEntry.SetValidationError(fmt.Errorf("invalid port"))
				return "", fmt.Errorf("RPC port is not valid")
			}
		}
		p2pPort := sekaiP2PPortEntry.Text
		if p2pPort == "" {
			p2pPort = sekaiP2PPortEntry.PlaceHolder
		} else {
			validate := httph.ValidatePortRange(p2pPort)
			if !validate {
				return "", fmt.Errorf("P2P port is not valid")
			}
		}
		interxPort := interxPortToJoinEntry.Text
		if interxPort == "" {
			interxPort = interxPortToJoinEntry.PlaceHolder
		} else {
			validate := httph.ValidatePortRange(rpcPort)
			if !validate {
				return "", fmt.Errorf("interx port is not valid")
			}
		}

		ip := ipToJoinEntry.Text
		validate := httph.ValidateIP(ip)
		if !validate {
			return "", fmt.Errorf(`ip <%v> is not valid`, ip)
		}

		mnemonic, err := mnemonicBinding.Get()
		if err != nil {
			return "", err
		}

		lCheck, err := localCheckBinding.Get()
		if err != nil {
			return "", err
		}
		cmd := fmt.Sprintf(`curl --silent --show-error --fail -X POST http://localhost:8282/api/execute -H "Content-Type: application/json" -d '{
			"command": "join",
			"args": {
				"ip": "%v",
				"interxPort": %v,
				"rpcPort": %v,
				"p2pPort": %v,
				"mnemonic": "%v",
				"local": %v,
				"enableInterx": %v
			}
		}'`, ip, interxPort, rpcPort, p2pPort, mnemonic, lCheck, true)
		return cmd, nil
	}

	deployErrorBinding := binding.NewBool()
	errorMessageBinding := binding.NewString()

	deployButton := widget.NewButton("Deploy", func() {
		cmdForJoin, err := constructJoinCmd()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}

		sP, err := sudoPasswordBinding.Get()
		if err != nil {
			dialog.ShowError(err, g.Window)
			return
		}
		cmdForDeploy := fmt.Sprintf(`echo '%v' | sudo -S sh -c "$(curl -s --show-error --fail https://raw.githubusercontent.com/KiraCore/sekin/main/scripts/bootstrap.sh 2>&1)"`, sP)
		// cmdForDeploy = fmt.Sprintf(`echo %v`, sP)
		showCmdExecDialogAndRunCmdV4(g, "Deploying", cmdForDeploy, true, deployErrorBinding, errorMessageBinding)

		errB, err := deployErrorBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		if errB {
			errMsg, err := errorMessageBinding.Get()
			if err != nil {
				g.showErrorDialog(err, binding.NewDataListener(func() {}))
				return
			}
			g.showErrorDialog(fmt.Errorf("error while checking the sudo password: %v ", errMsg), binding.NewDataListener(func() {}))
			return
		}

		showCmdExecDialogAndRunCmdV4(g, "Joining", cmdForJoin, false, deployErrorBinding, errorMessageBinding)

		errB, err = deployErrorBinding.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		if errB {
			errMsg, err := errorMessageBinding.Get()
			if err != nil {
				g.showErrorDialog(err, binding.NewDataListener(func() {}))
				return
			}
			g.showErrorDialog(fmt.Errorf("error when executing join command: %v ", errMsg), binding.NewDataListener(func() {}))
			return
		}

		doneListener.DataChanged()
		wizard.Hide()

	})

	deployButton.Disable()

	deployActivatorDataListener := binding.NewDataListener(func() {
		sCheck, err := sudoCheck.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		if sCheck {
			sudoPasswordEntryButton.Icon = theme.ConfirmIcon()
			sudoPasswordEntryButton.Refresh()
		} else {
			sudoPasswordEntryButton.Icon = theme.CancelIcon()
			sudoPasswordEntryButton.Refresh()
		}
		mCheck, err := mnemonicCheck.Get()
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		if mCheck {
			log.Println("changing mnemonicButtonIcon")
			mnemonicManagerDialogButton.Icon = theme.ConfirmIcon()
			mnemonicManagerDialogButton.Refresh()
		} else {
			mnemonicManagerDialogButton.Icon = theme.CancelIcon()
			mnemonicManagerDialogButton.Refresh()
		}

		if sCheck && mCheck {
			deployButton.Enable()
		} else {
			if !deployButton.Disabled() {
				deployButton.Disable()
			}
		}
	})

	closeButton := widget.NewButton("Close", func() {
		wizard.Hide()
	})

	mnemonicCheck.AddListener(deployActivatorDataListener)
	sudoCheck.AddListener(deployActivatorDataListener)

	content := container.NewVBox(
		widget.NewLabel("IP to join"),
		ipToJoinEntry,
		localCheck,
		widget.NewLabel("sekai rpc port to join"),
		sekaiRPCPortToJoinEntry,
		widget.NewLabel("sekai P2P port to join"),
		sekaiP2PPortEntry,
		widget.NewLabel("interx port to join"),
		interxPortToJoinEntry,
		sudoPasswordEntryButton,
		mnemonicManagerDialogButton,
		deployButton,
		closeButton,
	)

	wizard = dialogWizard.NewWizard("Enter connection info", content)
	wizard.Show(g.Window)
	wizard.Resize(fyne.NewSize(400, 500))

}

func showMnemonicManagerDialog(g *Gui, mnemonicBinding binding.String, doneAction binding.DataListener) {
	var wizard *dialogWizard.Wizard
	mnemonicDisplay := container.NewGridWithColumns(2)
	// generatedBinding := binding.NewString()
	warningConfirmDataListener := binding.NewDataListener(func() {
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
		m, err := mnemonicBinding.Get()
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
	// mnemonicChanged.DataChanged()

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
		data, err := mnemonicBinding.Get()
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

		err = mnemonicBinding.Set(masterMnemonic.String())
		if err != nil {
			g.showErrorDialog(err, binding.NewDataListener(func() {}))
			return
		}
		mnemonicChanged.DataChanged()
		log.Println(mnemonicBinding.Get())
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
	infoLabel := widget.NewLabel("")
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
