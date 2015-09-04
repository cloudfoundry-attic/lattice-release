package secure_shell

import (
	"github.com/docker/docker/pkg/term"
)

type SecureTerm struct{}

func (t *SecureTerm) SetRawTerminal(fd uintptr) (*term.State, error) {
	return term.SetRawTerminal(fd)
}

func (t *SecureTerm) RestoreTerminal(fd uintptr, state *term.State) error {
	return term.RestoreTerminal(fd, state)
}

func (t *SecureTerm) GetWinsize(fd uintptr) (width int, height int) {
	winSize, err := term.GetWinsize(fd)
	if err != nil {
		return 80, 43
	}

	return int(winSize.Width), int(winSize.Height)
}
