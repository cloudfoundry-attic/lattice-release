package main

import (
	"os"

	"github.com/pivotal-cf-experimental/diego-edge-cli/cli_config"
)

func main() {
	cliApp := cli_config.NewCliApp()
	cliApp.Run(os.Args)
}
