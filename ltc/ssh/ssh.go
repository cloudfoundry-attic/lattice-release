package ssh

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
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
	IsTTY(fd uintptr) bool
}

//go:generate counterfeiter -o mocks/fake_session_factory.go . SessionFactory
type SessionFactory interface {
	New(client Client, width, height int, desirePTY bool) (Session, error)
}

type SSH struct {
	Listener        Listener
	ClientDialer    ClientDialer
	Term            Term
	SessionFactory  SessionFactory
	SigWinchChannel chan os.Signal
	ExitHandler     exit_handler.ExitHandler
	client          Client
}

func New(exitHandler exit_handler.ExitHandler) *SSH {
	return &SSH{
		Listener:        &ChannelListener{},
		ClientDialer:    &AppDialer{},
		Term:            &DockerTerm{},
		SessionFactory:  &SSHAPISessionFactory{},
		SigWinchChannel: make(chan os.Signal),
		ExitHandler:     exitHandler,
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
				return err
			}
		case err, ok := <-errorChan:
			if !ok {
				return nil
			}

			return err
		}
	}
}

func (s *SSH) Shell(command string, desirePTY bool) error {
	if desirePTY {
		desirePTY = s.Term.IsTTY(os.Stdin.Fd())
	}

	width, height := s.Term.GetWinsize(os.Stdout.Fd())
	session, err := s.SessionFactory.New(s.client, width, height, desirePTY)
	if err != nil {
		return err
	}
	defer session.Close()

	if desirePTY {
		if state, err := s.Term.SetRawTerminal(os.Stdin.Fd()); err == nil {
			defer s.Term.RestoreTerminal(os.Stdin.Fd(), state)

			s.ExitHandler.OnExit(func() {
				s.Term.RestoreTerminal(os.Stdin.Fd(), state)
			})
		}
	}

	signal.Notify(s.SigWinchChannel, syscall.SIGWINCH)
	defer func() {
		signal.Stop(s.SigWinchChannel)
		close(s.SigWinchChannel)
	}()
	go s.resize(session, os.Stdout.Fd(), width, height)

	defer close(session.KeepAlive())

	if command == "" {
		session.Shell()
		session.Wait()
	} else {
		session.Run(command)
	}

	return nil
}

func (s *SSH) resize(session Session, terminalFd uintptr, initialWidth, initialHeight int) {
	previousWidth := initialWidth
	previousHeight := initialHeight

	for range s.SigWinchChannel {
		width, height := s.Term.GetWinsize(terminalFd)

		if width == previousWidth && height == previousHeight {
			continue
		}

		session.Resize(width, height)

		previousWidth = width
		previousHeight = height
	}
}
