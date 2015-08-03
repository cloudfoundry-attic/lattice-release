package command_factory_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory/fake_blob_store_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("CommandFactory", func() {
	var (
		stdinReader           *io.PipeReader
		stdinWriter           *io.PipeWriter
		outputBuffer          *gbytes.Buffer
		terminalUI            terminal.UI
		configPersister       persister.Persister
		config                *config_package.Config
		fakeTargetVerifier    *fake_target_verifier.FakeTargetVerifier
		fakeBlobStoreVerifier *fake_blob_store_verifier.FakeBlobStoreVerifier
		fakeExitHandler       *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		stdinReader, stdinWriter = io.Pipe()
		outputBuffer = gbytes.NewBuffer()
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		fakeBlobStoreVerifier = &fake_blob_store_verifier.FakeBlobStoreVerifier{}
		fakeExitHandler = new(fake_exit_handler.FakeExitHandler)
		terminalUI = terminal.NewUI(stdinReader, outputBuffer, nil)
		configPersister = persister.NewMemPersister()
		config = config_package.New(configPersister)
	})

	Describe("TargetBlobCommand", func() {
		var targetBlobCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewConfigCommandFactory(config, terminalUI, fakeTargetVerifier, fakeBlobStoreVerifier, fakeExitHandler)
			targetBlobCommand = commandFactory.MakeTargetBlobCommand()
		})

		Context("displaying the blob target", func() {
			It("outputs the current target", func() {
				config.SetBlobStore("192.168.11.11", "8980", "some-username", "some-password")
				Expect(config.Save()).To(Succeed())

				test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Blob Store:\t192.168.11.11:8980\n"))
				Expect(outputBuffer).To(test_helpers.Say("Username:\tsome-username"))
				Expect(outputBuffer).To(test_helpers.Say("Password:\tsome-password"))
			})

			It("alerts the user if no target is set", func() {
				config.SetBlobStore("", "", "", "")
				Expect(config.Save()).To(Succeed())

				test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Blob store not set"))
			})
		})

		Context("setting the blob target", func() {
			It("sets the blob target and credentials", func() {
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8980"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("some-username\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("some-password\n"))

				Eventually(doneChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.Say("Blob Location Set"))

				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				blobTargetInfo := fakeTargetVerifier.VerifyBlobTargetArgsForCall(0)
				Expect(blobTargetInfo.Host).To(Equal("192.168.11.11"))
				Expect(blobTargetInfo.Port).To(Equal("8980"))
				Expect(blobTargetInfo.Username).To(Equal("some-username"))
				Expect(blobTargetInfo.Password).To(Equal("some-password"))

				newConfig := config_package.New(configPersister)
				Expect(newConfig.Load()).To(Succeed())
				blobTarget := newConfig.BlobStore()
				Expect(blobTarget.Host).To(Equal("192.168.11.11"))
				Expect(blobTarget.Port).To(Equal("8980"))
				Expect(blobTarget.Username).To(Equal("some-username"))
				Expect(blobTarget.Password).To(Equal("some-password"))
			})

			It("sets the blob target and credentials using the default bucket name", func() {
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8980"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				stdinWriter.Write([]byte("some-username\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
				stdinWriter.Write([]byte("some-password\n"))

				Eventually(doneChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.Say("Blob Location Set"))

				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				blobTargetInfo := fakeTargetVerifier.VerifyBlobTargetArgsForCall(0)
				Expect(blobTargetInfo.Host).To(Equal("192.168.11.11"))
				Expect(blobTargetInfo.Port).To(Equal("8980"))
				Expect(blobTargetInfo.Username).To(Equal("some-username"))
				Expect(blobTargetInfo.Password).To(Equal("some-password"))

				newConfig := config_package.New(configPersister)
				Expect(newConfig.Load()).To(Succeed())
				blobTarget := newConfig.BlobStore()
				Expect(blobTarget.Host).To(Equal("192.168.11.11"))
				Expect(blobTarget.Port).To(Equal("8980"))
				Expect(blobTarget.Username).To(Equal("some-username"))
				Expect(blobTarget.Password).To(Equal("some-password"))
			})

			Context("invalid syntax", func() {
				It("errors when target is malformed", func() {
					test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{"huehue8980"})

					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed target"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(BeZero())
				})
				It("errors when port is non-numeric", func() {
					fakeTargetVerifier.VerifyBlobTargetReturns(errors.New("blob target is down: dial tcp: unknown port tcp/haiii"))

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:haiii"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					stdinWriter.Write([]byte("some-username\n"))
					Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
					stdinWriter.Write([]byte("some-password\n"))

					Eventually(doneChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.Say("Unable to verify blob store: blob target is down: dial tcp: unknown port tcp/haiii"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				})
				It("errors when port exceeds 65536", func() {
					fakeTargetVerifier.VerifyBlobTargetReturns(errors.New("blob target is down: dial tcp: invalid port 70000"))

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:70000"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					stdinWriter.Write([]byte("some-username\n"))
					Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
					stdinWriter.Write([]byte("some-password\n"))

					Eventually(doneChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.Say("Unable to verify blob store: blob target is down: dial tcp: invalid port 70000"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				})
			})

			Context("scenarios that should not save the config", func() {
				verifyConfigNotSaved := func(failMessage string) {
					Expect(outputBuffer).NotTo(test_helpers.Say("Blob Location Set"))
					Expect(outputBuffer).To(test_helpers.Say(failMessage))
					newConfig := config_package.New(configPersister)
					Expect(newConfig.Load()).To(Succeed())
					blobTarget := newConfig.BlobStore()
					Expect(blobTarget.Host).To(Equal("original-host"))
					Expect(blobTarget.Port).To(Equal("8989"))
					Expect(blobTarget.Username).To(Equal("original-key"))
					Expect(blobTarget.Password).To(Equal("original-secret"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
				}

				BeforeEach(func() {
					config.SetBlobStore("original-host", "8989", "original-key", "original-secret")
					Expect(config.Save()).To(Succeed())
				})

				It("does not save the config when there is an error connecting to the target", func() {
					fakeTargetVerifier.VerifyBlobTargetReturns(errors.New("fail"))

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8980"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					stdinWriter.Write([]byte("some-username\n"))
					Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
					stdinWriter.Write([]byte("some-password\n"))

					Eventually(doneChan).Should(BeClosed())

					verifyConfigNotSaved("Unable to verify blob store: fail")
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				})

				Context("when the persister returns errors", func() {
					BeforeEach(func() {
						commandFactory := command_factory.NewConfigCommandFactory(
							config_package.New(errorPersister("Failure setting blob target")),
							terminalUI,
							fakeTargetVerifier,
							fakeBlobStoreVerifier,
							fakeExitHandler,
						)
						targetBlobCommand = commandFactory.MakeTargetBlobCommand()
					})
					It("bubbles up errors from saving the config", func() {
						config.SetBlobStore("192.168.11.11", "8980", "some-username", "some-password")
						Expect(config.Save()).To(Succeed())

						doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"199.112.3432:8980"})

						Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
						stdinWriter.Write([]byte("some-different-username\n"))
						Eventually(outputBuffer).Should(test_helpers.Say("Password: "))
						stdinWriter.Write([]byte("some-different-password\n"))

						Eventually(doneChan).Should(BeClosed())

						Expect(outputBuffer).To(test_helpers.Say("Failure setting blob target"))

						newConfig := config_package.New(configPersister)
						Expect(newConfig.Load()).To(Succeed())
						blobTarget := newConfig.BlobStore()
						Expect(blobTarget.Host).To(Equal("192.168.11.11"))
						Expect(blobTarget.Port).To(Equal("8980"))
						Expect(blobTarget.Username).To(Equal("some-username"))
						Expect(blobTarget.Password).To(Equal("some-password"))

						Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
						Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
					})
				})
			})
		})

	})
})
