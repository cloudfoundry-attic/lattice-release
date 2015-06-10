package command_factory

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

type DropletRunnerCommandFactory struct {
	dropletRunner droplet_runner.DropletRunner
	ui            terminal.UI
	exitHandler   exit_handler.ExitHandler
}

func NewDropletRunnerCommandFactory(dropletRunner droplet_runner.DropletRunner, ui terminal.UI, exitHandler exit_handler.ExitHandler) *DropletRunnerCommandFactory {
	return &DropletRunnerCommandFactory{
		dropletRunner: dropletRunner,
		ui:            ui,
		exitHandler:   exitHandler,
	}
}

func (factory *DropletRunnerCommandFactory) MakeUploadBitsCommand() cli.Command {
	var uploadBitsCommand = cli.Command{
		Name:        "upload-bits",
		Aliases:     []string{"ub"},
		Usage:       "Upload bits to the blob store",
		Description: "ltc upload-bits BLOB_KEY /path/to/file-or-folder",
		Action:      factory.uploadBits,
	}

	return uploadBitsCommand
}

func (factory *DropletRunnerCommandFactory) uploadBits(context *cli.Context) {
	dropletName := context.Args().First()
	archivePath := context.Args().Get(1)

	if dropletName == "" || archivePath == "" {
		factory.ui.SayIncorrectUsage("")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	fileInfo, err := os.Stat(archivePath)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error opening %s: %s", archivePath, err))
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	if fileInfo.IsDir() {
		if archivePath, err = makeTar(archivePath); err != nil {
			factory.ui.Say(fmt.Sprintf("Error archiving %s: %s", context.Args().Get(1), err))
			factory.exitHandler.Exit(exit_codes.FileSystemError)
			return
		}
	}

	if err := factory.dropletRunner.UploadBits(dropletName, archivePath); err != nil {
		factory.ui.Say(fmt.Sprintf("Error uploading to %s: %s", dropletName, err))
		return
	}

	factory.ui.Say("Successfully uploaded " + dropletName)
}

func makeTar(path string) (string, error) {
	tmpPath, err := ioutil.TempDir(os.TempDir(), "droplet")
	if err != nil {
		return "", err
	}

	fileWriter, err := os.OpenFile(filepath.Join(tmpPath, "droplet.tar"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	tarWriter := tar.NewWriter(fileWriter)
	defer tarWriter.Close()

	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var relpath string
		if relpath, err = filepath.Rel(path, subpath); err != nil {
			return err
		}

		if relpath == fileWriter.Name() || relpath == "." || relpath == ".." {
			return nil
		}

		if h, _ := tar.FileInfoHeader(info, subpath); h != nil {
			h.Name = relpath
			if err := tarWriter.WriteHeader(h); err != nil {
				return err
			}
		}

		if !info.IsDir() {
			fr, err := os.Open(subpath)
			if err != nil {
				return err
			}
			defer fr.Close()
			if _, err := io.Copy(tarWriter, fr); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return fileWriter.Name(), nil
}
