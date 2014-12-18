package test_helpers

import (
	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func ExecuteCommandWithArgs(command cli.Command, commandArgs []string) {
	cliApp := cli.NewApp()
	cliApp.Commands = []cli.Command{command}
	cliAppArgs := append([]string{"lattice-cli", command.Name}, commandArgs...)
	err := cliApp.Run(cliAppArgs)
	Expect(err).To(BeNil())
}

func AsyncExecuteCommandWithArgs(command cli.Command, commandArgs []string) chan struct{} {
	commandDone := make(chan struct{})

	go func() {
		defer GinkgoRecover()
		ExecuteCommandWithArgs(command, commandArgs)
		close(commandDone)
	}()

	return commandDone
}
