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
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier/fake_target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	"github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
)

var _ = Describe("CommandFactory", func() {
	Describe("setApiEndpoint", func() {
		var (
			stdinReader      *io.PipeReader
			stdinWriter      *io.PipeWriter
			outputBuffer     *gbytes.Buffer
			setTargetCommand cli.Command
			config           *config_package.Config
			targetVerifier   *fake_target_verifier.FakeTargetVerifier
		)

		BeforeEach(func() {
			stdinReader, stdinWriter = io.Pipe()
			outputBuffer = gbytes.NewBuffer()
			targetVerifier = &fake_target_verifier.FakeTargetVerifier{}

			config = config_package.New(persister.NewFakePersister())

			commandFactory := command_factory.NewConfigCommandFactory(config, targetVerifier, stdinReader, output.New(outputBuffer))
			setTargetCommand = commandFactory.MakeSetTargetCommand()
		})

		Describe("targetCommand", func() {
			It("sets the api, username, password from the target specified", func() {
				targetVerifier.ValidateReceptorReturns(false)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("testusername\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))

				targetVerifier.ValidateReceptorReturns(true)
				stdinWriter.Write([]byte("testpassword\n"))

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
				Expect(outputBuffer).To(test_helpers.Say("Api Location Set"))

				Expect(targetVerifier.ValidateReceptorCallCount()).To(Equal(2))
				Expect(targetVerifier.ValidateReceptorArgsForCall(0)).To(Equal("http://receptor.myapi.com"))
				Expect(targetVerifier.ValidateReceptorArgsForCall(1)).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
			})

			It("clears out existing saved credentials", func() {
				targetVerifier.ValidateReceptorReturns(true)

				config.SetTarget("oldtarget.com")
				config.SetLogin("olduser", "oldpass")
				config.Save()

				test_helpers.ExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Expect(targetVerifier.ValidateReceptorCallCount()).To(Equal(1))
				Expect(targetVerifier.ValidateReceptorArgsForCall(0)).To(Equal("http://receptor.myapi.com"))
			})

			It("does not ask for username and password if it does not require auth", func() {
				targetVerifier.ValidateReceptorReturns(true)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(targetVerifier.ValidateReceptorCallCount()).To(Equal(1))
				Expect(targetVerifier.ValidateReceptorArgsForCall(0)).To(Equal("http://receptor.myapi.com"))

				Expect(config.Receptor()).To(Equal("http://receptor.myapi.com"))
			})

			It("does not save the config if the receptor is never validated", func() {
				targetVerifier.ValidateReceptorReturns(false)

				config.SetTarget("oldtarget.com")
				config.Save()

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(setTargetCommand, []string{"newtarget.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("notgood\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("evenworse\n"))

				Eventually(commandFinishChan).Should(BeClosed())
				Expect(outputBuffer).To(test_helpers.Say("Could not verify target."))

				config.Load()
				Expect(config.Target()).To(Equal("oldtarget.com"))
			})

			It("returns an error if the target is blank", func() {
				test_helpers.ExecuteCommandWithArgs(setTargetCommand, []string{""})

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Target required."))
			})

			It("bubbles up errors from setting the target", func() {
				targetVerifier.ValidateReceptorReturns(true)
				fakePersister := persister.NewFakePersisterWithError(errors.New("FAILURE setting api"))

				commandFactory := command_factory.NewConfigCommandFactory(config_package.New(fakePersister), targetVerifier, stdinReader, output.New(outputBuffer))
				setTargetCommand = commandFactory.MakeSetTargetCommand()

				test_helpers.ExecuteCommandWithArgs(setTargetCommand, []string{"myapi.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("FAILURE setting api"))
			})
		})

	})
})
