package command_factory_test

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("CommandFactory", func() {
	var (
		stdinReader     *io.PipeReader
		stdinWriter     *io.PipeWriter
		outputBuffer    *gbytes.Buffer
		terminalUI      terminal.UI
		config          *config_package.Config
		fakeExitHandler *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		stdinReader, stdinWriter = io.Pipe()
		outputBuffer = gbytes.NewBuffer()
		fakeExitHandler = new(fake_exit_handler.FakeExitHandler)
		terminalUI = terminal.NewUI(stdinReader, outputBuffer, nil)
		config = config_package.New(persister.NewMemPersister())
	})

	Describe("TargetBlobCommand", func() {
		var targetBlobCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewConfigCommandFactory(config, terminalUI, nil, fakeExitHandler)
			targetBlobCommand = commandFactory.MakeTargetBlobCommand()
		})

		Context("displaying the blob target", func() {
			It("outputs the current target", func() {
				config.SetTargetBlob("192.168.11.11", 8181, "datkeyyo", "supersecretJKJK")
				config.Save()

				test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Blob Target:\t192.168.11.11:8181\n"))
				Expect(outputBuffer).To(test_helpers.Say("Access Key:\tdatkeyyo"))
				Expect(outputBuffer).To(test_helpers.Say("Secret Key:\tsupersecretJKJK"))
			})

			It("alerts the user if no target is set", func() {
				config.SetTargetBlob("", 0, "", "")
				config.Save()

				test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Blob target not set"))
			})
		})

		Context("setting the blob target", func() {
			It("sets the blob target and credentials", func() {
				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8181"})

				Eventually(outputBuffer).Should(test_helpers.Say("Access Key: "))
				stdinWriter.Write([]byte("yaykey\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Secret Key: "))
				stdinWriter.Write([]byte("superserial\n"))

				Eventually(commandFinishChan).Should(BeClosed())

				blobTarget := config.BlobTarget()
				Expect(outputBuffer).To(test_helpers.Say("Blob Location Set"))
				Expect(blobTarget.TargetHost).To(Equal("192.168.11.11"))
				Expect(blobTarget.TargetPort).To(Equal(uint16(8181)))
				Expect(blobTarget.AccessKey).To(Equal("yaykey"))
				Expect(blobTarget.SecretKey).To(Equal("superserial"))
			})

			Context("invalid syntax", func() {
				It("errors when target is malformed", func() {
					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"huehue8181"})

					Eventually(commandFinishChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed target"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})
				It("errors when port is non-numeric", func() {
					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:haiii"})

					Eventually(commandFinishChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed port"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})
				It("errors when port exceeds 65536", func() {
					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:70000"})

					Eventually(commandFinishChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed port"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})
			})

			Context("scenarios that should not save the config", func() {
				// It("does not save the config when target can't be authorized", func() {})

				// It("does not save the config when there is an error connecting to the target", func() {})

				Context("when the persister returns errors", func() {
					BeforeEach(func() {
						commandFactory := command_factory.NewConfigCommandFactory(
							config_package.New(errorPersister("Failure setting blob target")),
							terminalUI,
							nil,
							fakeExitHandler,
						)
						targetBlobCommand = commandFactory.MakeTargetBlobCommand()
					})
					It("bubbles up errors from saving the config", func() {
						config.SetTargetBlob("192.168.11.11", 8181, "datkeyyo", "supersecretJKJK")

						commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"199.112.3432:8181"})
						Eventually(outputBuffer).Should(test_helpers.Say("Access Key: "))
						stdinWriter.Write([]byte("booookey\n"))
						Eventually(outputBuffer).Should(test_helpers.Say("Secret Key: "))
						stdinWriter.Write([]byte("unicorns\n"))

						Eventually(commandFinishChan).Should(BeClosed())

						Eventually(outputBuffer).Should(test_helpers.Say("Failure setting blob target"))
						config.Load()

						blobTarget := config.BlobTarget()
						Expect(blobTarget.TargetHost).To(Equal("192.168.11.11"))
						Expect(blobTarget.TargetPort).To(Equal(uint16(8181)))
						Expect(blobTarget.AccessKey).To(Equal("datkeyyo"))
						Expect(blobTarget.SecretKey).To(Equal("supersecretJKJK"))

						Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))

					})
				})
			})
		})

	})
})
