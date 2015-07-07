package command_factory_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
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
		stdinReader        *io.PipeReader
		stdinWriter        *io.PipeWriter
		outputBuffer       *gbytes.Buffer
		terminalUI         terminal.UI
		config             *config_package.Config
		fakeTargetVerifier *fake_target_verifier.FakeTargetVerifier
		fakeExitHandler    *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		stdinReader, stdinWriter = io.Pipe()
		outputBuffer = gbytes.NewBuffer()
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		fakeExitHandler = new(fake_exit_handler.FakeExitHandler)
		terminalUI = terminal.NewUI(stdinReader, outputBuffer, nil)
		config = config_package.New(persister.NewMemPersister())
	})

	Describe("TargetBlobCommand", func() {
		var targetBlobCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewConfigCommandFactory(config, terminalUI, fakeTargetVerifier, fakeExitHandler)
			targetBlobCommand = commandFactory.MakeTargetBlobCommand()
		})

		Context("displaying the blob target", func() {
			It("outputs the current target", func() {
				config.SetBlobTarget("192.168.11.11", 8980, "datkeyyo", "supersecretJKJK", "bucket")
				config.Save()

				test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{})

				Expect(outputBuffer).To(test_helpers.Say("Blob Target:\t192.168.11.11:8980\n"))
				Expect(outputBuffer).To(test_helpers.Say("Access Key:\tdatkeyyo"))
				Expect(outputBuffer).To(test_helpers.Say("Secret Key:\tsupersecretJKJK"))
				Expect(outputBuffer).To(test_helpers.Say("Bucket Name:\tbucket"))
			})

			It("alerts the user if no target is set", func() {
				config.SetBlobTarget("", 0, "", "", "")
				config.Save()

				test_helpers.ExecuteCommandWithArgs(targetBlobCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Blob target not set"))
			})
		})

		Context("setting the blob target", func() {
			It("sets the blob target and credentials", func() {
				fakeTargetVerifier.VerifyBlobTargetReturns(true, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8980"})

				Eventually(outputBuffer).Should(test_helpers.Say("Access Key: "))
				stdinWriter.Write([]byte("yaykey\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Secret Key: "))
				stdinWriter.Write([]byte("superserial\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Bucket Name [condenser-bucket]: "))
				stdinWriter.Write([]byte("bhuket\n"))

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				host, port, accessKey, secretKey, bucketName := fakeTargetVerifier.VerifyBlobTargetArgsForCall(0)
				Expect(host).To(Equal("192.168.11.11"))
				Expect(port).To(Equal(uint16(8980)))
				Expect(accessKey).To(Equal("yaykey"))
				Expect(secretKey).To(Equal("superserial"))
				Expect(bucketName).To(Equal("bhuket"))

				blobTarget := config.BlobTarget()
				Expect(outputBuffer).To(test_helpers.Say("Blob Location Set"))
				Expect(blobTarget.TargetHost).To(Equal("192.168.11.11"))
				Expect(blobTarget.TargetPort).To(Equal(uint16(8980)))
				Expect(blobTarget.AccessKey).To(Equal("yaykey"))
				Expect(blobTarget.SecretKey).To(Equal("superserial"))
			})

			It("sets the blob target and credentials using the default bucket name", func() {
				fakeTargetVerifier.VerifyBlobTargetReturns(true, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8980"})

				Eventually(outputBuffer).Should(test_helpers.Say("Access Key: "))
				stdinWriter.Write([]byte("yaykey\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Secret Key: "))
				stdinWriter.Write([]byte("superserial\n"))
				Eventually(outputBuffer).Should(test_helpers.Say("Bucket Name [condenser-bucket]: "))
				stdinWriter.Write([]byte("\n"))

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				host, port, accessKey, secretKey, bucketName := fakeTargetVerifier.VerifyBlobTargetArgsForCall(0)
				Expect(host).To(Equal("192.168.11.11"))
				Expect(port).To(Equal(uint16(8980)))
				Expect(accessKey).To(Equal("yaykey"))
				Expect(secretKey).To(Equal("superserial"))
				Expect(bucketName).To(Equal("condenser-bucket"))

				blobTarget := config.BlobTarget()
				Expect(outputBuffer).To(test_helpers.Say("Blob Location Set"))
				Expect(blobTarget.TargetHost).To(Equal("192.168.11.11"))
				Expect(blobTarget.TargetPort).To(Equal(uint16(8980)))
				Expect(blobTarget.AccessKey).To(Equal("yaykey"))
				Expect(blobTarget.SecretKey).To(Equal("superserial"))
			})

			Context("invalid syntax", func() {
				It("errors when target is malformed", func() {
					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"huehue8980"})

					Eventually(commandFinishChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed target"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(BeZero())
				})
				It("errors when port is non-numeric", func() {
					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:haiii"})

					Eventually(commandFinishChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed port"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(BeZero())
				})
				It("errors when port exceeds 65536", func() {
					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:70000"})

					Eventually(commandFinishChan).Should(BeClosed())
					Expect(outputBuffer).To(test_helpers.SayLine("Error setting blob target: malformed port"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(BeZero())
				})
			})

			Context("scenarios that should not save the config", func() {
				verifyConfigNotSaved := func(failMessage string) {
					Expect(outputBuffer).NotTo(test_helpers.Say("Blob Location Set"))
					Expect(outputBuffer).To(test_helpers.Say(failMessage))
					config.Load()
					blobTarget := config.BlobTarget()
					Expect(blobTarget.TargetHost).To(Equal("original-host"))
					Expect(blobTarget.TargetPort).To(Equal(uint16(8989)))
					Expect(blobTarget.AccessKey).To(Equal("original-key"))
					Expect(blobTarget.SecretKey).To(Equal("original-secret"))
					Expect(blobTarget.BucketName).To(Equal("original-bucket"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadTarget}))
				}

				BeforeEach(func() {
					config.SetBlobTarget("original-host", 8989, "original-key", "original-secret", "original-bucket")
					config.Save()
				})

				It("does not save the config when there is an error connecting to the target", func() {
					fakeTargetVerifier.VerifyBlobTargetReturns(false, errors.New("fail"))

					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"192.168.11.11:8980"})

					Eventually(outputBuffer).Should(test_helpers.Say("Access Key: "))
					stdinWriter.Write([]byte("yaykey\n"))
					Eventually(outputBuffer).Should(test_helpers.Say("Secret Key: "))
					stdinWriter.Write([]byte("superserial\n"))
					Eventually(outputBuffer).Should(test_helpers.Say("Bucket Name [condenser-bucket]: "))
					stdinWriter.Write([]byte("buckit\n"))

					Eventually(commandFinishChan).Should(BeClosed())

					Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))

					verifyConfigNotSaved("Unable to verify blob store: fail")
				})

				Context("when the persister returns errors", func() {
					BeforeEach(func() {
						fakeTargetVerifier.VerifyBlobTargetReturns(true, nil)
						commandFactory := command_factory.NewConfigCommandFactory(
							config_package.New(errorPersister("Failure setting blob target")),
							terminalUI,
							fakeTargetVerifier,
							fakeExitHandler,
						)
						targetBlobCommand = commandFactory.MakeTargetBlobCommand()
					})
					It("bubbles up errors from saving the config", func() {
						config.SetBlobTarget("192.168.11.11", 8980, "datkeyyo", "supersecretJKJK", "buckit")
						config.Save()

						commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(targetBlobCommand, []string{"199.112.3432:8980"})
						Eventually(outputBuffer).Should(test_helpers.Say("Access Key: "))
						stdinWriter.Write([]byte("booookey\n"))
						Eventually(outputBuffer).Should(test_helpers.Say("Secret Key: "))
						stdinWriter.Write([]byte("unicorns\n"))
						Eventually(outputBuffer).Should(test_helpers.Say("Bucket Name [condenser-bucket]: "))
						stdinWriter.Write([]byte("bkkt\n"))

						Eventually(commandFinishChan).Should(BeClosed())

						Expect(outputBuffer).To(test_helpers.Say("Failure setting blob target"))
						config.Load()

						blobTarget := config.BlobTarget()
						Expect(blobTarget.TargetHost).To(Equal("192.168.11.11"))
						Expect(blobTarget.TargetPort).To(Equal(uint16(8980)))
						Expect(blobTarget.AccessKey).To(Equal("datkeyyo"))
						Expect(blobTarget.SecretKey).To(Equal("supersecretJKJK"))
						Expect(blobTarget.BucketName).To(Equal("buckit"))

						Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
						Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
					})
				})
			})
		})

	})
})
