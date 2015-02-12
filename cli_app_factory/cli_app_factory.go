package cli_app_factory

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/noaa"
	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_app_runner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler"
	"github.com/pivotal-cf-experimental/lattice-cli/integration_test"
	"github.com/pivotal-cf-experimental/lattice-cli/logs"
	"github.com/pivotal-cf-experimental/lattice-cli/logs/console_tailed_logs_outputter"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"

	app_examiner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory"
	app_runner_command_factory "github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
	config_command_factory "github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
	integration_test_command_factory "github.com/pivotal-cf-experimental/lattice-cli/integration_test/command_factory"
	logs_command_factory "github.com/pivotal-cf-experimental/lattice-cli/logs/command_factory"
)

var nonTargetVerifiedCommandNames = map[string]struct{}{
	config_command_factory.TargetCommandName: {},
	"help": {},
}

const (
	LtcUsage   = "Command line interface for Lattice."
	AppName    = "ltc"
	timeoutVar = "LATTICE_CLI_TIMEOUT"
)

func MakeCliApp(exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, output *output.Output) *cli.App {
	config.Load()
	app := cli.NewApp()
	app.Name = AppName
	app.Author = "Pivotal"
	app.Usage = LtcUsage
	app.Commands = cliCommands(exitHandler, config, logger, targetVerifier, output)

	app.Before = func(context *cli.Context) error {
		args := context.Args()
		command := app.Command(args.First())

		if command == nil {
			return nil
		}

		if _, ok := nonTargetVerifiedCommandNames[command.Name]; ok || len(args) == 0 {
			return nil
		}

		if receptorUp, authorized, err := targetVerifier.VerifyTarget(config.Receptor()); !receptorUp {
			output.Say(fmt.Sprintf("Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.\n\tUnderlying error: %s", err.Error()))
			return err
		} else if !authorized {
			output.Say("Could not authenticate with the receptor. Please run ltc target with the correct credentials.")
			return errors.New("Could not authenticate with the receptor.")
		}
		return nil
	}

	return app
}

func cliCommands(exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, output *output.Output) []cli.Command {
	input := os.Stdin

	receptorClient := receptor.NewClient(config.Receptor())
	appRunner := docker_app_runner.New(receptorClient, config.Target())

	clock := clock.NewClock()

	logReader := logs.NewLogReader(noaa.NewConsumer(LoggregatorUrl(config.Loggregator()), nil, nil))
	tailedLogsOutputter := console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(output, logReader)

	appRunnerCommandFactoryConfig := app_runner_command_factory.AppRunnerCommandFactoryConfig{
		AppRunner:             appRunner,
		DockerMetadataFetcher: docker_metadata_fetcher.New(docker_metadata_fetcher.NewDockerSessionFactory()),
		Output:                output,
		Timeout:               Timeout(os.Getenv(timeoutVar)),
		Domain:                config.Target(),
		Env:                   os.Environ(),
		Clock:                 clock,
		Logger:                logger,
		TailedLogsOutputter:   tailedLogsOutputter,
	}

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)

	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(output, tailedLogsOutputter, exitHandler)

	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, targetVerifier, input, output, exitHandler)

	appExaminer := app_examiner.New(receptorClient)
	appExaminerCommandFactory := app_examiner_command_factory.NewAppExaminerCommandFactory(appExaminer, output, clock, exitHandler)

	testRunner := integration_test.NewIntegrationTestRunner(output, config)
	integrationTestCommandFactory := integration_test_command_factory.NewIntegrationTestCommandFactory(testRunner, output)

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
		integrationTestCommandFactory.MakeIntegrationTestCommand(),
	}
}

func Timeout(timeoutEnv string) time.Duration {
	if timeout, err := strconv.Atoi(timeoutEnv); err == nil {
		return time.Second * time.Duration(timeout)
	}

	return time.Minute
}

func LoggregatorUrl(loggregatorTarget string) string {
	return "ws://" + loggregatorTarget
}
