package sshapi

import (
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

type CryptoSSHSessionFactory struct {
	Client *ssh.Client
}

func (c *CryptoSSHSessionFactory) New() (SSHSession, error) {
	return c.Client.NewSession()
}

//go:generate counterfeiter -o mocks/fake_ssh_session.go . SSHSession
type SSHSession interface {
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	SendRequest(name string, wantReply bool, payload []byte) (bool, error)
	RequestPty(term string, h, w int, termmodes ssh.TerminalModes) error
	exposedSession
}

type exposedSession interface {
	Shell() error
	Run(string) error
	Wait() error
	Close() error
}

type Session struct {
	KeepAliveTicker *time.Ticker
	sshSession      SSHSession
	exposedSession
}

func (s *Session) KeepAlive() (stopChan chan<- struct{}) {
	receiveStopChan := make(chan struct{})

	go func() {
		for {
			select {
			case <-s.KeepAliveTicker.C:
				s.sshSession.SendRequest("keepalive@cloudfoundry.org", true, nil)
			case <-receiveStopChan:
				s.KeepAliveTicker.Stop()
				return
			}
		}
	}()

	return receiveStopChan
}

func (s *Session) Resize(width, height int) error {
	message := struct {
		Width       uint32
		Height      uint32
		PixelWidth  uint32
		PixelHeight uint32
	}{uint32(width), uint32(height), 0, 0}

	_, err := s.sshSession.SendRequest("window-change", false, ssh.Marshal(message))
	return err
}
