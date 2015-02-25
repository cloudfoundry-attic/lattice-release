package main

import (
	"os"

	"github.com/cloudfoundry-incubator/lattice/ltc/setup_cli"
)

func main() {
	defer os.Stdout.Write([]byte("\n"))
	cliApp := setup_cli.NewCliApp()
	cliApp.Run(os.Args)
}
