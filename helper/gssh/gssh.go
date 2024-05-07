package gssh

import (
	"bufio"
	"context"
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

type ResultV2 struct {
	Err error
}

func MakeSHH_ClientWithPassword(ipAndPort, user, psswrd string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(psswrd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SSH server
	client, err := ssh.Dial("tcp", ipAndPort, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func MakeSSH_ClientWithPrivKey(ipAndPort, user string, key []byte) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", ipAndPort, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func MakeSSH_ClientWithPrivKeyAndPassphrase(ipAndPort, user string, key, passphrase []byte) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, passphrase)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", ipAndPort, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func CheckIfPassphraseNeeded(privateKeyBytes []byte) (bool, error) {
	_, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// func RunSudoCmd(sshClient *ssh.Client) ([]byte, error) {

// }

func ExecuteSSHCommandV2(client *ssh.Client, command string, outputChan chan<- string, resultChan chan<- ResultV2) {
	log.Printf("RUNNING CMD:\n%s", command)
	session, err := client.NewSession()
	if err != nil {
		resultChan <- ResultV2{Err: err}
		return
	}
	defer session.Close()

	// Setting up stdout and stderr pipes
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		resultChan <- ResultV2{Err: err}
		return
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		resultChan <- ResultV2{Err: err}
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This will be called if an early return occurs
	var wg sync.WaitGroup
	wg.Add(2)
	// Start the command
	err = session.Start(command)
	if err != nil {
		cancel()
		close(outputChan) // Close the channel on error
		resultChan <- ResultV2{Err: err}
		return
	}

	// Read from stdout and stderr concurrently
	go streamOutput(ctx, stdoutPipe, outputChan, &wg)
	go streamOutput(ctx, stderrPipe, outputChan, &wg)

	err = session.Wait()
	cancel()
	wg.Wait()
	close(outputChan) // Close the channel when done
	if err != nil {
		resultChan <- ResultV2{Err: err}
		return
	}

	resultChan <- ResultV2{Err: err}
}

func streamOutput(ctx context.Context, reader io.Reader, outputChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for {
		select {
		case <-ctx.Done():
			return // Exit if context is cancelled
		default:
			if scanner.Scan() {
				outputChan <- scanner.Text()
			} else {
				return // Exit if there's nothing more to read
			}
		}
	}
}
