package secure_shell

import (
	"github.com/docker/docker/pkg/term"
)

type dockerTerm struct{}

func NewSecureTerm() *dockerTerm {
	return &dockerTerm{}
}

func (t *dockerTerm) SetRawTerminal(fd uintptr) (*term.State, error) {
	return term.SetRawTerminal(fd)
}

func (t *dockerTerm) RestoreTerminal(fd uintptr, state *term.State) error {
	return term.RestoreTerminal(fd, state)
}
