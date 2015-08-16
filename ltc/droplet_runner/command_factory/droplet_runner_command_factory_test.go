package command_factory_test

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore/fake_cf_ignore"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/fake_droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
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
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeTaskExaminer, fakeDropletRunner, fakeCFIgnore)
			buildDropletCommand = commandFactory.MakeBuildDropletCommand()
		})

		Context("when the archive path is a folder and exists", func() {
			var (
				prevDir, tmpDir string
				err             error
			)

			BeforeEach(func() {
				tmpDir, err = ioutil.TempDir(os.TempDir(), "zip_contents")
				Expect(err).NotTo(HaveOccurred())

				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "aaa"), []byte("aaa contents"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "bbb"), []byte("bbb contents"), 0750)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "ccc"), []byte("ccc contents"), 0644)).To(Succeed())
				Expect(os.Symlink("ccc", filepath.Join(tmpDir, "ddd"))).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "some-ignored-file"), []byte("ignored contents"), 0644)).To(Succeed())

				Expect(os.Mkdir(filepath.Join(tmpDir, "subfolder"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "subfolder", "sub"), []byte("sub contents"), 0644)).To(Succeed())

				prevDir, err = os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				Expect(os.Chdir(tmpDir)).To(Succeed())

				fakeCFIgnore.ShouldIgnoreStub = func(path string) bool {
					return path == "some-ignored-file"
				}
			})

			AfterEach(func() {
				Expect(os.Chdir(prevDir)).To(Succeed())
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
			})

			It("zips up current working folder and uploads as the droplet name", func() {
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.Say("Submitted build of droplet-name"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				dropletName, uploadPath := fakeDropletRunner.UploadBitsArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))

				Expect(uploadPath).ToNot(BeNil())
				Expect(uploadPath).To(HaveSuffix(".zip"))

				zipReader, err := zip.OpenReader(uploadPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(zipReader.File).To(HaveLen(6))

				buffer := make([]byte, 12)
				h := zipReader.File[0].FileHeader
				f, err := zipReader.File[0].Open()
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(h.Name).To(Equal("aaa"))
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0700)))
				Expect(f.Read(buffer)).To(Equal(12))
				Expect(string(buffer)).To(Equal("aaa contents"))

				buffer = make([]byte, 12)
				h = zipReader.File[1].FileHeader
				f, err = zipReader.File[1].Open()
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(h.Name).To(Equal("bbb"))
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0750)))
				Expect(f.Read(buffer)).To(Equal(12))
				Expect(string(buffer)).To(Equal("bbb contents"))

				buffer = make([]byte, 12)
				h = zipReader.File[2].FileHeader
				f, err = zipReader.File[2].Open()
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(h.Name).To(Equal("ccc"))
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
				Expect(f.Read(buffer)).To(Equal(12))
				Expect(string(buffer)).To(Equal("ccc contents"))

				buffer = make([]byte, 3)
				h = zipReader.File[3].FileHeader
				f, err = zipReader.File[3].Open()
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(h.Name).To(Equal("ddd"))
				Expect(h.FileInfo().Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
				Expect(f.Read(buffer)).To(Equal(3))
				Expect(string(buffer)).To(Equal("ccc"))

				buffer = make([]byte, 1)
				h = zipReader.File[4].FileHeader
				f, err = zipReader.File[4].Open()
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(h.Name).To(Equal("subfolder/"))
				Expect(h.FileInfo().IsDir()).To(BeTrue())
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(os.ModeDir | 0755)))
				_, err = f.Read(buffer)
				Expect(err).To(MatchError("EOF"))

				buffer = make([]byte, 12)
				h = zipReader.File[5].FileHeader
				f, err = zipReader.File[5].Open()
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(h.Name).To(Equal("subfolder/sub"))
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
				Expect(f.Read(buffer)).To(Equal(12))
				Expect(string(buffer)).To(Equal("sub contents"))
			})

			Context("when a single file is passed to -p", func() {
				It("uses it as the app bits if it's already a zip archive", func() {
					tmpFile, err := ioutil.TempFile(os.TempDir(), "singlezip")
					Expect(err).NotTo(HaveOccurred())

					zipFilePath := tmpFile.Name() + ".zip"

					zipCommand := exec.Command("/usr/bin/zip", "--symlinks", "-yr", zipFilePath, "aaa", "bbb", "ccc", "ddd", "some-ignored-file", "subfolder")
					zipCommand.Dir = tmpDir
					Expect(zipCommand.Run()).To(Succeed())
					defer os.Remove(zipFilePath)
					defer tmpFile.Close()
					defer os.Remove(tmpFile.Name())

					args := []string{
						"droplet-name",
						"http://some.url/for/buildpack",
						"-p",
						zipFilePath,
					}

					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, args)

					Expect(outputBuffer).To(test_helpers.Say("Submitted build of droplet-name"))
					Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
					dropletName, uploadPath := fakeDropletRunner.UploadBitsArgsForCall(0)
					Expect(dropletName).To(Equal("droplet-name"))

					Expect(uploadPath).ToNot(BeNil())
					Expect(uploadPath).To(HaveSuffix(".zip"))

					zipReader, err := zip.OpenReader(uploadPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(zipReader.File).To(HaveLen(7))

					buffer := make([]byte, 12)
					h := zipReader.File[0].FileHeader
					f, err := zipReader.File[0].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("aaa"))
					Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0700)))
					Expect(f.Read(buffer)).To(Equal(12))
					Expect(string(buffer)).To(Equal("aaa contents"))

					buffer = make([]byte, 12)
					h = zipReader.File[1].FileHeader
					f, err = zipReader.File[1].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("bbb"))
					Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0750)))
					Expect(f.Read(buffer)).To(Equal(12))
					Expect(string(buffer)).To(Equal("bbb contents"))

					buffer = make([]byte, 12)
					h = zipReader.File[2].FileHeader
					f, err = zipReader.File[2].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("ccc"))
					Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
					Expect(f.Read(buffer)).To(Equal(12))
					Expect(string(buffer)).To(Equal("ccc contents"))

					buffer = make([]byte, 3)
					h = zipReader.File[3].FileHeader
					f, err = zipReader.File[3].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("ddd"))
					Expect(h.FileInfo().Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
					Expect(f.Read(buffer)).To(Equal(3))
					Expect(string(buffer)).To(Equal("ccc"))

					buffer = make([]byte, 12)
					h = zipReader.File[4].FileHeader
					f, err = zipReader.File[4].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("some-ignored-file"))
					Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
					Expect(f.Read(buffer)).To(Equal(12))
					Expect(string(buffer)).To(Equal("ignored cont"))

					buffer = make([]byte, 1)
					h = zipReader.File[5].FileHeader
					f, err = zipReader.File[5].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("subfolder/"))
					Expect(h.FileInfo().IsDir()).To(BeTrue())
					Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(os.ModeDir | 0755)))
					_, err = f.Read(buffer)
					Expect(err).To(MatchError("EOF"))

					buffer = make([]byte, 12)
					h = zipReader.File[6].FileHeader
					f, err = zipReader.File[6].Open()
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()
					Expect(h.Name).To(Equal("subfolder/sub"))
					Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
					Expect(f.Read(buffer)).To(Equal(12))
					Expect(string(buffer)).To(Equal("sub contents"))
				})

				It("prints an error if the file isn't a zip archive", func() {
					args := []string{
						"droplet-name",
						"http://some.url/for/buildpack",
						"-p",
						path.Join(tmpDir, "ccc"),
					}

					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, args)

					Expect(outputBuffer).To(gbytes.Say("Error archiving .*: ccc must be a zip archive"))
					Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
					Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
				})
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

				_, _, _, env, _, _, _ := fakeDropletRunner.BuildDropletArgsForCall(0)
				Expect(env["AAAA"]).To(Equal("xyz"))
				Expect(env["BBBB"]).To(Equal("2"))
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

			Describe(".cfignore", func() {
				Context("when a .cfignore file is present", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(tmpDir, ".cfignore"), []byte("cfignore contents"), 0644)).To(Succeed())
					})

					It("parses a .cfignore file if present", func() {
						test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

						Expect(fakeCFIgnore.ParseCallCount()).To(Equal(1))
						Expect(ioutil.ReadAll(fakeCFIgnore.ParseArgsForCall(0))).To(Equal([]byte("cfignore contents")))
					})

					Context("when parsing the .cfignore file fails", func() {
						It("returns an error without uploading any bits", func() {
							fakeCFIgnore.ParseReturns(errors.New("some cfignore parse error"))

							test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

							Expect(outputBuffer).To(test_helpers.Say("some cfignore parse error"))
							Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
							Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
						})
					})
				})

				It("does not parse a .cfignore file when missing", func() {
					test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

					Expect(fakeCFIgnore.ParseCallCount()).To(Equal(0))
				})
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

					Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: invalid buildpack ¥¥¥¥://¥¥¥¥¥¥¥¥"))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
					Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
					Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
				})
			})
		})

		Context("when the droplet runner returns an error", func() {
			It("prints the error from upload bits", func() {
				fakeDropletRunner.UploadBitsReturns(errors.New("uploading bits failed"))

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.Say("Error uploading to droplet-name: uploading bits failed"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
			})

			It("prints the error from build droplet", func() {
				fakeDropletRunner.BuildDropletReturns(errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name", "http://some.url/for/buildpack"})

				Expect(outputBuffer).To(test_helpers.Say("Error submitting build of droplet-name: failed"))
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

				Eventually(outputBuffer).Should(test_helpers.Say("Submitted build of droplet-name"))

				Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("build-droplet-droplet-name"))

				Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
				Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("build-droplet-droplet-name"))

				fakeClock.IncrementBySeconds(1)
				Expect(doneChan).ShouldNot(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "RUNNING"}, nil)

				fakeClock.IncrementBySeconds(1)
				Expect(doneChan).ShouldNot(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "COMPLETED"}, nil)

				fakeClock.IncrementBySeconds(1)
				Eventually(doneChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.SayLine("Build completed"))
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
			})

			Context("when the build doesn't complete before the timeout elapses", func() {
				It("alerts the user the build took too long", func() {
					args := []string{
						"droppo-the-clown",
						"http://some.url/for/buildpack",
						"-t",
						"17s",
					}

					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "RUNNING"}, nil)

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(buildDropletCommand, args)

					Eventually(outputBuffer).Should(test_helpers.Say("Submitted build of droppo-the-clown"))
					Expect(outputBuffer).To(test_helpers.SayNewLine())

					fakeClock.IncrementBySeconds(17)

					Eventually(doneChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.Say(colors.Red("Timed out waiting for the build to complete.")))
					Expect(outputBuffer).To(test_helpers.SayNewLine())
					Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still building your application in the background."))
					Expect(outputBuffer).To(test_helpers.SayLine("To view logs:\n\tltc logs build-droplet-droppo-the-clown"))
					Expect(outputBuffer).To(test_helpers.SayLine("To view status:\n\tltc status build-droplet-droppo-the-clown"))
				})
			})

			Context("when the build completes", func() {
				It("alerts the user of a complete but failed build", func() {
					args := []string{"droppo-the-clown", "http://some.url/for/buildpack"}

					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{State: "PENDING"}, nil)

					test_helpers.AsyncExecuteCommandWithArgs(buildDropletCommand, args)

					fakeTaskExaminer.TaskStatusReturns(task_examiner.TaskInfo{
						State:         "COMPLETED",
						Failed:        true,
						FailureReason: "oops",
					}, nil)

					fakeClock.IncrementBySeconds(1)

					Eventually(outputBuffer).Should(test_helpers.SayLine("Build failed: oops"))
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

					Eventually(outputBuffer).Should(test_helpers.Say("Submitted build of droppo-the-clown"))

					Expect(fakeTaskExaminer.TaskStatusCallCount()).To(Equal(1))
					Expect(fakeTaskExaminer.TaskStatusArgsForCall(0)).To(Equal("build-droplet-droppo-the-clown"))

					fakeClock.IncrementBySeconds(1)
					Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

					fakeClock.IncrementBySeconds(1)
					Eventually(doneChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.SayNewLine())
					Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error requesting task status: dropped the ball")))
					Expect(outputBuffer).ToNot(test_helpers.Say("Timed out waiting for the build to complete."))
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
				})
			})
		})

		Context("invalid syntax", func() {
			It("rejects less than two positional arguments", func() {
				test_helpers.ExecuteCommandWithArgs(buildDropletCommand, []string{"droplet-name"})

				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
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

				Expect(outputBuffer).To(test_helpers.Say("Invalid CPU Weight"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
				Expect(fakeDropletRunner.BuildDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})

	Describe("ListDropletsCommand", func() {
		var listDropletsCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeTaskExaminer, fakeDropletRunner, nil)
			listDropletsCommand = commandFactory.MakeListDropletsCommand()
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

				Expect(outputBuffer).To(test_helpers.Say("Error listing droplets: failed"))
				Expect(fakeDropletRunner.ListDropletsCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})

	Describe("LaunchDropletCommand", func() {
		var launchDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeTaskExaminer, fakeDropletRunner, nil)
			launchDropletCommand = commandFactory.MakeLaunchDropletCommand()
		})

		Context("when a malformed tcp routes flag is passed", func() {
			It("errors out when the container port is not an int", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--tcp-routes=woo:50000",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidRoutePortErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("errors out when the tcp route is incomplete", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--tcp-routes=5222,50000",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MalformedTcpRouteErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

		})
		It("launches the specified droplet with tcp routes", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
			args := []string{
				"--routes=4444:ninetyninety,9090:fourtyfourfourtyfour",
				"--tcp-routes=5222:50000,5223:50001",
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
				app_runner.RouteOverride{HostnamePrefix: "ninetyninety", Port: 4444},
				app_runner.RouteOverride{HostnamePrefix: "fourtyfourfourtyfour", Port: 9090},
			}))
			tcpRoutes := appEnvParam.TcpRoutes

			Expect(tcpRoutes).ShouldNot(BeNil())
			Expect(tcpRoutes).Should(ContainExactly(
				app_runner.TcpRoutes{
					app_runner.TcpRoute{
						ExternalPort: 50000,
						Port:         5222,
					},
					app_runner.TcpRoute{
						ExternalPort: 50001,
						Port:         5223,
					},
				},
			))
		})

		It("launches the specified droplet", func() {
			fakeAppExaminer.RunningAppInstancesInfoReturns(11, false, nil)
			args := []string{
				"--cpu-weight=57",
				"--memory-mb=12",
				"--disk-mb=12",
				"--routes=4444:ninetyninety,9090:fourtyfourfourtyfour",
				"--working-dir=/xxx",
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
			Expect(appEnvParam.WorkingDir).To(Equal("/xxx"))
			Expect(appEnvParam.CPUWeight).To(Equal(uint(57)))
			Expect(appEnvParam.MemoryMB).To(Equal(12))
			Expect(appEnvParam.DiskMB).To(Equal(12))
			Expect(appEnvParam.Privileged).To(BeTrue())
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
			Expect(appEnvParam.WorkingDir).To(Equal("/home/vcap/app"))
			Expect(appEnvParam.Privileged).To(BeTrue())
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

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: APP_NAME and DROPLET_NAME are required"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
			})

			It("validates that the terminator -- is passed in when a start command is specified", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"not-the-terminator",
					"start-me-up",
				}
				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: '--' Required before start command"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
			})

			It("validates the CPU weight is in 1-100", func() {
				args := []string{
					"cool-app",
					"cool-droplet",
					"--cpu-weight=0",
				}

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Invalid CPU Weight"))
				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
			})
		})

		Context("when a malformed routes flag is passed", func() {
			It("errors out when the port is not an int", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--routes=woo:aahh",
				}

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MalformedRouteErrorMessage))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("errors out when there is no colon", func() {
				args := []string{
					"cool-web-app",
					"cool-web-droplet",
					"--routes=8888",
				}

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, args)

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MalformedRouteErrorMessage))
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
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
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
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
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
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
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
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MonitorPortNotExposed))
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
				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MonitorPortNotExposed))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Context("when the droplet runner returns errors", func() {
			It("prints an error", func() {
				fakeDropletRunner.LaunchDropletReturns(errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(launchDropletCommand, []string{"droppy", "droplet-name"})

				Expect(fakeDropletRunner.LaunchDropletCallCount()).To(Equal(1))

				Expect(outputBuffer).To(test_helpers.Say("Error launching app droppy from droplet droplet-name: failed"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})

	Describe("RemoveDropletCommand", func() {
		var removeDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeTaskExaminer, fakeDropletRunner, nil)
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

				Expect(outputBuffer).To(test_helpers.Say("Error removing droplet droppo: failed"))
				Expect(fakeDropletRunner.RemoveDropletCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})

	Describe("ExportDropletCommand", func() {
		var (
			exportDropletCommand           cli.Command
			exportDir, workingDir, prevDir string
			err                            error
		)

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, fakeTaskExaminer, fakeDropletRunner, nil)
			exportDropletCommand = commandFactory.MakeExportDropletCommand()

		})

		BeforeEach(func() {
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

			Expect(os.Stat("droppo.tgz")).ToNot(BeNil())
			Expect(os.Stat("droppo-metadata.json")).ToNot(BeNil())
		})

		Context("when the droplet runner returns errors", func() {
			It("prints an error", func() {
				fakeDropletRunner.ExportDropletReturns(nil, nil, errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(exportDropletCommand, []string{"droppo"})

				Expect(outputBuffer).To(test_helpers.Say("Error exporting droplet droppo: failed"))
				Expect(fakeDropletRunner.ExportDropletCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when the required arguments are missing", func() {
			It("prints incorrect usage", func() {
				test_helpers.ExecuteCommandWithArgs(exportDropletCommand, []string{""})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
				Expect(fakeDropletRunner.ExportDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})

	Describe("ImportDropletCommand", func() {
		var importDropletCommand cli.Command

		BeforeEach(func() {
			commandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(appRunnerCommandFactory, nil, fakeDropletRunner, nil)
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

				Expect(outputBuffer).To(test_helpers.Say("Imported droplet-name"))

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

					Expect(outputBuffer).To(test_helpers.Say("Error importing droplet-name: dont tread on me"))
					Expect(fakeDropletRunner.ImportDropletCallCount()).To(Equal(1))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})
		})

		Context("when required arguments are missing", func() {
			It("prints incorrect usage", func() {
				test_helpers.ExecuteCommandWithArgs(importDropletCommand, []string{"droplet-name", "some-path"})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
				Expect(fakeDropletRunner.ImportDropletCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})
	})
})
