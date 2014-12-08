package test_helpers

import (
	"github.com/dajulia3/cli"
)

func ExecuteCommandWithArgs(command cli.Command, commandArgs []string) error {
	cliApp := cli.NewApp()
	cliApp.Commands = []cli.Command{command}
	cliAppArgs := append([]string{"diego-edge-cli", command.Name}, commandArgs...)
	return cliApp.Run(cliAppArgs)
}
