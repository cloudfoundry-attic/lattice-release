package test_helpers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/codegangsta/cli"
)

func ExecuteCommandWithArgs(command cli.Command, commandArgs []string) {
	executeCommandWithArgs(command, commandArgs)
}

func AsyncExecuteCommandWithArgs(command cli.Command, commandArgs []string) chan struct{} {
	doneChan := make(chan struct{})

	go func() {
		defer GinkgoRecover()
		executeCommandWithArgs(command, commandArgs)
		close(doneChan)
	}()

	return doneChan
}

func executeCommandWithArgs(command cli.Command, commandArgs []string) {
	cliApp := cli.NewApp()
	cliApp.Commands = []cli.Command{command}
	cliAppArgs := append([]string{"ltc", command.Name}, commandArgs...)

	Expect(cliApp.Run(cliAppArgs)).To(Succeed())
}
