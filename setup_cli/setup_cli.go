package setup_cli

import (
	"os"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/noaa"
	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/app_runner"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config/config_helpers"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config/persister"
	"github.com/pivotal-cf-experimental/diego-edge-cli/logs"
	"github.com/pivotal-cf-experimental/diego-edge-cli/logs/logs_helpers"

	app_runner_command_factory "github.com/pivotal-cf-experimental/diego-edge-cli/app_runner/command_factory"
	config_command_factory "github.com/pivotal-cf-experimental/diego-edge-cli/config/command_factory"
	logs_command_factory "github.com/pivotal-cf-experimental/diego-edge-cli/logs/command_factory"
)

func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Diego"
	app.Usage = "Command line interface for diego."

	config := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(userHome())))
	config.Load()

	receptorClient := receptor.NewClient(config.Api())
	appRunner := app_runner.NewDiegoAppRunner(receptorClient)

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunner, os.Stdout)

	logReader := logs.NewLogReader(noaa.NewConsumer(logs_helpers.LoggregatorUrl(config.Loggregator()), nil, nil))
	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(logReader, os.Stdout)

	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, os.Stdout)

	app.Commands = []cli.Command{
		appRunnerCommandFactory.MakeStartDiegoAppCommand(),
		appRunnerCommandFactory.MakeScaleDiegoAppCommand(),
		appRunnerCommandFactory.MakeStopDiegoAppCommand(),
		logsCommandFactory.MakeLogsCommand(),
		configCommandFactory.MakeSetTargetCommand(),
		configCommandFactory.MakeSetTargetLoggregatorCommand(),
	}
	return app
}

func userHome() string {
	if os.Getenv("DIEGO_CLI_HOME") != "" {
		return os.Getenv("DIEGO_CLI_HOME")
	}

	return os.Getenv("HOME")
}
