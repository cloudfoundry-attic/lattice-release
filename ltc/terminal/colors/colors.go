package colors

import (
	"fmt"
	"strings"
)

var ColorCodeLength = len(red) + len(defaultStyle)

const (
	red             string = "\x1b[91m"
	cyan            string = "\x1b[36m"
	green           string = "\x1b[32m"
	yellow          string = "\x1b[33m"
	purpleUnderline string = "\x1b[35;4m"
	defaultStyle    string = "\x1b[0m"
	boldStyle       string = "\x1b[1m"
	grayColor       string = "\x1b[90m"
)

func Red(output string) string {
	return colorText(output, red)
}

func Green(output string) string {
	return colorText(output, green)
}

func Cyan(output string) string {
	return colorText(output, cyan)
}

func Yellow(output string) string {
	return colorText(output, yellow)
}

func Gray(output string) string {
	return colorText(output, grayColor)
}

func NoColor(output string) string {
	return colorText(output, defaultStyle)
}

func Bold(output string) string {
	return colorText(output, boldStyle)
}

func PurpleUnderline(output string) string {
	return colorText(output, purpleUnderline)
}

func colorText(output string, color string) string {
	if strings.TrimSpace(output) == "" {
		return output
	}
	return fmt.Sprintf("%s%s%s", color, output, defaultStyle)
}
