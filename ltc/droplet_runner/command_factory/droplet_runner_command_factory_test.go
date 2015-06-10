package command_factory_test

import (
	"archive/tar"
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

		Context("when the archive path is a file and exists", func() {
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
				dropletName, uploadPath := fakeDropletRunner.UploadBitsArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))
				Expect(uploadPath).To(Equal(tmpFile.Name()))
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

		Context("when the archive path is a folder and exists", func() {
			var (
				tmpDir string
				err    error
			)

			BeforeEach(func() {
				tmpDir, err = ioutil.TempDir(os.TempDir(), "tar_contents")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(tmpDir, "aaa"), []byte("AAAAAAAAA"), 0700)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(filepath.Join(tmpDir, "bbb"), []byte("BBBBBBB"), 0750)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(filepath.Join(tmpDir, "ccc"), []byte("CCCCCC"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = os.Mkdir(filepath.Join(tmpDir, "subfolder"), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(filepath.Join(tmpDir, "subfolder", "sub"), []byte("SUBSUB"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
			})

			It("tars up the folder and uploads as the droplet name", func() {
				test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"droplet-name", tmpDir})

				Expect(outputBuffer).To(test_helpers.Say("Successfully uploaded droplet-name"))
				Expect(fakeDropletRunner.UploadBitsCallCount()).To(Equal(1))
				dropletName, uploadPath := fakeDropletRunner.UploadBitsArgsForCall(0)
				Expect(dropletName).To(Equal("droplet-name"))

				Expect(uploadPath).ToNot(BeNil())
				Expect(uploadPath).To(HaveSuffix(".tar"))

				file, _ := os.Open(uploadPath)
				tarReader := tar.NewReader(file)

				var h *tar.Header
				h, err = tarReader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(h.FileInfo().Name()).To(Equal("aaa"))
				Expect(h.FileInfo().IsDir()).To(BeFalse())
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0700)))

				h, err = tarReader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(h.FileInfo().Name()).To(Equal("bbb"))
				Expect(h.FileInfo().IsDir()).To(BeFalse())
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0750)))

				h, err = tarReader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(h.FileInfo().Name()).To(Equal("ccc"))
				Expect(h.FileInfo().IsDir()).To(BeFalse())
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))

				h, err = tarReader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(h.FileInfo().Name()).To(Equal("subfolder"))
				Expect(h.FileInfo().IsDir()).To(BeTrue())
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(os.ModeDir | 0755)))

				h, err = tarReader.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(h.FileInfo().Name()).To(Equal("sub"))
				Expect(h.FileInfo().IsDir()).To(BeFalse())
				Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))

				_, err = tarReader.Next()
				Expect(err).To(HaveOccurred())
			})
		})

		It("errors when opening a non-existent archive file", func() {
			nonExistentFile := filepath.Join(os.TempDir(), "non_existent_file")

			test_helpers.ExecuteCommandWithArgs(uploadBitsCommand, []string{"droplet-name", nonExistentFile})

			Expect(outputBuffer).To(test_helpers.Say("Error opening " + nonExistentFile))
			Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.FileSystemError}))
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
