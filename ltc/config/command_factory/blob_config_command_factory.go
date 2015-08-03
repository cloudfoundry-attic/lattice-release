package command_factory

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/codegangsta/cli"
)

const TargetBlobCommandName = "target-blob"

func (factory *ConfigCommandFactory) MakeTargetBlobCommand() cli.Command {
	var targetBlobCommand = cli.Command{
		Name:        TargetBlobCommandName,
		Aliases:     []string{"tb"},
		Usage:       "Targets a lattice blob store",
		Description: "ltc target-blob TARGET:PORT (e.g., 192.168.11.11:8444)",
		Action:      factory.targetBlob,
	}

	return targetBlobCommand
}

func (factory *ConfigCommandFactory) targetBlob(context *cli.Context) {
	endpoint := context.Args().First()

	if endpoint == "" {
		blobTarget := factory.config.BlobTarget()
		if blobTarget.Host == "" {
			factory.ui.SayLine("Blob store not set")
			return
		}
		factory.ui.Say(fmt.Sprintf("Blob Store:\t%s:%d\n", blobTarget.Host, blobTarget.Port))
		factory.ui.Say(fmt.Sprintf("Username:\t%s\n", blobTarget.Username))
		factory.ui.Say(fmt.Sprintf("Password:\t%s\n", blobTarget.Password))
		return
	}

	blobHost, blobPort, err := net.SplitHostPort(endpoint)
	if err != nil {
		factory.ui.SayLine("Error setting blob target: malformed target")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	port, err := strconv.Atoi(blobPort)
	if err != nil || port > 65536 {
		factory.ui.SayLine("Error setting blob target: malformed port")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	username := factory.ui.Prompt("Username")
	password := factory.ui.Prompt("Password")

	if err := factory.targetVerifier.VerifyBlobTarget(dav_blob_store.Config{
		Host:     blobHost,
		Port:     uint16(port),
		Username: username,
		Password: password,
	}); err != nil {
		factory.ui.Say("Unable to verify blob store: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	factory.config.SetBlobTarget(blobHost, uint16(port), username, password)
	if err := factory.config.Save(); err != nil {
		factory.ui.SayLine(err.Error())
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	factory.ui.Say("Blob Location Set")
}
