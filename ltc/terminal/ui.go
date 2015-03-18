package terminal

import (
	"bufio"
	"io"
	"strings"
)

type UI interface {
	Say(message string)
	SayLine(message string)
	IncorrectUsage(message string)
	Dot()
	NewLine()
	Prompt(promptText string) string

	io.Reader
	io.Writer
}

type terminalUI struct {
	io.Reader
	io.Writer
}

func NewUI(input io.Reader, output io.Writer) UI {
	return &terminalUI{
		input,
		output,
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

func (t *terminalUI) Prompt(promptText string) (answer string) {
	reader := bufio.NewReader(t)
	t.Say(promptText)

	result, _ := reader.ReadString('\n')

	return strings.TrimSuffix(result, "\n")
}
