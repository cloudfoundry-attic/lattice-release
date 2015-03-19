package terminal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader"
)

type UI interface {
	Say(message string)
	SayLine(message string)
	IncorrectUsage(message string)
	Dot()
	NewLine()
	Prompt(promptText string, args ...interface{}) string

	io.ReadWriter
	password_reader.PasswordReader
}

type terminalUI struct {
	io.Reader
	io.Writer
	password_reader.PasswordReader
}

func NewUI(input io.Reader, output io.Writer, passwordReader password_reader.PasswordReader) UI {
	return &terminalUI{
		input,
		output,
		passwordReader,
	}
}

func (t *terminalUI) Say(message string) {
	t.Write([]byte(message))
}

func (t *terminalUI) SayLine(message string) {
	t.Write([]byte(message + "\n"))
}

func (t *terminalUI) IncorrectUsage(message string) {
	if len(message) > 0 {
		t.Say("Incorrect Usage: " + message)
	} else {
		t.Say("Incorrect Usage")
	}
}

func (t *terminalUI) Dot() {
	t.Say(".")
}

func (t *terminalUI) NewLine() {
	t.Say("\n")
}

func (t *terminalUI) Prompt(promptText string, args ...interface{}) (answer string) {
	reader := bufio.NewReader(t)
	fmt.Fprintf(t.Writer, promptText, args...)

	result, _ := reader.ReadString('\n')

	return strings.TrimSuffix(result, "\n")
}
