package main

import (
	"flag"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/logs"
	"io"
	"os"
	"os/exec"
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
    `) //TODO: VALIDATE FLAG PASSED IN!!

	flag.StringVar(
		&sourceInstance,
		"sourceInstance",
		"",
		"The label for the log source instance that shows up when consuming the stream",
	) //TODO: VALIDATE FLAG PASSED IN!!
	flag.Parse()
}

func main() {
	args := flag.Args()
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
	go logs.ScanLogStream(latticeDebugStreamId, args[0], sourceInstance, dropsondeStdoutReader) //TODO: UNIT TEST THIS, integration test won't catch appId
	go logs.ScanErrorLogStream(latticeDebugStreamId, args[0], sourceInstance, dropsondeStderrReader)

	err = cmd.Start()
	if err != nil {
		panic(err) //TODO: COULD USE A UNIT TEST!
	}

	cmd.Wait()
}
