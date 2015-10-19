package command_factory_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore/fake_cf_ignore"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/fake_blob_store_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/zipper/fake_zipper"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/fake_droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/fake_task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"

	app_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	droplet_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory"
)

var _ = Describe("CommandFactory", func() {
	var (
		outputBuffer            *gbytes.Buffer
		fakeDropletRunner       *fake_droplet_runner.FakeDropletRunner
		fakeExitHandler         *fake_exit_handler.FakeExitHandler
		fakeTailedLogsOutputter *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		fakeClock               *fakeclock.FakeClock
		fakeAppExaminer         *fake_app_examiner.FakeAppExaminer
		fakeTaskExaminer        *fake_task_examiner.FakeTaskExaminer
		fakeCFIgnore            *fake_cf_ignore.FakeCFIgnore
		fakeZipper              *fake_zipper.FakeZipper
		fakeBlobStoreVerifier   *fake_blob_store_verifier.FakeBlobStoreVerifier
		config                  *config_package.Config

		appRunnerCommandFactory app_runner_command_factory.AppRunnerCommandFactory
	)

	BeforeEach(func() {
		fakeDropletRunner = &fake_droplet_runner.FakeDropletRunner{}
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		fakeClock = fakeclock.NewFakeClock(time.Now())
		fakeAppExaminer = &fake_app_examiner.FakeAppExaminer{}
		fakeTaskExaminer = &fake_task_examiner.FakeTaskExaminer{}
		fakeCFIgnore = &fake_cf_ignore.FakeCFIgnore{}
		fakeZipper = &fake_zipper.FakeZipper{}
		fakeBlobStoreVerifier = &fake_blob_store_verifier.FakeBlobStoreVerifier{}
		config = config_package.New(nil)

		outputBuffer = gbytes.NewBuffer()
		appRunnerCommandFactory = app_runner_command_factory.AppRunnerCommandFactory{
			AppRunner:           &fake_app_runner.FakeAppRunner{},
			AppExaminer:         fakeAppExaminer,
			UI:                  terminal.NewUI(nil, outputBuffer, nil),
			ExitHandler:         fakeExitHandler,
			TailedLogsOutputter: fakeTailedLogsOutputter,
			Clock:               fakeClock,
			Domain:              "192.168.11.11.xip.io",
			Env:                 []string{"SHELL=/bin/bash", "COLOR=Black", "AAAA=xyz"},
		}
	})

	Describe("BuildDropletCommand", func() {
		var buildDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeBlobStoreVerifier, fakeTaskExaminer, fakeDropletRunner, fakeCFIgnore, fakeZipper, config)
			buildDropletCommand = commandFactory.MakeBuildDropletCommand()
			fakeBlobStoreVerifier.VerifyReturns(true, nil)
		})

		Context("when the archive path is a folder and exists", func() {
			BeforeEach(func() {
				fakeCFIgnore.ShouldIgnoreStub = func(path string) bool {
					return path == "some-ignored-file"
				}
			})

			It("zips up current working folder and uploads as the droplet name", func() {
				fakeZipper.IsZipFileReturns(false)
				fakeZipper.ZipReturns("xyz.zip", nil)

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.SayLine("Uploading application bits..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Uploaded."))

				Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
				Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(config))

				Expect(outputBuffer).To(test_helpers.SayLine("Submitted build of droplet-name"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				dropletName, uploadPath := fakeDropletRunner.UploadBitsArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))

				Expect(uploadPath).NotTo(BeNil())
				Expect(uploadPath).To(Equal("xyz.zip"))

				Expect(fakeZipper.ZipCallCount()).To(Equal(1))
				_, cfIgnore := fakeZipper.ZipArgsForCall(0)
				Expect(cfIgnore).To(Equal(fakeCFIgnore))
			})

			It("re-zips an existing .zip passed to -p and uploads as the droplet name", func() {
				fakeZipper.IsZipFileReturns(true)
				fakeZipper.UnzipReturns(nil)
				fakeZipper.ZipReturns("xyz.zip", nil)

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack", "-p", "abc.zip"})

				Expect(fakeZipper.IsZipFileCallCount()).To(Equal(1))
				Expect(fakeZipper.IsZipFileArgsForCall(0)).To(Equal("abc.zip"))
				Expect(fakeZipper.UnzipCallCount()).To(Equal(1))
				src, dest := fakeZipper.UnzipArgsForCall(0)
				Expect(src).To(Equal("abc.zip"))
				Expect(fakeZipper.ZipCallCount()).To(Equal(1))
				zip, cfIgnore := fakeZipper.ZipArgsForCall(0)
				Expect(zip).To(Equal(dest))
				Expect(cfIgnore).To(Equal(fakeCFIgnore))

				Expect(outputBuffer).To(test_helpers.SayLine("Uploading application bits..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Uploaded."))

				Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
				Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(config))

				Expect(outputBuffer).To(test_helpers.SayLine("Submitted build of droplet-name"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				dropletName, uploadPath := fakeDropletRunner.UploadBitsArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))

				Expect(uploadPath).NotTo(BeNil())
				Expect(uploadPath).To(Equal("xyz.zip"))
			})

			It("passes through environment variables from the command-line", func() {
				args := []string{
					"-e",
					"AAAA",
					"-e",
					"BBBB=2",
					"droplet-name",
					"http://some.url/for/buildpack",
				}
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Submitted build of droplet-name"))

				_, _, _, envVars, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)

				aaaaVar, found := envVars["AAAA"]
				Expect(found).To(BeTrue())
				Expect(aaaaVar).To(Equal("xyz"))
				bbbbVar, found := envVars["BBBB"]
				Expect(found).To(BeTrue())
				Expect(bbbbVar).To(Equal("2"))
			})

			It("allows specifying resource parameters on the command-line", func() {
				args := []string{
					"-c",
					"75",
					"-m",
					"512",
					"-d",
					"800",
					"droplet-name",
					"http://some.url/for/buildpack",
				}
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Submitted build of droplet-name"))

				_, _, _, _, mem, cpu, disk := fakeDropletRunner.BuildDropletArgsForCall(0)
				Expect(cpu).To(Equal(75))
				Expect(mem).To(Equal(512))
				Expect(disk).To(Equal(800))
			})

			Describe("buildpack aliases", func() {
				It("uses the correct buildpack URL for go", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "go"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/go-buildpack.git"))
				})

				It("uses the correct buildpack URL for java", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "java"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/java-buildpack.git"))
				})

				It("uses the correct buildpack URL for python", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "python"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/python-buildpack.git"))
				})

				It("uses the correct buildpack URL for ruby", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "ruby"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/ruby-buildpack.git"))
				})

				It("uses the correct buildpack URL for nodejs", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "nodejs"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/nodejs-buildpack.git"))
				})

				It("uses the correct buildpack URL for php", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "php"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/php-buildpack.git"))
				})

				It("uses the correct buildpack URL for binary", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "binary"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/binary-buildpack.git"))
				})

				It("uses the correct buildpack URL for staticfile", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "staticfile"})
					_, _, buildpackUrl, _, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
					Expect(buildpackUrl).To(Equal("https://github.com/cloudfoundry/staticfile-buildpack.git"))
				})

				It("rejects unknown buildpack alias or unparseable URL", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "¥¥¥¥://¥¥¥¥¥¥¥¥"})

					Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: invalid buildpack ¥¥¥¥://¥¥¥¥¥¥¥¥"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
					Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
					Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
				})
			})
		})

		Context("when the blob store cannot be verified", func() {
			It("prints the error and stops when verification fails", func() {
				fakeBlobStoreVerifier.VerifyReturns(false, errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error verifying droplet store: failed"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
			})

			It("prints the error and stops when unauthorized", func() {
				fakeBlobStoreVerifier.VerifyReturns(false, nil)

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error verifying droplet store: unauthorized"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
			})
		})

		Context("when the zipper returns an error", func() {
			It("prints the error from Unzip", func() {
				fakeZipper.IsZipFileReturns(true)
				fakeZipper.UnzipReturns(errors.New("oop"))
				fakeZipper.ZipReturns("xyz.zip", nil)

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack", "-p", "abc.zip"})

				Expect(fakeZipper.IsZipFileCallCount()).To(Equal(1))
				Expect(fakeZipper.IsZipFileArgsForCall(0)).To(Equal("abc.zip"))
				Expect(fakeZipper.UnzipCallCount()).To(Equal(1))
				Expect(fakeZipper.ZipCallCount()).To(Equal(0))

				Expect(outputBuffer).To(test_helpers.SayLine("Error unarchiving abc.zip: oop"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))

				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
			})

			It("prints the error from re-Zip", func() {
				fakeZipper.IsZipFileReturns(true)
				fakeZipper.UnzipReturns(nil)
				fakeZipper.ZipReturns("", errors.New("oop"))

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack", "-p", "abc.zip"})

				Expect(fakeZipper.IsZipFileCallCount()).To(Equal(1))
				Expect(fakeZipper.IsZipFileArgsForCall(0)).To(Equal("abc.zip"))
				Expect(fakeZipper.UnzipCallCount()).To(Equal(1))
				Expect(fakeZipper.ZipCallCount()).To(Equal(1))

				Expect(outputBuffer).To(test_helpers.SayLine("Error re-archiving abc.zip: oop"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))

				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
			})
		})

		Context("when the droplet runner returns an error", func() {
			It("prints the error from upload bits", func() {
				fakeDropletRunner.UploadBitsReturns(errors.New("uploading bits failed"))

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error uploading droplet-name: uploading bits failed"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
			})

			It("prints the error from build droplet", func() {
				fakeDropletRunner.BuildDropletReturns(errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error submitting build of droplet-name: failed"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Describe("waiting for the build to finish", func() {
			It("polls for the build to complete, outputting logs while the build runs", func() {
				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "PENDING"}, nil)

				args := []string{
					"droplet-name",
					"http://some.url/for/buildpack",
				}
				doneChan := test_helpers.AsyncExecuteCommandWithArgs(buildDropletCommand, args)

				Eventually(outputBuffer).Should(test_helpers.SayLine("Submitted build of droplet-name"))

				Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("build-droplet-droplet-name"))

				Eventually(fakeTaskExaminer.TaskStatusCallCount).Should(Equal(1))
				Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("build-droplet-droplet-name"))

				fakeClock.IncrementBySeconds(1)
				Expect(doneChan).NotTo(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "RUNNING"}, nil)

				fakeClock.IncrementBySeconds(1)
				Expect(doneChan).NotTo(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "COMPLETED"}, nil)

				fakeClock.IncrementBySeconds(1)
				Eventually(doneChan, 3).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.SayLine("Build completed"))
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
			})

			Context("when the build doesn't complete before the timeout elapses", func() {
				It("alerts the user the build took too long", func() {
					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "RUNNING"}, nil)

					args := []string{
						"droppo-the-clown",
						"http://some.url/for/buildpack",
						"-t",
						"17s",
					}
					doneChan := test_helpers.AsyncExecuteCommandWithArgs(buildDropletCommand, args)

					Eventually(outputBuffer).Should(test_helpers.SayLine("Submitted build of droppo-the-clown"))

					fakeClock.IncrementBySeconds(17)

					Eventually(doneChan, 5).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.SayLine(colors.Red("Timed out waiting for the build to complete.")))
					Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still building your application in the background."))
					Expect(outputBuffer).To(test_helpers.SayLine("To view logs:"))
					Expect(outputBuffer).To(test_helpers.SayLine("ltc logs build-droplet-droppo-the-clown"))
					Expect(outputBuffer).To(test_helpers.SayLine("To view status:"))
					Expect(outputBuffer).To(test_helpers.SayLine("ltc status build-droplet-droppo-the-clown"))
				})
			})

			Context("when the build completes", func() {
				It("alerts the user of a complete but failed build", func() {
					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "PENDING"}, nil)

					args := []string{"droppo-the-clown", "http://some.url/for/buildpack"}
					doneChan := test_helpers.AsyncExecuteCommandWithArgs(buildDropletCommand, args)

					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{
						State:         "COMPLETED",
						Failed:        true,
						FailureReason: "oops",
					}, nil)

					fakeClock.IncrementBySeconds(1)

					Eventually(doneChan, 3).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.SayLine("Build failed: oops"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})

			Context("when there is an error when polling for the build to complete", func() {
				It("prints an error message and exits", func() {
					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{}, errors.New("dropped the ball"))

					args := []string{
						"droppo-the-clown",
						"http://some.url/for/buildpack",
					}
					doneChan := test_helpers.AsyncExecuteCommandWithArgs(buildDropletCommand, args)

					Eventually(outputBuffer).Should(test_helpers.SayLine("Submitted build of droppo-the-clown"))

					Eventually(fakeTaskExaminer.TaskStatusCallCount).Should(Equal(1))
					Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("build-droplet-droppo-the-clown"))

					fakeClock.IncrementBySeconds(1)
					Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

					fakeClock.IncrementBySeconds(1)
					Eventually(doneChan, 3).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.SayLine(colors.Red("Error requesting task status: dropped the ball")))
					Expect(outputBuffer).NotTo(test_helpers.SayLine("Timed out waiting for the build to complete."))
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
				})
			})
		})

		Context("invalid syntax", func() {
			It("rejects less than two positional arguments", func() {
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name"})

				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: DROPLET_NAME and BUILDPACK_URL are required"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("tests for an empty droplet name", func() {
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"", "buildpack-name"})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates cpuWeight is between 1 and 100", func() {
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"-c", "9999", "droplet-name", "java"})

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: invalid CPU Weight"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})

	Describe("ListDropletsCommand", func() {
		var listDropletsCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeBlobStoreVerifier, fakeTaskExaminer, fakeDropletRunner, nil, fakeZipper, config)
			listDropletsCommand = commandFactory.MakeListDropletsCommand()
			fakeBlobStoreVerifier.VerifyReturns(true, nil)
		})

		It("lists the droplets most recent first", func() {
			times := []time.Time{
				time.Date(2014, 12, 31, 8, 22, 44, 0, time.Local),
				time.Date(2015, 6, 15, 16, 11, 33, 0, time.Local),
			}
			droplets := []droplet_runner.Droplet{
				{Name: "drop-a", Created: times[0], Size: 789 * 1024 * 1024},
				{Name: "drop-b", Created: times[1], Size: 456 * 1024},
			}
			fakeDropletRunner.ListDropletsReturns(droplets, nil)

			test_helpers.ExecuteCommandWithArgs(listDropletsCommand, []string{})

			Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
			Expect(fakeBlobStoreVerifier.VerifyArgsForCall(0)).To(Equal(config))

			Expect(outputBuffer).To(test_helpers.SayLine("Droplet\t\tCreated At\t\tSize"))
			Expect(outputBuffer).To(test_helpers.SayLine("drop-b\t\t06/15 16:11:33.00\t456K"))
			Expect(outputBuffer).To(test_helpers.SayLine("drop-a\t\t12/31 08:22:44.00\t789M"))
		})

		It("doesn't print a time if Created is nil", func() {
			time := time.Date(2014, 12, 31, 14, 33, 52, 0, time.Local)

			droplets := []droplet_runner.Droplet{
				{Name: "drop-a", Created: time, Size: 789 * 1024 * 1024},
				{Name: "drop-b", Size: 456 * 1024},
			}
			fakeDropletRunner.ListDropletsReturns(droplets, nil)

			test_helpers.ExecuteCommandWithArgs(listDropletsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.SayLine("Droplet\t\tCreated At\t\tSize"))
			Expect(outputBuffer).To(test_helpers.SayLine("drop-b\t\t\t\t\t456K"))
			Expect(outputBuffer).To(test_helpers.SayLine("drop-a\t\t12/31 14:33:52.00\t789M"))
		})

		Context("when the droplet runner returns errors", func() {
			It("prints an error", func() {
				fakeDropletRunner.ListDropletsReturns(nil, errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(listDropletsCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error listing droplets: failed"))
				Expect(fakeDropletRunner.ListDropletsCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when the blob store cannot be verified", func() {
			It("prints the error and stops when verification fails", func() {
				fakeBlobStoreVerifier.VerifyReturns(false, errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(listDropletsCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error verifying droplet store: failed"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
			})

			It("prints the error and stops when unauthorized", func() {
				fakeBlobStoreVerifier.VerifyReturns(false, nil)

				test_helpers.ExecuteCommandWithArgs(listDropletsCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error verifying droplet store: unauthorized"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				Expect(fakeBlobStoreVerifier.VerifyCallCount()).To(Equal(1))
			})
		})

	})

	Describe("LaunchDropletCommand", func() {
		var launchDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeBlobStoreVerifier, fakeTaskExaminer, fakeDropletRunner, nil, fakeZipper, config)
			launchDropletCommand = commandFactory.MakeLaunchDropletCommand()
		})

		Context("when a malformed tcp route passed", func() {
			It("errors out", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--tcp-route=woo:50000",
					"--",
					"/start-me-please",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.InvalidPortErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		It("launches the specified droplet with tcp routes", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
			fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{
				Routes: route_helpers.Routes{
					AppRoutes: []route_helpers.AppRoute{
						{
							Hostnames: []string{"ninetyninety.192.168.11.11.xip.io"},
							Port:      4444,
						},
						{
							Hostnames: []string{"fourtyfourfourtyfour.192.168.11.11.xip.io"},
							Port:      9090,
						},
					},
				},
			}, nil)

			args := []string{
				"--ports=4444",
				"--http-route=ninetyninety",
				"--http-route=fourtyfourfourtyfour:9090",
				"--tcp-route=50000",
				"--tcp-route=50001:5223",
				"droppy",
				"droplet-name",
				"--",
				"start-em",
			}
			test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

			Expect(outputBuffer).To(test_helpers.SayLine("Creating App: droppy"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("droppy is now running.")))
			Expect(outputBuffer).To(test_helpers.SayLine("App is reachable at:"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://ninetyninety.192.168.11.11.xip.io")))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://fourtyfourfourtyfour.192.168.11.11.xip.io")))

			Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(1))
			appName, dropletNameParam, startCommandParam, _, appEnvParam := fakeDropletRunner.LaunchDropletArgsForCall(0)
			Expect(appName).To(Equal("droppy"))
			Expect(dropletNameParam).To(Equal("droplet-name"))
			Expect(startCommandParam).To(Equal("start-em"))
			Expect(appEnvParam.Instances).To(Equal(1))
			Expect(appEnvParam.NoRoutes).To(BeFalse())
			Expect(appEnvParam.RouteOverrides).To(ContainExactly(app_runner.RouteOverrides{
				{HostnamePrefix: "ninetyninety", Port: 4444},
				{HostnamePrefix: "fourtyfourfourtyfour", Port: 9090},
			}))
			Expect(appEnvParam.TcpRoutes).To(ContainExactly(app_runner.TcpRoutes{
				{ExternalPort: 50000, Port: 4444},
				{ExternalPort: 50001, Port: 5223},
			}))
		})

		It("launches the specified droplet", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(11, false, nil)
			fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{
				Routes: route_helpers.Routes{
					AppRoutes: []route_helpers.AppRoute{
						{
							Hostnames: []string{"ninetyninety.192.168.11.11.xip.io"},
							Port:      4444,
						},
						{
							Hostnames: []string{"fourtyfourfourtyfour.192.168.11.11.xip.io"},
							Port:      9090,
						},
					},
				},
			}, nil)

			args := []string{
				"--cpu-weight=57",
				"--memory-mb=12",
				"--disk-mb=12",
				"--http-route=ninetyninety:4444",
				"--http-route=fourtyfourfourtyfour:9090",
				"--instances=11",
				"--env=TIMEZONE=CST",
				`--env=LANG="Chicago English"`,
				"--env=COLOR",
				"--env=UNSET",
				"--monitor-timeout=4s",
				"--ports=8081",
				"droppy",
				"droplet-name",
				"--",
				"start-em",
				"-app-arg",
			}
			test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

			Expect(outputBuffer).To(test_helpers.SayLine("Creating App: droppy"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("droppy is now running.")))
			Expect(outputBuffer).To(test_helpers.SayLine("App is reachable at:"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://ninetyninety.192.168.11.11.xip.io")))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://fourtyfourfourtyfour.192.168.11.11.xip.io")))

			Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(1))
			appName, dropletNameParam, startCommandParam, startArgsParam, appEnvParam := fakeDropletRunner.LaunchDropletArgsForCall(0)
			Expect(appName).To(Equal("droppy"))
			Expect(dropletNameParam).To(Equal("droplet-name"))
			Expect(startCommandParam).To(Equal("start-em"))
			Expect(startArgsParam).To(Equal([]string{"-app-arg"}))
			Expect(appEnvParam.CPUWeight).To(Equal(uint(57)))
			Expect(appEnvParam.MemoryMB).To(Equal(12))
			Expect(appEnvParam.DiskMB).To(Equal(12))
			Expect(appEnvParam.Privileged).To(BeFalse())
			Expect(appEnvParam.Instances).To(Equal(11))
			Expect(appEnvParam.NoRoutes).To(BeFalse())
			Expect(appEnvParam.Monitor).To(Equal(app_runner.MonitorConfig{
				Method:  app_runner.PortMonitor,
				Port:    8081,
				Timeout: 4 * time.Second,
			}))
			Expect(appEnvParam.EnvironmentVariables).To(Equal(map[string]string{
				"PROCESS_GUID": "droppy",
				"TIMEZONE":     "CST",
				"LANG":         `"Chicago English"`,
				"COLOR":        "Black",
				"UNSET":        "",
				"MEMORY_LIMIT": "12M",
			}))
			Expect(appEnvParam.RouteOverrides).To(ContainExactly(app_runner.RouteOverrides{
				{HostnamePrefix: "ninetyninety", Port: 4444},
				{HostnamePrefix: "fourtyfourfourtyfour", Port: 9090},
			}))
		})

		It("launches the specified droplet with default values", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
			fakeAppExaminer.AppStatusReturns(app_examiner.AppInfo{
				Routes: route_helpers.Routes{
					AppRoutes: []route_helpers.AppRoute{
						{
							Hostnames: []string{"droppy.192.168.11.11.xip.io"},
							Port:      8080,
						},
					},
				},
			}, nil)

			test_helpers.ExecuteCommandWithArgs(launchDropletCommand, []string{"droppy", "droplet-name"})

			Expect(outputBuffer).To(test_helpers.SayLine("No port specified. Defaulting to 8080."))
			Expect(outputBuffer).To(test_helpers.SayLine("Creating App: droppy"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("droppy is now running.")))
			Expect(outputBuffer).To(test_helpers.SayLine("App is reachable at:"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://droppy.192.168.11.11.xip.io")))

			Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(1))
			appName, dropletNameParam, startCommandParam, startArgsParam, appEnvParam := fakeDropletRunner.LaunchDropletArgsForCall(0)
			Expect(appName).To(Equal("droppy"))
			Expect(dropletNameParam).To(Equal("droplet-name"))
			Expect(startCommandParam).To(Equal(""))
			Expect(startArgsParam).To(BeNil())
			Expect(appEnvParam.Privileged).To(BeFalse())
			Expect(appEnvParam.User).To(Equal("vcap"))
			Expect(appEnvParam.Instances).To(Equal(1))
			Expect(appEnvParam.Monitor).To(Equal(app_runner.MonitorConfig{
				Method:  app_runner.PortMonitor,
				Port:    8080,
				Timeout: 1 * time.Second,
			}))
			Expect(appEnvParam.EnvironmentVariables).To(Equal(map[string]string{
				"PROCESS_GUID": "droppy",
				"MEMORY_LIMIT": "128M",
			}))
			Expect(appEnvParam.RouteOverrides).To(BeNil())
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{"appy"}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: APP_NAME and DROPLET_NAME are required"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the terminator -- is passed in when a start command is specified", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"not-the-terminator",
					"start-me-up",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: '--' Required before start command"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates the CPU weight is in 1-100", func() {
				args := []string{
					"cool-app",
					"cool-droplet",
					"--cpu-weight=0",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Incorrect Usage: invalid CPU Weight"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when a malformed routes flag is passed", func() {
			It("errors out when the port is not an int", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--http-route=woo:aahh",
				}

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.InvalidPortErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when a malformed ports flag is passed", func() {
			It("errors out when the port is not an int", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--ports=kablowww",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.InvalidPortErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when a bad monitor config is passed", func() {
			It("monitor url is malformed", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--monitor-url=8080",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.InvalidPortErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("monitor url has invalid port", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--monitor-url=port:path",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.InvalidPortErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("monitor url port isn't exposed", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--ports=9090",
					"--monitor-url=8080:/path",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.MonitorPortNotExposed))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("monitor port isn't exposed", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--ports=9090",
					"--monitor-port=8080",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.SayLine(app_runner_command_factory.MonitorPortNotExposed))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when the droplet runner returns errors", func() {
			It("prints an error", func() {
				fakeDropletRunner.LaunchDropletReturns(errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, []string{"droppy", "droplet-name"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error launching app droppy from droplet droplet-name: failed"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})

	Describe("RemoveDropletCommand", func() {
		var removeDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeBlobStoreVerifier, fakeTaskExaminer, fakeDropletRunner, nil, fakeZipper, config)
			removeDropletCommand = commandFactory.MakeRemoveDropletCommand()
		})

		It("removes the droplet", func() {
			test_helpers.ExecuteCommandWithArgs(removeDropletCommand, []string{"droppo"})

			Expect(outputBuffer).To(test_helpers.SayLine("Droplet removed"))
			Expect(fakeDropletRunner.RemoveDropletCallCount()).To(Equal(1))
			Expect(fakeDropletRunner.RemoveDropletArgsForCall(0)).To(Equal("droppo"))
		})

		Context("when the droplet runner returns errors", func() {
			It("prints an error", func() {
				fakeDropletRunner.RemoveDropletReturns(errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(removeDropletCommand, []string{"droppo"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error removing droplet droppo: failed"))
				Expect(fakeDropletRunner.RemoveDropletCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
		Context("when the required arguments are missing", func() {
			It("prints an error", func() {
				test_helpers.ExecuteCommandWithArgs(removeDropletCommand, []string{""})
				Expect(outputBuffer).To(test_helpers.SayLine("DROPLET_NAME is required"))
				Expect(fakeDropletRunner.RemoveDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})

	Describe("ExportDropletCommand", func() {
		var (
			exportDropletCommand           cli.Command
			exportDir, workingDir, prevDir string
		)

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeBlobStoreVerifier, fakeTaskExaminer, fakeDropletRunner, nil, fakeZipper, config)
			exportDropletCommand = commandFactory.MakeExportDropletCommand()

		})

		BeforeEach(func() {
			var err error
			exportDir, err = ioutil.TempDir(os.TempDir(), "exported_stuff")
			Expect(err).NotTo(HaveOccurred())

			Expect(ioutil.WriteFile(filepath.Join(exportDir, "droppo.tgz"), []byte("tar"), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(exportDir, "droppo-metadata.json"), []byte("json"), 0644)).To(Succeed())

			workingDir, err = ioutil.TempDir("", "working_dir")
			Expect(err).NotTo(HaveOccurred())
			prevDir, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			Expect(os.Chdir(workingDir)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Chdir(prevDir)).To(Succeed())

			Expect(os.RemoveAll(exportDir)).To(Succeed())
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		It("exports the droplet", func() {
			dropletReader, err := os.Open(filepath.Join(exportDir, "droppo.tgz"))
			Expect(err).NotTo(HaveOccurred())
			defer dropletReader.Close()

			metadataReader, err := os.Open(filepath.Join(exportDir, "droppo-metadata.json"))
			Expect(err).NotTo(HaveOccurred())
			defer metadataReader.Close()

			fakeDropletRunner.ExportDropletReturns(dropletReader, metadataReader, nil)

			test_helpers.ExecuteCommandWithArgs(exportDropletCommand, []string{"droppo"})

			Expect(outputBuffer).To(test_helpers.SayLine("Droplet 'droppo' exported to droppo.tgz and droppo-metadata.json."))
			Expect(fakeDropletRunner.ExportDropletCallCount()).To(Equal(1))
			Expect(fakeDropletRunner.ExportDropletArgsForCall(0)).To(Equal("droppo"))

			Expect(os.Stat("droppo.tgz")).NotTo(BeNil())
			Expect(os.Stat("droppo-metadata.json")).NotTo(BeNil())
		})

		Context("when the droplet runner returns errors", func() {
			It("prints an error", func() {
				fakeDropletRunner.ExportDropletReturns(nil, nil, errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(exportDropletCommand, []string{"droppo"})

				Expect(outputBuffer).To(test_helpers.SayLine("Error exporting droplet droppo: failed"))
				Expect(fakeDropletRunner.ExportDropletCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when the required arguments are missing", func() {
			It("prints incorrect usage", func() {
				test_helpers.ExecuteCommandWithArgs(exportDropletCommand, []string{""})

				Expect(outputBuffer).To(test_helpers.SayLine("DROPLET_NAME is required"))
				Expect(fakeDropletRunner.ExportDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})

	Describe("ImportDropletCommand", func() {
		var importDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeBlobStoreVerifier, nil, fakeDropletRunner, nil, fakeZipper, config)
			importDropletCommand = commandFactory.MakeImportDropletCommand()
		})

		Context("when the droplet files exist", func() {
			var tmpDir, dropletPathArg, metadataPathArg string

			BeforeEach(func() {
				var err error
				tmpDir, err = ioutil.TempDir(os.TempDir(), "droplet")
				Expect(err).NotTo(HaveOccurred())

				dropletPathArg = filepath.Join(tmpDir, "droplet.tgz")
				metadataPathArg = filepath.Join(tmpDir, "result.json")
				Expect(ioutil.WriteFile(dropletPathArg, []byte("droplet contents"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(metadataPathArg, []byte("result metadata"), 0644)).To(Succeed())
			})
			AfterEach(func() {
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
			})

			It("imports the droplet", func() {
				test_helpers.ExecuteCommandWithArgs(importDropletCommand, []string{"droplet-name", dropletPathArg, metadataPathArg})

				Expect(outputBuffer).To(test_helpers.SayLine("Imported droplet-name"))

				Expect(fakeDropletRunner.ImportDropletCallCount()).To(Equal(1))
				dropletName, dropletPath, metadataPath := fakeDropletRunner.ImportDropletArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))
				Expect(dropletPath).To(Equal(dropletPathArg))
				Expect(metadataPath).To(Equal(metadataPathArg))
			})

			Context("when the droplet runner returns an error", func() {
				It("prints the error message", func() {
					fakeDropletRunner.ImportDropletReturns(errors.New("dont tread on me"))

					test_helpers.ExecuteCommandWithArgs(importDropletCommand, []string{"droplet-name", dropletPathArg, metadataPathArg})

					Expect(outputBuffer).To(test_helpers.SayLine("Error importing droplet-name: dont tread on me"))
					Expect(fakeDropletRunner.ImportDropletCallCount()).To(Equal(1))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})
		})

		Context("when required arguments are missing", func() {
			It("prints incorrect usage", func() {
				test_helpers.ExecuteCommandWithArgs(importDropletCommand, []string{"droplet-name", "some-path"})

				Expect(outputBuffer).To(test_helpers.SayLine("DROPLET_NAME,DROPLET_PATH and METADATA_PATH are required"))
				Expect(fakeDropletRunner.ImportDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})
})
