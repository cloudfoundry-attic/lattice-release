package command_factory

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
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
		if blobTarget.TargetHost == "" {
			factory.ui.SayLine("Blob target not set")
			return
		}
		factory.ui.Say(fmt.Sprintf("Blob Target:\t%s:%d\n", blobTarget.TargetHost, blobTarget.TargetPort))
		factory.ui.Say(fmt.Sprintf("Access Key:\t%s\n", blobTarget.AccessKey))
		factory.ui.Say(fmt.Sprintf("Secret Key:\t%s\n", blobTarget.SecretKey))
		factory.ui.Say(fmt.Sprintf("Bucket Name:\t%s\n", blobTarget.BucketName))
		return
	}

	var port int
	endpointArr := strings.Split(endpoint, ":")
	if len(endpointArr) != 2 {
		factory.ui.SayLine("Error setting blob target: malformed target")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}
	host := endpointArr[0]

	port, err := strconv.Atoi(endpointArr[1])
	if err != nil || port > 65536 {
		factory.ui.SayLine("Error setting blob target: malformed port")
		factory.exitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	accessKey := factory.ui.Prompt("Access Key")
	secretKey := factory.ui.Prompt("Secret Key")
	bucketName := factory.ui.PromptWithDefault("Bucket Name", "condenser-bucket")

	if err := factory.targetVerifier.VerifyBlobTarget(config.BlobTargetInfo{
		TargetHost: host,
		TargetPort: uint16(port),
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: bucketName,
	}); err != nil {
		factory.ui.Say("Unable to verify blob store: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	factory.config.SetBlobTarget(host, uint16(port), accessKey, secretKey, bucketName)
	if err := factory.config.Save(); err != nil {
		factory.ui.SayLine(err.Error())
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	factory.ui.Say("Blob Location Set")
}
