package gssh

import "golang.org/x/crypto/ssh"

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
