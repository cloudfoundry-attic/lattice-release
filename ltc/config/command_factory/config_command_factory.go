package command_factory

import (
	"fmt"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

const (
	TargetCommandName = "target"
	blobTargetPort    = "8444"
)

type ConfigCommandFactory struct {
	config            *config.Config
	ui                terminal.UI
	targetVerifier    target_verifier.TargetVerifier
	blobStoreVerifier BlobStoreVerifier
	exitHandler       exit_handler.ExitHandler
}

//go:generate counterfeiter -o fake_blob_store_verifier/fake_blob_store_verifier.go . BlobStoreVerifier
type BlobStoreVerifier interface {
	Verify(config dav_blob_store.Config) (authorized bool, err error)
}

func NewConfigCommandFactory(config *config.Config, ui terminal.UI, targetVerifier target_verifier.TargetVerifier, blobStoreVerifier BlobStoreVerifier, exitHandler exit_handler.ExitHandler) *ConfigCommandFactory {
	return &ConfigCommandFactory{config, ui, targetVerifier, blobStoreVerifier, exitHandler}
}

func (factory *ConfigCommandFactory) MakeTargetCommand() cli.Command {
	var targetCommand = cli.Command{
		Name:        TargetCommandName,
		Aliases:     []string{"ta"},
		Usage:       "Targets a lattice cluster",
		Description: "ltc target TARGET (e.g., 192.168.11.11.xip.io)",
		Action:      factory.target,
	}

	return targetCommand
}

func (factory *ConfigCommandFactory) target(context *cli.Context) {
	target := context.Args().First()

	if target == "" {
		factory.printTarget()
		factory.printBlobTarget()
		return
	}

	factory.config.SetTarget(target)
	factory.config.SetLogin("", "")

	_, authorized, err := factory.targetVerifier.VerifyTarget(factory.config.Receptor())
	if err != nil {
		factory.ui.Say("Error verifying target: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}
	if authorized {
		if !factory.verifyBlobStore(target, blobTargetPort, "", "") {
			return
		}

		factory.ui.SayLine("Blob store is targeted.")
		factory.save()
		return
	}

	username := factory.ui.Prompt("Username")
	password := factory.ui.PromptForPassword("Password")

	factory.config.SetLogin(username, password)
	_, authorized, err = factory.targetVerifier.VerifyTarget(factory.config.Receptor())
	if err != nil {
		factory.ui.Say("Error verifying target: " + err.Error())
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}
	if !authorized {
		factory.ui.Say("Could not authorize target.")
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	if !factory.verifyBlobStore(target, blobTargetPort, username, password) {
		return
	}

	factory.ui.SayLine("Blob store is targeted.")
	factory.save()
}

func (factory *ConfigCommandFactory) verifyBlobStore(host, port, username, password string) bool {
	factory.config.SetBlobStore(host, port, username, password)
	authorized, err := factory.blobStoreVerifier.Verify(factory.config.BlobStore())
	if err != nil {
		factory.config.SetBlobStore("", "", "", "")
		factory.ui.SayLine("Warning: Blob store not running, buildpack support disabled.")
		factory.save()
		return false
	}
	if !authorized {
		factory.ui.SayLine("Blob store requires authorization.")
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return false
	}
	return true
}

func (factory *ConfigCommandFactory) save() {
	err := factory.config.Save()
	if err != nil {
		factory.ui.Say(err.Error())
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	factory.ui.Say("Api Location Set")
}

func (factory *ConfigCommandFactory) printTarget() {
	if factory.config.Target() == "" {
		factory.ui.SayLine("Target not set.")
		return
	}
	target := factory.config.Target()
	if username := factory.config.Username(); username != "" {
		target = fmt.Sprintf("%s@%s", username, target)
	}
	factory.ui.SayLine(fmt.Sprintf("Target:\t\t%s", target))
}

func (factory *ConfigCommandFactory) printBlobTarget() {
	blobStore := factory.config.BlobStore()
	if blobStore.Host == "" {
		factory.ui.SayLine("\tNo blob store specified.")
		return
	}

	endpoint := fmt.Sprintf("%s:%s", blobStore.Host, blobStore.Port)
	if username := blobStore.Username; username != "" {
		endpoint = fmt.Sprintf("%s@%s", username, endpoint)
	}
	factory.ui.SayLine(fmt.Sprintf("Blob Store:\t%s", endpoint))
}
