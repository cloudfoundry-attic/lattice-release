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

//go:generate counterfeiter -o fake_dialer/fake_dialer.go . Dialer
type Dialer interface {
	Dial(user, authUser, authPassword, address string) (SecureSession, error)
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

//go:generate counterfeiter -o fake_term/fake_term.go . Term
type Term interface {
	SetRawTerminal(fd uintptr) (*term_package.State, error)
	RestoreTerminal(fd uintptr, state *term_package.State) error
	GetWinsize(fd uintptr) (width int, height int)
}

type SecureShell struct {
	Dialer Dialer
	Term   Term
}

func (ss *SecureShell) ConnectToShell(appName string, instanceIndex int, config *config_package.Config) error {
	diegoSSHUser := fmt.Sprintf("diego:%s/%d", appName, instanceIndex)
	address := fmt.Sprintf("%s:2222", config.Target())

	session, err := ss.Dialer.Dial(diegoSSHUser, config.Username(), config.Password(), address)
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

	width, height := ss.Term.GetWinsize(os.Stdout.Fd())

	terminalType := os.Getenv("TERM")
	if terminalType == "" {
		terminalType = "xterm"
	}

	err = session.RequestPty(terminalType, height, width, modes)
	if err != nil {
		return err
	}

	if state, err := ss.Term.SetRawTerminal(os.Stdin.Fd()); err == nil {
		defer ss.Term.RestoreTerminal(os.Stdin.Fd(), state)
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
