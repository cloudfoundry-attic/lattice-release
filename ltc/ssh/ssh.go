package ssh

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/docker/docker/pkg/term"
)

//go:generate counterfeiter -o mocks/fake_listener.go . Listener
type Listener interface {
	Listen(network, laddr string) (<-chan io.ReadWriteCloser, <-chan error)
}

//go:generate counterfeiter -o mocks/fake_client_dialer.go . ClientDialer
type ClientDialer interface {
	Dial(appName string, instanceIndex int, config *config_package.Config) (Client, error)
}

//go:generate counterfeiter -o mocks/fake_term.go . Term
type Term interface {
	SetRawTerminal(fd uintptr) (*term.State, error)
	RestoreTerminal(fd uintptr, state *term.State) error
	GetWinsize(fd uintptr) (width int, height int)
}

//go:generate counterfeiter -o mocks/fake_session_factory.go . SessionFactory
type SessionFactory interface {
	New(client Client, width, height int) (Session, error)
}

type SSH struct {
	Listener       Listener
	ClientDialer   ClientDialer
	Term           Term
	SessionFactory SessionFactory
	client         Client
}

func New() *SSH {
	return &SSH{
		Listener:       &ChannelListener{},
		ClientDialer:   &AppDialer{},
		Term:           &DockerTerm{},
		SessionFactory: &SSHAPISessionFactory{},
	}
}

func (s *SSH) Connect(appName string, instanceIndex int, config *config_package.Config) error {
	if s.client != nil {
		return errors.New("already connected")
	}
	var err error
	s.client, err = s.ClientDialer.Dial(appName, instanceIndex, config)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSH) Forward(localAddress, remoteAddress string) error {
	acceptChan, errorChan := s.Listener.Listen("tcp", localAddress)

	for {
		select {
		case conn, ok := <-acceptChan:
			if !ok {
				return nil
			}

			if err := s.client.Forward(conn, remoteAddress); err != nil {
				panic(err)
			}
		case err, ok := <-errorChan:
			if !ok {
				return nil
			}

			panic(err)
		}
	}
}

func (s *SSH) Shell(command string) error {
	width, height := s.Term.GetWinsize(os.Stdout.Fd())
	session, err := s.SessionFactory.New(s.client, width, height)
	if err != nil {
		return err
	}
	defer session.Close()

	if state, err := s.Term.SetRawTerminal(os.Stdin.Fd()); err == nil {
		defer s.Term.RestoreTerminal(os.Stdin.Fd(), state)
	}

	resized := make(chan os.Signal, 16)
	signal.Notify(resized, syscall.SIGWINCH)
	defer func() {
		signal.Stop(resized)
		close(resized)
	}()
	go s.resize(resized, session, os.Stdout.Fd(), width, height)

	defer close(session.KeepAlive())

	if command == "" {
		session.Shell()
		session.Wait()
	} else {
		session.Run(command)
	}

	return nil
}

func (s *SSH) resize(resized <-chan os.Signal, session Session, terminalFd uintptr, initialWidth, initialHeight int) {
	previousWidth := initialWidth
	previousHeight := initialHeight

	for range resized {
		width, height := s.Term.GetWinsize(terminalFd)

		if width == previousWidth && height == previousHeight {
			continue
		}

		session.Resize(width, height)

		previousWidth = width
		previousHeight = height
	}
}
