package secure_shell

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type SecureDialer struct {
	DialFunc func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
}

func (s *SecureDialer) Dial(user, authUser, authPassword, address string) (SecureSession, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(fmt.Sprintf("%s:%s", authUser, authPassword))},
	}

	client, err := s.DialFunc("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return client.NewSession()
}
