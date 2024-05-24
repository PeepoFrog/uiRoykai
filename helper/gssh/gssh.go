package gssh

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This will be called if an early return occurs
	session, err := client.NewSession()
	if err != nil {
		if err == io.EOF {
			err = fmt.Errorf("ssh EOF, probably ssh server was down, please restart Kensho: %w", err)
		}
		log.Println("Error when creating new session: ", err.Error())
		cancel()
		close(outputChan)
		resultChan <- ResultV2{Err: err}
		return
	}
	defer session.Close()

	// Setting up stdout and stderr pipes
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		log.Println("Error when creating stdoutPipe: ", err.Error())
		resultChan <- ResultV2{Err: err}
		close(outputChan)
		return
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		log.Println("Error when creating stderrPipe: ", err.Error())
		resultChan <- ResultV2{Err: err}
		return
	}

	var wg sync.WaitGroup
	wg.Add(3)
	// Start the command
	err = session.Start(command)
	if err != nil {
		log.Println("Error when starting new session: ", err.Error())
		cancel()
		close(outputChan) // Close the channel on error
		resultChan <- ResultV2{Err: err}
		return
	}

	// Read from stdout and stderr concurrently
	go streamOutput(ctx, stdoutPipe, outputChan, &wg)
	go streamOutput(ctx, stderrPipe, outputChan, &wg)

	go monitorConnection(ctx, client.Conn, outputChan, &wg)
	err = session.Wait()
	if err != nil {
		cancel()
		resultChan <- ResultV2{Err: err}
		return
	}
	cancel()
	wg.Wait()
	close(outputChan) // Close the channel when done

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
				log.Println(scanner.Text)
			} else {
				return // Exit if there's nothing more to read
			}
		}
	}
}

func monitorConnection(ctx context.Context, conn ssh.Conn, outputChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Using the keepalive mechanism to detect if the connection is closed
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _, err := conn.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				outputChan <- "Connection closed by remote host"
				return
			}
		}
	}
}
