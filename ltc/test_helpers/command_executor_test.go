package test_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("CommandExecutor", func() {

	var (
		commandRan bool
		cliCommand cli.Command
	)

	BeforeEach(func() {
		commandRan = false

		cliCommand = cli.Command{
			Name: "exec",
			Action: func(*cli.Context) {
				commandRan = true
			},
		}
	})

	Describe("ExecuteCommandWithArgs", func() {
		It("executes the command", func() {
			test_helpers.ExecuteCommandWithArgs(cliCommand, []string{})

			Expect(commandRan).To(BeTrue())
		})
	})

	Describe("AsyncExecuteCommandWithArgs", func() {
		It("executes the command", func() {
			doneChan := test_helpers.AsyncExecuteCommandWithArgs(cliCommand, []string{})

			Eventually(doneChan).Should(BeClosed())
			Expect(commandRan).To(BeTrue())
		})
	})
})
