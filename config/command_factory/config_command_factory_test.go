package command_factory_test

import (
	"errors"
	"io"

	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	config_package "github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	"github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
)

var _ = Describe("CommandFactory", func() {
	Describe("setApiEndpoint", func() {
		var (
			stdinReader      *io.PipeReader
			stdinWriter      *io.PipeWriter
			buffer           *gbytes.Buffer
			setTargetCommand cli.Command
			config           *config_package.Config
		)

		BeforeEach(func() {
			stdinReader, stdinWriter = io.Pipe()
			buffer = gbytes.NewBuffer()
			config = config_package.New(persister.NewFakePersister())

			commandFactory := command_factory.NewConfigCommandFactory(config, stdinReader, output.New(buffer))
			setTargetCommand = commandFactory.MakeSetTargetCommand()
		})

		Describe("targetCommand", func() {

			It("sets the api, username, password from the target specified", func() {
				setTargetCommandDone := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(buffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("testusername\n"))
				Eventually(buffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("testpassword\n"))

				<-setTargetCommandDone

				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
				Expect(buffer).To(test_helpers.Say("Api Location Set"))
			})

			It("does not update the config if error on reading username", func() {
				config.SetTarget("oldtarget.com")
				config.SetLogin("olduser", "oldpass")

				setTargetCommandDone := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(buffer).Should(test_helpers.Say("Username:"))
				stdinWriter.Close()

				Consistently(buffer).ShouldNot(test_helpers.Say("Api Location Set"))
				<-setTargetCommandDone

				Expect(config.Receptor()).To(Equal("http://olduser:oldpass@receptor.oldtarget.com"))
			})

			It("does not update the config if error on reading password", func() {
				config.SetTarget("oldtarget.com")
				config.SetLogin("olduser", "oldpass")

				commandDone := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(buffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("testusername\n"))
				Eventually(buffer).Should(test_helpers.Say("Password:"))
				stdinWriter.Close()

				<-commandDone

				Consistently(buffer).ShouldNot(test_helpers.Say("Api Location Set"))
				Expect(config.Receptor()).To(Equal("http://olduser:oldpass@receptor.oldtarget.com"))
			})

			It("does not set a username or password if none are passed in", func() {

				commandDone := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(buffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("\n"))

				Eventually(buffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("\n"))

				<-commandDone

				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://receptor.myapi.com"))
				Expect(buffer).To(test_helpers.Say("Api Location Set"))
			})

			It("returns an error if the target is blank", func() {
				test_helpers.ExecuteCommandWithArgs(setTargetCommand, []string{""})

				Expect(buffer).To(test_helpers.Say("Incorrect Usage: Target required."))
			})

			It("bubbles errors from setting the target", func() {
				fakePersister := persister.NewFakePersisterWithError(errors.New("FAILURE setting api"))

				commandFactory := command_factory.NewConfigCommandFactory(config_package.New(fakePersister), stdinReader, output.New(buffer))
				setTargetCommand = commandFactory.MakeSetTargetCommand()

				commandDone := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(buffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("\n"))
				Eventually(buffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("\n"))

				<-commandDone
				Eventually(buffer).Should(test_helpers.Say("FAILURE setting api"))
			})
		})

	})
})
