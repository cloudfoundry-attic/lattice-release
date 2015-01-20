package setup_cli

import (
	"os"
	"os/signal"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/noaa"
	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/config_helpers"
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier/receptor_client_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler"
	"github.com/pivotal-cf-experimental/lattice-cli/logs"
	"github.com/pivotal-cf-experimental/lattice-cli/ltc/setup_cli/setup_cli_helpers"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-golang/lager"

	app_examiner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory"
	app_runner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
	config_command_factory "github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
	logs_command_factory "github.com/pivotal-cf-experimental/lattice-cli/logs/command_factory"
)

func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "ltc"
	app.Usage = "Command line interface for Lattice."

	config := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(userHome())))
	config.Load()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	exitHandler := exit_handler.New(signalChan, os.Exit)
	go exitHandler.Run()

	app.Commands = cliCommands(exitHandler, config, logger())

	return app
}

func logger() lager.Logger {
	logger := lager.NewLogger("ltc")
	var logLevel lager.LogLevel

	if os.Getenv("LTC_LOG_LEVEL") == "DEBUG" {
		logLevel = lager.DEBUG
	} else {
		logLevel = lager.INFO
	}

	logger.RegisterSink(lager.NewWriterSink(os.Stderr, logLevel))
	return logger
}

func cliCommands(exitHandler *exit_handler.ExitHandler, config *config.Config, logger lager.Logger) []cli.Command {
	input := os.Stdin
	output := output.New(os.Stdout)

	receptorClient := receptor.NewClient(config.Receptor())
	appRunner := app_runner.New(receptorClient, config.Target())

	timeprovider := timeprovider.NewTimeProvider()

	appRunnerCommandFactoryConfig := app_runner_command_factory.AppRunnerCommandFactoryConfig{
		AppRunner:             appRunner,
		DockerMetadataFetcher: docker_metadata_fetcher.New(docker_metadata_fetcher.NewDockerSessionFactory()),
		Output:                output,
		Timeout:               timeout(),
		Domain:                config.Target(),
		Env:                   os.Environ(),
		TimeProvider:          timeprovider,
		Logger:                logger,
	}

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)

	logReader := logs.NewLogReader(noaa.NewConsumer(setup_cli_helpers.LoggregatorUrl(config.Loggregator()), nil, nil))
	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(logReader, output)

	targetVerifier := target_verifier.New(receptor_client_factory.MakeReceptorClient)
	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, targetVerifier, input, output)

	appExaminer := app_examiner.New(receptorClient)
	appExaminerCommandFactory := app_examiner_command_factory.NewAppExaminerCommandFactory(appExaminer, output, timeprovider, exitHandler)

	return []cli.Command{
		appRunnerCommandFactory.MakeStartAppCommand(),
		appRunnerCommandFactory.MakeScaleAppCommand(),
		appRunnerCommandFactory.MakeStopAppCommand(),
		appRunnerCommandFactory.MakeRemoveAppCommand(),
		logsCommandFactory.MakeLogsCommand(),
		configCommandFactory.MakeTargetCommand(),
		appExaminerCommandFactory.MakeListAppCommand(),
		appExaminerCommandFactory.MakeStatusCommand(),
		appExaminerCommandFactory.MakeVisualizeCommand(),
	}
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
