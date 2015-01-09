package colors

import (
	"fmt"
	"strings"
)

var ColorCodeLength = len(red) + len(defaultStyle)

const (
	red          string = "\x1b[91m"
	cyan         string = "\x1b[36m"
	green        string = "\x1b[32m"
	yellow       string = "\x1b[33m"
	defaultStyle string = "\x1b[0m"
	boldStyle    string = "\x1b[1m"
)

func Red(output string) string {
	return colorize(output, red)
}

func Green(output string) string {
	return colorize(output, green)
}

func Cyan(output string) string {
	return colorize(output, cyan)
}

func Yellow(output string) string {
	return colorize(output, yellow)
}

func NoColor(output string) string {
	return colorize(output, defaultStyle)
}

func Bold(output string) string {
	return colorize(output, boldStyle)
}

func colorize(output string, color string) string {
	if strings.TrimSpace(output) == "" {
		return output
	}
	return fmt.Sprintf("%s%s%s", color, output, defaultStyle)
}
