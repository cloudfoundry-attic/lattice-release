package terminal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader"
)

type UI interface {
	io.ReadWriter
	password_reader.PasswordReader

	Prompt(promptText string, args ...interface{}) string
	PromptWithDefault(promptText, defaultValue string, args ...interface{}) string
	Say(message string)
	SayIncorrectUsage(message string)
	SayLine(message string)
	SayNewLine()
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

func (t *terminalUI) Prompt(promptText string, args ...interface{}) (answer string) {
	reader := bufio.NewReader(t)
	fmt.Fprintf(t.Writer, promptText+": ", args...)

	result, _ := reader.ReadString('\n')
	return strings.TrimSuffix(result, "\n")
}

func (t *terminalUI) PromptWithDefault(promptText, defaultValue string, args ...interface{}) (answer string) {
	reader := bufio.NewReader(t)
	fmt.Fprintf(t.Writer, promptText+fmt.Sprintf(" [%s]: ", defaultValue), args...)

	result, _ := reader.ReadString('\n')
	result = strings.TrimSuffix(result, "\n")

	if result == "" {
		return defaultValue
	}

	return result
}

func (t *terminalUI) Say(message string) {
	t.Write([]byte(message))
}

func (t *terminalUI) SayIncorrectUsage(message string) {
	if len(message) > 0 {
		t.SayLine("Incorrect Usage: " + message)
	} else {
		t.SayLine("Incorrect Usage")
	}
}

func (t *terminalUI) SayLine(message string) {
	t.Write([]byte(message + "\n"))
}

func (t *terminalUI) SayNewLine() {
	t.Say("\n")
}
