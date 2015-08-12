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
		Name:        "target",
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
	factory.config.SetBlobStore(target, "8444", "", "")

	_, authorized, err := factory.targetVerifier.VerifyTarget(factory.config.Receptor())
	if err != nil {
		factory.ui.SayLine(fmt.Sprint("Error verifying target: ", err))
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}
	if authorized {
		if !factory.verifyBlobStore() {
			factory.exitHandler.Exit(exit_codes.BadTarget)
			return
		}

		factory.save()
		return
	}

	username := factory.ui.Prompt("Username")
	password := factory.ui.PromptForPassword("Password")

	factory.config.SetLogin(username, password)
	factory.config.SetBlobStore(target, "8444", username, password)

	_, authorized, err = factory.targetVerifier.VerifyTarget(factory.config.Receptor())
	if err != nil {
		factory.ui.SayLine(fmt.Sprint("Error verifying target: ", err))
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}
	if !authorized {
		factory.ui.SayLine("Could not authorize target.")
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	if !factory.verifyBlobStore() {
		factory.exitHandler.Exit(exit_codes.BadTarget)
		return
	}

	factory.save()
}

func (factory *ConfigCommandFactory) verifyBlobStore() bool {
	authorized, err := factory.blobStoreVerifier.Verify(factory.config.BlobStore())
	if err != nil {
		factory.ui.SayLine("Could not connect to the droplet store.")
		return false
	}
	if !authorized {
		factory.ui.SayLine("Could not authenticate with the droplet store.")
		return false
	}
	return true
}

func (factory *ConfigCommandFactory) save() {
	err := factory.config.Save()
	if err != nil {
		factory.ui.SayLine(err.Error())
		factory.exitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	factory.ui.SayLine("API location set.")
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
		factory.ui.SayLine("\tNo droplet store specified.")
		return
	}

	endpoint := fmt.Sprintf("%s:%s", blobStore.Host, blobStore.Port)
	if username := blobStore.Username; username != "" {
		endpoint = fmt.Sprintf("%s@%s", username, endpoint)
	}
	factory.ui.SayLine(fmt.Sprintf("Droplet store:\t%s", endpoint))
}
