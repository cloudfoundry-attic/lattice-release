package password_reader

import (
	"io"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
)

//go:generate counterfeiter -o fake_password_reader/fake_password_reader.go . PasswordReader
type PasswordReader interface {
	PromptForPassword(promptText string, args ...interface{}) (passwd string)
}

type passwordReader struct {
	io.Reader
	exitHandler exit_handler.ExitHandler
}

func NewPasswordReader(exitHandler exit_handler.ExitHandler) *passwordReader {
	return &passwordReader{
		exitHandler: exitHandler,
	}
}
