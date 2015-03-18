// Copied from https://code.google.com/p/gopass/

// +build darwin freebsd linux netbsd openbsd

package terminal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

//	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
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

// TODO:  attach method to command factory or pass exit handler as arg
func (pr passwordReader) PromptForPassword(promptText string, args ...interface{}) (passwd string) {

	// Display the prompt.
	fmt.Printf(promptText)

	// File descriptors for stdin, stdout, and stderr.
	fd := []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()}

	pr.exitHandler.OnExit(func() {
		echoOn(fd)
	})

	pid, err := echoOff(fd)
	defer echoOn(fd)
	if err != nil {
		return
	}

	passwd = readPassword(pid)

	// Carriage return after the user input.
	fmt.Println("")

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
		//		return 0, fmt.Errorf(T("failed turning off console echo for password entry:\n{{.ErrorDescription}}", map[string]interface{}{"ErrorDescription": err}))
		return 0, errors.New("blah")
	}

	return pid, nil
}

// echoOn turns back on the terminal echo.
func echoOn(fd []uintptr) {
	// Turn on the terminal echo.
	pid, e := syscall.ForkExec(sttyArg0, sttyArgvEOn, &syscall.ProcAttr{Dir: exec_cwdir, Files: fd})

	if e == nil {
		syscall.Wait4(pid, &ws, 0, nil)
	}
}
