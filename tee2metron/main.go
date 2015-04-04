package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/logs"
)

var dropsondeDestination, sourceInstance string

const latticeDebugStreamId = "lattice-debug"

func init() {
	flag.StringVar(
		&dropsondeDestination,
		"dropsondeDestination",
		"",
		`where to stream logs to in the form of hostname:port
    eg. -dropsondeDestination=127.0.0.1:3457
    `)

	flag.StringVar(
		&sourceInstance,
		"sourceInstance",
		"",
		"The label for the log source instance that shows up when consuming the stream",
	)
	flag.Parse()
}

func main() {
	if dropsondeDestination == "" {
		fmt.Println("dropsondeDestination flag is required")
		os.Exit(1)
	}
	if sourceInstance == "" {
		fmt.Println("sourceInstance flag is required")
		os.Exit(1)
	}

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Command not specified!")
		fmt.Println("Usage: tee2metron -dropsondeDestionation=127.0.0.1:3457 -sourceInstance=lattice-cell-21 COMMAND")
		os.Exit(3)
	}
	err := dropsonde.Initialize(dropsondeDestination, sourceInstance, args[0])

	if err != nil {
		panic("error initializing dropsonde" + err.Error())
	}

	dropsondeStdoutReader, dropsondeStdoutWriter := io.Pipe()
	dropsondeStderrReader, dropsondeStderrWriter := io.Pipe()

	stdoutTeeWriter := io.MultiWriter(dropsondeStdoutWriter, os.Stdout)
	stderrTeeWriter := io.MultiWriter(dropsondeStderrWriter, os.Stderr)

	defer dropsondeStdoutReader.Close()
	defer dropsondeStderrReader.Close()
	defer dropsondeStdoutWriter.Close()
	defer dropsondeStderrWriter.Close()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = stdoutTeeWriter
	cmd.Stderr = stderrTeeWriter
	go logs.ScanLogStream(latticeDebugStreamId, args[0], sourceInstance, dropsondeStdoutReader)
	go logs.ScanErrorLogStream(latticeDebugStreamId, args[0], sourceInstance, dropsondeStderrReader)

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	cmd.Wait()
}
