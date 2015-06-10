package command_factory

import (
	"fmt"
	"os"

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
		Description: "ltc upload-bits BLOB_KEY /path/to/file",
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

	uploadFile, err := os.Open(archivePath)
	if err != nil {
		factory.ui.Say(fmt.Sprintf("Error opening %s: %s", archivePath, err))
		return
	}

	if err := factory.dropletRunner.UploadBits(dropletName, uploadFile); err != nil {
		factory.ui.Say(fmt.Sprintf("Error uploading to %s: %s", dropletName, err))
		return
	}

	factory.ui.Say("Successfully uploaded " + dropletName)
}
