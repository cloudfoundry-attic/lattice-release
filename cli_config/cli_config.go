package cli_config

import (
	"os"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/app_runner"

	start_app_command_factory "github.com/pivotal-cf-experimental/diego-edge-cli/app_runner/command_factory"
)

func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Diego"
	app.Usage = "Command line interface for diego."

	receptorAddress := os.Getenv("DIEGO_RECEPTOR_ADDRESS")
	receptorUsername := os.Getenv("DIEGO_RECEPTOR_USERNAME")
	receptorPassword := os.Getenv("DIEGO_RECEPTOR_PASSWORD")

	if receptorAddress == "" {
		panic("DIEGO_RECEPTOR_ADDRESS environment variable was not set")
	}

	receptorClient := receptor.NewClient(receptorAddress, receptorUsername, receptorPassword)

	appRunner := app_runner.NewDiegoAppRunner(receptorClient)

	app.Commands = []cli.Command{
		start_app_command_factory.NewStartAppCommandFactory(appRunner, os.Stdout).MakeCommand(),
	}
	return app
}
