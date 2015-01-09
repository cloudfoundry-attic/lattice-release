package setup_cli

import (
	"os"
	"os/signal"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/noaa"
	"github.com/dajulia3/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/config_helpers"
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier/receptor_client_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/logs"
	"github.com/pivotal-cf-experimental/lattice-cli/ltc/setup_cli/setup_cli_helpers"
	"github.com/pivotal-cf-experimental/lattice-cli/output"

	app_examiner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory"
	app_runner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
	config_command_factory "github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
	logs_command_factory "github.com/pivotal-cf-experimental/lattice-cli/logs/command_factory"
)

func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "ltc"
	app.Usage = "Command line interface for Lattice."

	input := os.Stdin
	output := output.New(os.Stdout)

	config := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(userHome())))
	config.Load()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan)

	receptorClient := receptor.NewClient(config.Receptor())
	appRunner := app_runner.New(receptorClient, config.Target())

	timeprovider := timeprovider.NewTimeProvider()
	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunner, docker_metadata_fetcher.New(), output, timeout(), config.Target(), os.Environ(), timeprovider)

	logReader := logs.NewLogReader(noaa.NewConsumer(setup_cli_helpers.LoggregatorUrl(config.Loggregator()), nil, nil))
	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(logReader, output)

	targetVerifier := target_verifier.New(receptor_client_factory.MakeReceptorClient)
	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, targetVerifier, input, output)

	appExaminer := app_examiner.New(receptorClient)
	appExaminerCommandFactory := app_examiner_command_factory.NewAppExaminerCommandFactory(appExaminer, output, timeprovider, signalChan)

	app.Commands = []cli.Command{
		appRunnerCommandFactory.MakeStartAppCommand(),
		appRunnerCommandFactory.MakeScaleAppCommand(),
		appRunnerCommandFactory.MakeStopAppCommand(),
		appRunnerCommandFactory.MakeRemoveAppCommand(),
		logsCommandFactory.MakeLogsCommand(),
		configCommandFactory.MakeTargetCommand(),
		appExaminerCommandFactory.MakeListAppCommand(),
		appExaminerCommandFactory.MakeVisualizeCommand(),
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
