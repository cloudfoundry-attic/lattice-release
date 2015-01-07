package colors

import "fmt"

const (
	red          string = "\x1b[91m"
	cyan         string = "\x1b[36m"
	green        string = "\x1b[32m"
	yellow       string = "\x1b[33m"
	defaultStyle string = "\x1b[0m"
	boldStyle    string = "\x1b[1m"
)

func Red(output string) string {
	return fmt.Sprintf("%s%s%s", red, output, defaultStyle)
}

func Green(output string) string {
	return fmt.Sprintf("%s%s%s", green, output, defaultStyle)
}

func Cyan(output string) string {
	return fmt.Sprintf("%s%s%s", cyan, output, defaultStyle)
}

func Yellow(output string) string {
	return fmt.Sprintf("%s%s%s", yellow, output, defaultStyle)
}

func NoColor(output string) string {
	return fmt.Sprintf("%s%s%s", defaultStyle, output, defaultStyle)
}

func Bold(output string) string {
	return fmt.Sprintf("%s%s%s", boldStyle, output, defaultStyle)
}
