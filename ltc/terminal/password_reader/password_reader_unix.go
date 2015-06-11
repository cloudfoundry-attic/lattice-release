// Copied from https://code.google.com/p/gopass/

// +build darwin freebsd linux netbsd openbsd
package password_reader

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

const (
	sttyArg0   = "/bin/stty"
	exec_cwdir = ""
)

// Tells the terminal to turn echo off.
var sttyArgvEOff []string = []string{"stty", "-echo"}

// Tells the terminal to turn echo on.
var sttyArgvEOn []string = []string{"stty", "echo"}

var ws syscall.WaitStatus = 0

func (pr passwordReader) PromptForPassword(promptText string, args ...interface{}) (passwd string) {

	// Display the prompt.
	fmt.Printf(promptText+": ", args...)

	// File descriptors for stdin, stdout, and stderr.
	fd := []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()}

	pr.exitHandler.OnExit(func() { echoOn(fd) })

	pid, err := echoOff(fd)
	defer echoOn(fd)
	if err != nil {
		return
	}

	passwd = readPassword(pid)

	fmt.Println("") // Carriage return after the user input.

	return
}

func readPassword(pid int) string {
	rd := bufio.NewReader(os.Stdin)
	syscall.Wait4(pid, &ws, 0, nil)

	line, err := rd.ReadString('\n')
	if err == nil {
		return strings.TrimSpace(line)
	}
	return ""
}

func echoOff(fd []uintptr) (int, error) {
	pid, err := syscall.ForkExec(sttyArg0, sttyArgvEOff, &syscall.ProcAttr{Dir: exec_cwdir, Files: fd})
	if err != nil {
		return 0, fmt.Errorf("failed turning off console echo for password entry:\n{{.ErrorDescription}}", map[string]interface{}{"ErrorDescription": err})
	}

	return pid, nil
}

func echoOn(fd []uintptr) {
	pid, e := syscall.ForkExec(sttyArg0, sttyArgvEOn, &syscall.ProcAttr{Dir: exec_cwdir, Files: fd})
	if e == nil {
		syscall.Wait4(pid, &ws, 0, nil)
	}
}
