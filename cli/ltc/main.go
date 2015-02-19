package main

import (
	"os"

	"github.com/pivotal-cf-experimental/lattice-cli/ltc/setup_cli"
)

func main() {
	defer os.Stdout.Write([]byte("\n"))
	cliApp := setup_cli.NewCliApp()
	cliApp.Run(os.Args)
}
