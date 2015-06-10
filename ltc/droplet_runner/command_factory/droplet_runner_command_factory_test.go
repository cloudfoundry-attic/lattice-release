package command_factory_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/fake_droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("CommandFactory", func() {
	var (
		outputBuffer      *gbytes.Buffer
		terminalUI        terminal.UI
		fakeDropletRunner *fake_droplet_runner.FakeDropletRunner
		fakeExitHandler   *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		fakeDropletRunner = &fake_droplet_runner.FakeDropletRunner{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}

	})

	Describe("UploadBitsCommand", func() {
		var (
			uploadBitsCommand cli.Command
		)

		BeforeEach(func() {
			commandFactory := command_factory.NewDropletRunnerCommandFactory(fakeDropletRunner, terminalUI, fakeExitHandler)
			uploadBitsCommand = commandFactory.MakeUploadBitsCommand()
		})

		Context("when the archive file exists", func() {
			var (
				tmpFile *os.File
				err     error
			)
			BeforeEach(func() {
				tmpDir := os.TempDir()
				tmpFile, err = ioutil.TempFile(tmpDir, "tmp_file")
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				os.RemoveAll(tmpFile.Name())
			})

			It("checks the file exists and calls the droplet runner", func() {
				test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"droplet-name", tmpFile.Name()})

				Expect(outputBuffer).To(test_helpers.Say("Successfully uploaded droplet-name"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				dropletName, fileReader := fakeDropletRunner.UploadBitsArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))
				Expect(fileReader).ToNot(BeNil())
			})

			Context("when the droplet runner returns an error", func() {
				It("prints the error", func() {
					fakeDropletRunner.UploadBitsReturns(errors.New("uploading bits failed"))

					test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"droplet-name", tmpFile.Name()})

					Expect(outputBuffer).To(test_helpers.Say("Error uploading to droplet-name: uploading bits failed"))
					Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				})
			})
		})

		It("errors when opening an archive file", func() {
			nonExistentFile := filepath.Join(os.TempDir(), "non_existant_file")

			test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"droplet-name", nonExistentFile})

			Expect(outputBuffer).To(test_helpers.Say("Error opening " + nonExistentFile))
			Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
		})

		Context("invalid syntax", func() {
			It("rejects less than two positional arguments", func() {
				test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"droplet-name"})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
			})

			It("tests for an empty droplet name", func() {
				test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"", "my-file-name"})

				Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(0))
			})
		})

	})
})
