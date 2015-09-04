package secure_shell

import (
	"fmt"
	"io"
	"os"
	"sync"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	term_package "github.com/docker/docker/pkg/term"
	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fake_secure_shell/fake_secure_shell.go . SecureShell
type SecureShell interface {
	ConnectToShell(appName string, instanceIndex int, config *config_package.Config) error
}

//go:generate counterfeiter -o fake_secure_dialer/fake_secure_dialer.go . SecureDialer
type SecureDialer interface {
	Dial(network, addr string, config *ssh.ClientConfig) (SecureClient, error)
}

//go:generate counterfeiter -o fake_secure_term/fake_secure_term.go . SecureTerm
type SecureTerm interface {
	SetRawTerminal(fd uintptr) (*term_package.State, error)
	RestoreTerminal(fd uintptr, state *term_package.State) error
}

//go:generate counterfeiter -o fake_secure_client/fake_secure_client.go . SecureClient
type SecureClient interface {
	NewSession() (SecureSession, error)
}

//go:generate counterfeiter -o fake_secure_session/fake_secure_session.go . SecureSession
type SecureSession interface {
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	RequestPty(term string, h, w int, termmodes ssh.TerminalModes) error
	Shell() error
	Wait() error
	Close() error
}

type secureShell struct {
	config *config_package.Config
	dialer SecureDialer
	term   SecureTerm
}

type secureSession struct {
	session *ssh.Session
}

func NewSecureShell(config *config_package.Config, dialer SecureDialer, term SecureTerm) SecureShell {
	return &secureShell{
		config: config,
		dialer: dialer,
		term:   term,
	}
}

func (ss *secureShell) ConnectToShell(appName string, instanceIndex int, config *config_package.Config) error {
	secureConfig := &ssh.ClientConfig{
		User: fmt.Sprintf("diego:%s/%d", appName, instanceIndex),
		Auth: []ssh.AuthMethod{ssh.Password(fmt.Sprintf("%s:%s", config.Username(), config.Password()))},
	}

	client, err := ss.dialer.Dial("tcp", fmt.Sprintf("%s:2222", config.Target()), secureConfig)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	sessionIn, err := session.StdinPipe()
	if err != nil {
		return err
	}

	sessionOut, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	sessionErr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 115200,
		ssh.TTY_OP_OSPEED: 115200,
	}

	err = session.RequestPty("xterm", 80, 60, modes)
	if err != nil {
		return err
	}

	if state, err := ss.term.SetRawTerminal(os.Stdin.Fd()); err == nil {
		defer ss.term.RestoreTerminal(os.Stdin.Fd(), state)
	}

	go copyAndClose(nil, sessionIn, os.Stdin)
	go io.Copy(os.Stdout, sessionOut)
	go io.Copy(os.Stderr, sessionErr)

	session.Shell()
	session.Wait()

	return nil
}

func copyAndClose(wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader) {
	io.Copy(dest, src)
	dest.Close()
	if wg != nil {
		wg.Done()
	}
}
