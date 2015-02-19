package cursor

import "fmt"

const csi = "\033["

func Up(lines int) string {
	return fmt.Sprintf("%s%dA", csi, lines)
}

func ClearToEndOfLine() string {
	return fmt.Sprintf("%s%dK", csi, 0)
}

func ClearToEndOfDisplay() string {
	return fmt.Sprintf("%s%dJ", csi, 0)
}

func Show() string {
	return csi + "?25h"
}

func Hide() string {
	return csi + "?25l"
}
