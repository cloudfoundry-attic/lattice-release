package command_factory_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory/fake_blob_store_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader/fake_password_reader"
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
		config                *config_package.Config
		configPersister       persister.Persister
		fakeTargetVerifier    *fake_target_verifier.FakeTargetVerifier
		fakeBlobStoreVerifier *fake_blob_store_verifier.FakeBlobStoreVerifier
		fakeExitHandler       *fake_exit_handler.FakeExitHandler
		fakePasswordReader    *fake_password_reader.FakePasswordReader
	)

	BeforeEach(func() {
		stdinReader, stdinWriter = io.Pipe()
		outputBuffer = gbytes.NewBuffer()
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
		fakePasswordReader = &fake_password_reader.FakePasswordReader{}
		terminalUI = terminal.NewUI(stdinReader, outputBuffer, fakePasswordReader)
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		fakeBlobStoreVerifier = &fake_blob_store_verifier.FakeBlobStoreVerifier{}
		configPersister = persister.NewMemPersister()
		config = config_package.New(configPersister)
	})

	Describe("TargetCommand", func() {
		var targetCommand cli.Command

		verifyOldTargetStillSet := func() {
			newConfig := config_package.New(configPersister)
			Expect(newConfig.Load()).To(Succeed())

			Expect(newConfig.Receptor()).To(Equal("http://olduser:oldpass@receptor.oldtarget.com"))
		}

		BeforeEach(func() {
			commandFactory := command_factory.NewConfigCommandFactory(config, terminalUI, fakeTargetVerifier, fakeBlobStoreVerifier, fakeExitHandler)
			targetCommand = commandFactory.MakeTargetCommand()
		})

		JustBeforeEach(func() {
			config.SetTarget("oldtarget.com")
			config.SetLogin("olduser", "oldpass")
			Expect(config.Save()).To(Succeed())
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

		Context("setting receptor target without auth", func() {
			BeforeEach(func() {
				fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
				fakeBlobStoreVerifier.VerifyReturns(true, nil)
			})

			It("saves the new receptor target", func() {
				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
				Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.myapi.com"))

				newConfig := config_package.New(configPersister)
				Expect(newConfig.Load()).To(Succeed())
				Expect(newConfig.Receptor()).To(Equal("http://receptor.myapi.com"))
			})

			It("clears out existing saved target credentials", func() {
				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
				Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.myapi.com"))
			})

			Context("when the blob store is online", func() {
				It("saves the new blob target", func() {
					fakeBlobStoreVerifier.VerifyReturns(true, nil)

					test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

					Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
					Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(dav_blob_store.Config{
						Host: "myapi.com",
						Port: "8444",
					}))

					newConfig := config_package.New(configPersister)
					Expect(newConfig.Load()).To(Succeed())
					Expect(newConfig.BlobTarget()).To(Equal(dav_blob_store.Config{
						Host: "myapi.com",
						Port: "8444",
					}))

					Expect(outputBuffer).To(test_helpers.Say("Blob store is targeted."))
				})

				Context("when the blob store requires authorization", func() {
					It("exits", func() {
						fakeBlobStoreVerifier.VerifyReturns(false, nil)

						test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

						Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
						Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(dav_blob_store.Config{
							Host: "myapi.com",
							Port: "8444",
						}))

						Expect(outputBuffer).To(test_helpers.Say("Blob store requires authorization"))
						verifyOldTargetStillSet()
						Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
					})
				})
			})

			Context("when the blob store target is offline", func() {
				It("saves the receptor target but does not save any blob target", func() {
					fakeBlobStoreVerifier.VerifyReturns(false, errors.New("unable to connect to blob store"))

					test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

					Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
					Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(dav_blob_store.Config{
						Host: "myapi.com",
						Port: "8444",
					}))

					newConfig := config_package.New(configPersister)
					Expect(newConfig.Load()).To(Succeed())
					Expect(newConfig.Receptor()).To(Equal("http://receptor.myapi.com"))
					Expect(newConfig.BlobTarget()).To(Equal(dav_blob_store.Config{}))
				})
			})

			Context("when the persister returns errors", func() {
				BeforeEach(func() {
					commandFactory := command_factory.NewConfigCommandFactory(config_package.New(errorPersister("FAILURE setting api")), terminalUI, fakeTargetVerifier, fakeBlobStoreVerifier, fakeExitHandler)
					targetCommand = commandFactory.MakeTargetCommand()
				})

				It("exits", func() {
					test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

					Eventually(outputBuffer).Should(test_helpers.Say("FAILURE setting api"))
					verifyOldTargetStillSet()
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
				})
			})
		})

		Context("setting target that requires auth", func() {
			BeforeEach(func() {
				fakeTargetVerifier.VerifyTargetReturns(true, false, nil)
				fakeBlobStoreVerifier.VerifyReturns(true, nil)
				fakePasswordReader.PromptForPasswordReturns("testpassword")
			})

			It("sets the api, username, password from the target specified", func() {
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

				Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
				fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
				stdinWriter.Write([]byte("testusername\n"))

				Eventually(doneChan).Should(BeClosed())

				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
				Expect(outputBuffer).To(test_helpers.Say("Api Location Set"))

				Expect(fakePasswordReader.PromptForPasswordCallCount()).To(Equal(1))
				Expect(fakePasswordReader.PromptForPasswordArgsForCall(0)).To(Equal("Password"))

				Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(2))
				Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.myapi.com"))
				Expect(fakeTargetVerifier.VerifyTargetArgsForCall(1)).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
			})

			Context("scenarios that should not save the config", func() {
				BeforeEach(func() {
					fakePasswordReader.PromptForPasswordReturns("evenworse")
				})

				AfterEach(func() {
					verifyOldTargetStillSet()
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
				})

				It("does not save the config if the receptor is never authorized", func() {
					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"newtarget.com"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					stdinWriter.Write([]byte("notgood\n"))

					Eventually(doneChan).Should(BeClosed())

					Expect(fakePasswordReader.PromptForPasswordCallCount()).To(Equal(1))
					Expect(fakePasswordReader.PromptForPasswordArgsForCall(0)).To(Equal("Password"))

					Expect(outputBuffer).To(test_helpers.Say("Could not authorize target."))
				})

				It("does not save the config if there is an error connecting to the receptor after prompting", func() {
					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"newtarget.com"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					fakeTargetVerifier.VerifyTargetReturns(true, false, errors.New("Unknown Error"))
					stdinWriter.Write([]byte("notgood\n"))

					Eventually(doneChan).Should(BeClosed())

					Expect(fakePasswordReader.PromptForPasswordCallCount()).To(Equal(1))
					Expect(fakePasswordReader.PromptForPasswordArgsForCall(0)).To(Equal("Password"))

					Expect(outputBuffer).To(test_helpers.Say("Error verifying target: Unknown Error"))
				})
			})

			Context("when the receptor credentials work on the blob store", func() {
				It("saves the new blob target", func() {
					fakeBlobStoreVerifier.VerifyReturns(true, nil)

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
					stdinWriter.Write([]byte("testusername\n"))

					Eventually(doneChan).Should(BeClosed())

					Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
					Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(dav_blob_store.Config{
						Host:     "myapi.com",
						Port:     "8444",
						Username: "testusername",
						Password: "testpassword",
					}))

					newConfig := config_package.New(configPersister)
					Expect(newConfig.Load()).To(Succeed())
					Expect(newConfig.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
					Expect(newConfig.BlobTarget()).To(Equal(dav_blob_store.Config{
						Host:     "myapi.com",
						Port:     "8444",
						Username: "testusername",
						Password: "testpassword",
					}))

					Expect(outputBuffer).To(test_helpers.Say("Blob store is targeted."))
				})

				Context("when the receptor credentials don't work on the blob store", func() {
					It("exits", func() {
						fakeBlobStoreVerifier.VerifyReturns(false, nil)

						doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

						Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
						fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
						stdinWriter.Write([]byte("testusername\n"))

						Eventually(doneChan).Should(BeClosed())

						Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
						Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(dav_blob_store.Config{
							Host:     "myapi.com",
							Port:     "8444",
							Username: "testusername",
							Password: "testpassword",
						}))

						Expect(outputBuffer).To(test_helpers.Say("Invalid credentials for blob store."))
						verifyOldTargetStillSet()
						Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
					})
				})
			})

			Context("when the blob store is offline", func() {
				It("saves the receptor target but does not save any blob target", func() {
					fakeBlobStoreVerifier.VerifyReturns(false, errors.New("unable to connect to blob store"))

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(targetCommand, []string{"myapi.com"})

					Eventually(outputBuffer).Should(test_helpers.Say("Username: "))
					fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
					stdinWriter.Write([]byte("testusername\n"))

					Eventually(doneChan).Should(BeClosed())

					Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
					Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(dav_blob_store.Config{
						Host:     "myapi.com",
						Port:     "8444",
						Username: "testusername",
						Password: "testpassword",
					}))

					newConfig := config_package.New(configPersister)
					Expect(newConfig.Load()).To(Succeed())
					Expect(newConfig.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
					Expect(newConfig.BlobTarget()).To(Equal(dav_blob_store.Config{}))
				})
			})

		})

		Context("setting an invalid target", func() {
			It("does not save the config if the target verifier returns an error", func() {
				fakeTargetVerifier.VerifyTargetReturns(true, false, errors.New("Unknown Error"))

				test_helpers.ExecuteCommandWithArgs(targetCommand, []string{"newtarget.com"})

				Expect(outputBuffer).To(test_helpers.Say("Error verifying target: Unknown Error"))

				verifyOldTargetStillSet()
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
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
