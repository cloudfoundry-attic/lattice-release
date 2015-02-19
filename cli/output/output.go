package output

import (
	"io"
)

func New(writer io.Writer) *Output {
	return &Output{writer}
}

type Output struct {
	io.Writer
}

func (o *Output) Say(message string) {
	o.Write([]byte(message))
}

func (o *Output) SayLine(message string) {
	o.Write([]byte(message + "\n"))
}

func (o *Output) IncorrectUsage(message string) {
	if len(message) > 0 {
		o.Say("Incorrect Usage: " + message)
	} else {
		o.Say("Incorrect Usage")
	}
}

func (o *Output) Dot() {
	o.Say(".")
}

func (o *Output) NewLine() {
	o.Say("\n")
}
