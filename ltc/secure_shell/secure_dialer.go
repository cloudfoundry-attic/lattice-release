package secure_shell

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

type SecureDialer struct {
	DialFunc func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
}

func (s *SecureDialer) Dial(user, authUser, authPassword, address string) (Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(fmt.Sprintf("%s:%s", authUser, authPassword))},
	}

	client, err := s.DialFunc("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return &SecureClient{client}, nil
}

type SecureClient struct {
	Client *ssh.Client
}

func (s *SecureClient) NewSession() (SecureSession, error) {
	session, err := s.NewSession()
	return session, err
}

func (s *SecureClient) Dial(n, addr string) (net.Conn, error) {
	conn, err := s.Dial(n, addr)
	return conn, err
}
