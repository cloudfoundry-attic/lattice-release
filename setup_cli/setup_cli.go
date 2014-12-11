package setup_cli

import (
	"os"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/noaa"
	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/config_helpers"
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/logs"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/setup_cli/setup_cli_helpers"

	app_runner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
	config_command_factory "github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
	logs_command_factory "github.com/pivotal-cf-experimental/lattice-cli/logs/command_factory"
)

func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "ltc"
	app.Usage = "Command line interface for Lattice."

	output := output.New(os.Stdout)

	config := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(userHome())))
	config.Load()

	receptorClient := receptor.NewClient(config.Receptor())
	appRunner := app_runner.NewAppRunner(receptorClient, config.Target())

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunner, output, timeout(), config.Target(), os.Environ())

	logReader := logs.NewLogReader(noaa.NewConsumer(setup_cli_helpers.LoggregatorUrl(config.Loggregator()), nil, nil))
	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(logReader, output)

	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, output)

	app.Commands = []cli.Command{
		appRunnerCommandFactory.MakeStartAppCommand(),
		appRunnerCommandFactory.MakeScaleAppCommand(),
		appRunnerCommandFactory.MakeStopAppCommand(),
		logsCommandFactory.MakeLogsCommand(),
		configCommandFactory.MakeSetTargetCommand(),
	}
	return app
}

func userHome() string {
	if os.Getenv("LATTICE_CLI_HOME") != "" {
		return os.Getenv("LATTICE_CLI_HOME")
	}

	return os.Getenv("HOME")
}

func timeout() time.Duration {
	return setup_cli_helpers.Timeout(os.Getenv("LATTICE_CLI_TIMEOUT"))
}
