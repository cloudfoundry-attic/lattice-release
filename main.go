package main

import (
	"os"

	"github.com/pivotal-cf-experimental/diego-edge-cli/setup_cli"
)

func main() {
	cliApp := setup_cli.NewCliApp()
	cliApp.Run(os.Args)
}
