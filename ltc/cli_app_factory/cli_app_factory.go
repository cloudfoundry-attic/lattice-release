package cli_app_factory

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/noaa"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"

	app_examiner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory"
	app_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	config_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
	integration_test_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/integration_test/command_factory"
	logs_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/logs/command_factory"
)

var nonTargetVerifiedCommandNames = map[string]struct{}{
	config_command_factory.TargetCommandName: {},
	"help": {},
}

const (
	LtcUsage          = "Command line interface for Lattice."
	AppName           = "ltc"
	latticeCliAuthor  = "Pivotal"
	latticeCliHomeVar = "LATTICE_CLI_HOME"
)

func MakeCliApp(timeoutStr, latticeVersion, ltcConfigRoot string, exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, cliStdout io.Writer) *cli.App {
	config.Load()
	app := cli.NewApp()
	app.Name = AppName
	app.Author = latticeCliAuthor
	app.Version = defaultVersion(latticeVersion)
	app.Usage = LtcUsage
	app.Email = "lattice@cloudfoundry.org"

	ui := terminal.NewUI(os.Stdin, cliStdout, password_reader.NewPasswordReader(exitHandler))

	app.Commands = cliCommands(timeoutStr, ltcConfigRoot, exitHandler, config, logger, targetVerifier, ui)

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
			ui.Say(fmt.Sprintf("Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.\n\tUnderlying error: %s", err.Error()))
			return err
		} else if !authorized {
			ui.Say("Could not authenticate with the receptor. Please run ltc target with the correct credentials.")
			return errors.New("Could not authenticate with the receptor.")
		}
		return nil
	}

	return app
}

func cliCommands(timeoutStr, ltcConfigRoot string, exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, ui terminal.UI) []cli.Command {

	receptorClient := receptor.NewClient(config.Receptor())
	appRunner := docker_app_runner.New(receptorClient, config.Target())

	clock := clock.NewClock()

	logReader := logs.NewLogReader(noaa.NewConsumer(LoggregatorUrl(config.Loggregator()), nil, nil))
	tailedLogsOutputter := console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(ui, logReader)

	appExaminer := app_examiner.New(receptorClient)
	appExaminerCommandFactory := app_examiner_command_factory.NewAppExaminerCommandFactory(appExaminer, ui, clock, exitHandler)

	appRunnerCommandFactoryConfig := app_runner_command_factory.AppRunnerCommandFactoryConfig{
		AppRunner:             appRunner,
		AppExaminer:           appExaminer,
		DockerMetadataFetcher: docker_metadata_fetcher.New(docker_metadata_fetcher.NewDockerSessionFactory()),
		UI:                  ui,
		Timeout:             Timeout(timeoutStr),
		Domain:              config.Target(),
		Env:                 os.Environ(),
		Clock:               clock,
		Logger:              logger,
		TailedLogsOutputter: tailedLogsOutputter,
		ExitHandler:         exitHandler,
	}

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)

	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(ui, tailedLogsOutputter, exitHandler)

	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, ui, targetVerifier, exitHandler)

	testRunner := integration_test.NewIntegrationTestRunner(config, ltcConfigRoot)
	integrationTestCommandFactory := integration_test_command_factory.NewIntegrationTestCommandFactory(testRunner)

	return []cli.Command{
		appRunnerCommandFactory.MakeCreateAppCommand(),
		appRunnerCommandFactory.MakeCreateAppFromJsonCommand(),
		logsCommandFactory.MakeDebugLogsCommand(),
		appExaminerCommandFactory.MakeListAppCommand(),
		logsCommandFactory.MakeLogsCommand(appExaminer),
		appRunnerCommandFactory.MakeRemoveAppCommand(),
		appRunnerCommandFactory.MakeScaleAppCommand(),
		appExaminerCommandFactory.MakeStatusCommand(),
		configCommandFactory.MakeTargetCommand(),
		integrationTestCommandFactory.MakeIntegrationTestCommand(),
		appRunnerCommandFactory.MakeUpdateRoutesCommand(),
		appExaminerCommandFactory.MakeVisualizeCommand(),
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

func defaultVersion(latticeVersion string) string {
	if latticeVersion == "" {
		return "development (not versioned)"
	}
	return latticeVersion
}
