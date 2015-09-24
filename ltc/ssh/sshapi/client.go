package sshapi

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var DialFunc = ssh.Dial

//go:generate counterfeiter -o mocks/fake_dialer.go . Dialer
type Dialer interface {
	Dial(n, addr string) (net.Conn, error)
}

//go:generate counterfeiter -o mocks/fake_ssh_session_factory.go . SSHSessionFactory
type SSHSessionFactory interface {
	New() (SSHSession, error)
}

type Client struct {
	Dialer            Dialer
	SSHSessionFactory SSHSessionFactory
	Stdin             io.Reader
	Stdout            io.Writer
	Stderr            io.Writer
}

func New(user, authUser, authPassword, address string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(fmt.Sprintf("%s:%s", authUser, authPassword))},
	}

	client, err := DialFunc("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Dialer:            client,
		SSHSessionFactory: &CryptoSSHSessionFactory{client},
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
	}, nil
}

func (c *Client) Forward(localConn io.ReadWriteCloser, remoteAddress string) error {
	remoteConn, err := c.Dialer.Dial("tcp", remoteAddress)
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

func (c *Client) Open(width, height int, desirePTY bool) (*Session, error) {
	session, err := c.SSHSessionFactory.New()
	if err != nil {
		return nil, err
	}

	sessionIn, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	sessionOut, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	sessionErr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	if desirePTY {
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 115200,
			ssh.TTY_OP_OSPEED: 115200,
		}

		terminalType := os.Getenv("TERM")
		if terminalType == "" {
			terminalType = "xterm"
		}

		if err := session.RequestPty(terminalType, height, width, modes); err != nil {
			return nil, err
		}
	}

	go copyAndClose(nil, sessionIn, c.Stdin)
	go io.Copy(c.Stdout, sessionOut)
	go io.Copy(c.Stderr, sessionErr)

	return &Session{time.NewTicker(30 * time.Second), session, session}, nil
}

func copyAndClose(wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader) {
	io.Copy(dest, src)
	dest.Close()
	if wg != nil {
		wg.Done()
	}
}
