package command_factory_test

import (
	"errors"
	"io"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
)

var _ = Describe("CommandFactory", func() {
	var (
		stdinReader     *io.PipeReader
		stdinWriter     *io.PipeWriter
		outputBuffer    *gbytes.Buffer
		targetCommand   cli.Command
		config          *config_package.Config
		targetVerifier  *fake_target_verifier.FakeTargetVerifier
		fakeExitHandler *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		stdinReader, stdinWriter = io.Pipe()
		outputBuffer = gbytes.NewBuffer()
		targetVerifier = &fake_target_verifier.FakeTargetVerifier{}

		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}

		config = config_package.New(persister.NewMemPersister())

		commandFactory := command_factory.NewConfigCommandFactory(config, targetVerifier, stdinReader, output.New(outputBuffer), fakeExitHandler)
		targetCommand = commandFactory.MakeTargetCommand()
	})

	Describe("TargetCommand", func() {
		verifyOldTargetStillSet := func() {
			config.Load()
			Expect(config.Receptor()).To(Equal("http://olduser:oldpass@receptor.oldtarget.com"))
		}

		BeforeEach(func() {
			config.SetTarget("oldtarget.com")
			config.SetLogin("olduser", "oldpass")
			config.Save()
		})

		Context("displaying the target", func() {
			It("outputs the current target", func() {
				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Target:\t\toldtarget.com\n"))
				Expect(outputBuffer).To(test_helpers.Say("Username:\tolduser"))
			})

			It("does not show the username if no username is set", func() {
				config.SetLogin("", "")

				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{})

				Expect(outputBuffer).ToNot(test_helpers.Say("Username:"))
			})

			It("alerts the user if no target is set", func() {
				config.SetTarget("")
				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Target not set."))
			})
		})

		Context("setting target without auth", func() {
			BeforeEach(func() {
				targetVerifier.VerifyTargetReturns(true, true, nil)
			})

			It("saves the new target", func() {
				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(targetVerifier.VerifyTargetCallCount()).To(Equal(1))
				Expect(targetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.myapi.com"))

				Expect(config.Receptor()).To(Equal("http://receptor.myapi.com"))
			})

			It("clears out existing saved target credentials", func() {
				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Expect(targetVerifier.VerifyTargetCallCount()).To(Equal(1))
				Expect(targetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.myapi.com"))
			})

			It("bubbles up errors from setting the target", func() {
				commandFactory := command_factory.NewConfigCommandFactory(config_package.New(errorPersister("FAILURE setting api")), targetVerifier, stdinReader, output.New(outputBuffer), fakeExitHandler)
				targetCommand = commandFactory.MakeTargetCommand()

				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("FAILURE setting api"))
			})
		})

		Context("setting target that requiries auth", func() {
			BeforeEach(func() {
				targetVerifier.VerifyTargetReturns(true, false, nil)
			})

			It("sets the api, username, password from the target specified", func() {
				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("testusername\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))

				targetVerifier.VerifyTargetReturns(true, true, nil)
				stdinWriter.Write([]byte("testpassword\n"))

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
				Expect(outputBuffer).To(test_helpers.Say("Api Location Set"))

				Expect(targetVerifier.VerifyTargetCallCount()).To(Equal(2))
				Expect(targetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.myapi.com"))
				Expect(targetVerifier.VerifyTargetArgsForCall(1)).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
			})

			It("does not save the config if the receptor is never authorized", func() {
				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"newtarget.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("notgood\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("evenworse\n"))

				Eventually(commandFinishChan).Should(BeClosed())
				Expect(outputBuffer).To(test_helpers.Say("Could not authorize target."))

				verifyOldTargetStillSet()
				Expect(fakeExitHandler.ExitCalledWith[0]).To(Equal(exit_codes.BadTarget))

			})

			It("does not save the config if there is an error connecting to the receptor after prompting", func() {
				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"newtarget.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("notgood\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))

				targetVerifier.VerifyTargetReturns(true, false, errors.New("Unknown Error"))
				stdinWriter.Write([]byte("evenworse\n"))

				Eventually(commandFinishChan).Should(BeClosed())
				Expect(outputBuffer).To(test_helpers.Say("Error verifying target: Unknown Error"))

				verifyOldTargetStillSet()
				Expect(fakeExitHandler.ExitCalledWith[0]).To(Equal(exit_codes.BadTarget))
			})
		})

		Context("setting an invalid target", func() {
			It("does not save the config if the target verifier returns an error", func() {
				targetVerifier.VerifyTargetReturns(true, false, errors.New("Unknown Error"))

				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"newtarget.com"})

				Expect(outputBuffer).To(test_helpers.Say("Error verifying target: Unknown Error"))

				verifyOldTargetStillSet()
				Expect(fakeExitHandler.ExitCalledWith[0]).To(Equal(exit_codes.BadTarget))
			})
		})
	})
})

type errorPersister string

func (f errorPersister) Load(i interface{}) error {
	return errors.New(string(f))
}

func (f errorPersister) Save(i interface{}) error {
	return errors.New(string(f))
}
