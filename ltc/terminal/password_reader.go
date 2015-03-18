package terminal

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"io"
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

//func (pr passwordReader) PromptForPassword(promptText string) string {
//    reader := bufio.NewReader(pr)
//    fmt.Print(promptText)
//
//    //disable echo
//    result, _ := reader.ReadString('\n')
//    //enable echo
//
//    return strings.TrimSuffix(result, "\n")
//}
