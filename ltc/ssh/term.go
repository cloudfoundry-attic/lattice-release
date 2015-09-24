package ssh

import "github.com/docker/docker/pkg/term"

type DockerTerm struct{}

func (*DockerTerm) SetRawTerminal(fd uintptr) (*term.State, error) {
	return term.SetRawTerminal(fd)
}

func (*DockerTerm) RestoreTerminal(fd uintptr, state *term.State) error {
	return term.RestoreTerminal(fd, state)
}

func (*DockerTerm) GetWinsize(fd uintptr) (width, height int) {
	winSize, err := term.GetWinsize(fd)
	if err != nil {
		return 80, 43
	}

	return int(winSize.Width), int(winSize.Height)
}

func (*DockerTerm) IsTTY(fd uintptr) bool {
	return term.IsTerminal(fd)
}
