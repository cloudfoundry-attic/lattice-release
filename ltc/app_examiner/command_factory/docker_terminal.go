package command_factory

import (
	"os"

	"github.com/docker/docker/pkg/term"
)

type DockerTerminal struct{}

func (t *DockerTerminal) GetWindowWidth() (uint16, error) {
	winsize, err := term.GetWinsize(os.Stdout.Fd())
	if err != nil {
		return 0, err
	}
	return winsize.Width, nil
}
