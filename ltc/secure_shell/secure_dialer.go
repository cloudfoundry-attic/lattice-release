package secure_shell

import (
	"fmt"
	"io"
	"net"
	"sync"

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

type SSHSessionFactory struct {
	Client *ssh.Client
}

func (s *SSHSessionFactory) New() (Session, error) {
	return s.Client.NewSession()
}

//go:generate counterfeiter -o fake_ssh_dialer/fake_ssh_dialer.go . SSHDialer
type SSHDialer interface {
	Dial(n, addr string) (net.Conn, error)
}

//go:generate counterfeiter -o fake_session_factory/fake_session_factory.go . SessionFactory
type SessionFactory interface {
	New() (Session, error)
}

type SecureClient struct {
	Client   SSHDialer
	Sessions SessionFactory
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
}

func (s *SecureClient) Accept(localConn io.ReadWriteCloser, remoteAddress string) error {
	remoteConn, err := s.Client.Dial("tcp", remoteAddress)
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go copyAndClose(wg, localConn, remoteConn)
	go copyAndClose(wg, remoteConn, localConn)
	wg.Wait()

	return nil
}

func copyAndClose(wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader) {
	io.Copy(dest, src) // TODO: test error
	dest.Close()
	if wg != nil {
		wg.Done()
	}
}

func (s *SecureClient) Open(height, width uint32) error {
	return nil
}
