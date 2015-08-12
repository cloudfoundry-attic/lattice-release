package cli_app_factory

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/graphical"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/cluster_test"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/docker_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/noaa"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"

	app_examiner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory"
	app_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	cluster_test_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/cluster_test/command_factory"
	config_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
	docker_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/docker_runner/command_factory"
	droplet_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory"
	logs_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/logs/command_factory"
	task_examiner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/task_examiner/command_factory"
	task_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/task_runner/command_factory"
)

var (
	nonTargetVerifiedCommandNames = map[string]struct{}{
		"target": {},
		"help":   {},
	}

	defaultAction = func(context *cli.Context) {
		args := context.Args()
		if len(args) > 0 {
			cli.ShowCommandHelp(context, args[0])
		} else {
			showAppHelp(context.App.Writer, appHelpTemplate(), context.App)
		}
	}
)

const (
	LtcUsage          = "Command line interface for Lattice."
	AppName           = "ltc"
	latticeCliAuthor  = "Pivotal"
	latticeCliHomeVar = "LATTICE_CLI_HOME"
	unknownCommand    = "ltc: '%s' is not a registered command. See 'ltc help'"
)

func init() {
	cli.HelpPrinter = ShowHelp
	cli.AppHelpTemplate = appHelpTemplate()
	cli.CommandHelpTemplate = `NAME:
   {{join .Names ", "}} - {{.Usage}}
{{with .ShortName}}
ALIAS:
   {{.Aliases}}
{{end}}
USAGE:
   {{.Description}}{{with .Flags}}
OPTIONS:
{{range .}}   {{.}}
{{end}}{{else}}
{{end}}`
}

func MakeCliApp(
	diegoVersion string,
	latticeVersion string,
	ltcConfigRoot string,
	exitHandler exit_handler.ExitHandler,
	config *config.Config,
	logger lager.Logger,
	targetVerifier target_verifier.TargetVerifier,
	cliStdout io.Writer,
) *cli.App {
	config.Load()
	app := cli.NewApp()
	app.Name = AppName
	app.Author = latticeCliAuthor
	app.Version = defaultVersion(diegoVersion, latticeVersion)
	app.Usage = LtcUsage
	app.Email = "cf-lattice@lists.cloudfoundry.org"

	ui := terminal.NewUI(os.Stdin, cliStdout, password_reader.NewPasswordReader(exitHandler))
	app.Writer = ui

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
			ui.SayLine(fmt.Sprintf("Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.\n\tUnderlying error: %s", err.Error()))
			return err
		} else if !authorized {
			ui.SayLine("Could not authenticate with the receptor. Please run ltc target with the correct credentials.")
			return errors.New("Could not authenticate with the receptor.")
		}
		return nil
	}

	app.Action = defaultAction
	app.CommandNotFound = func(c *cli.Context, command string) {
		ui.SayLine(fmt.Sprintf(unknownCommand, command))
		exitHandler.Exit(1)
	}
	app.Commands = cliCommands(ltcConfigRoot, exitHandler, config, logger, targetVerifier, ui)
	return app
}

func cliCommands(ltcConfigRoot string, exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, ui terminal.UI) []cli.Command {

	receptorClient := receptor.NewClient(config.Receptor())
	noaaConsumer := noaa.NewConsumer(LoggregatorUrl(config.Loggregator()), nil, nil)
	appRunner := app_runner.New(receptorClient, config.Target())

	clock := clock.NewClock()

	logReader := logs.NewLogReader(noaaConsumer)
	tailedLogsOutputter := console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(ui, logReader)

	taskExaminer := task_examiner.New(receptorClient)
	taskExaminerCommandFactory := task_examiner_command_factory.NewTaskExaminerCommandFactory(taskExaminer, ui, exitHandler)

	taskRunner := task_runner.New(receptorClient, taskExaminer)
	taskRunnerCommandFactory := task_runner_command_factory.NewTaskRunnerCommandFactory(taskRunner, ui, exitHandler)

	appExaminer := app_examiner.New(receptorClient, app_examiner.NewNoaaConsumer(noaaConsumer))
	graphicalVisualizer := graphical.NewGraphicalVisualizer(appExaminer)
	appExaminerCommandFactory := app_examiner_command_factory.NewAppExaminerCommandFactory(appExaminer, ui, clock, exitHandler, graphicalVisualizer, taskExaminer)

	appRunnerCommandFactoryConfig := app_runner_command_factory.AppRunnerCommandFactoryConfig{
		AppRunner:           appRunner,
		AppExaminer:         appExaminer,
		UI:                  ui,
		Domain:              config.Target(),
		Env:                 os.Environ(),
		Clock:               clock,
		Logger:              logger,
		TailedLogsOutputter: tailedLogsOutputter,
		ExitHandler:         exitHandler,
	}

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)

	dockerRunnerCommandFactoryConfig := docker_runner_command_factory.DockerRunnerCommandFactoryConfig{
		AppRunner:             appRunner,
		AppExaminer:           appExaminer,
		UI:                    ui,
		Domain:                config.Target(),
		Env:                   os.Environ(),
		Clock:                 clock,
		Logger:                logger,
		ExitHandler:           exitHandler,
		TailedLogsOutputter:   tailedLogsOutputter,
		DockerMetadataFetcher: docker_metadata_fetcher.New(docker_metadata_fetcher.NewDockerSessionFactory()),
	}
	dockerRunnerCommandFactory := docker_runner_command_factory.NewDockerRunnerCommandFactory(dockerRunnerCommandFactoryConfig)

	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(appExaminer, taskExaminer, ui, tailedLogsOutputter, exitHandler)

	clusterTestRunner := cluster_test.NewClusterTestRunner(config, ltcConfigRoot)
	clusterTestCommandFactory := cluster_test_command_factory.NewClusterTestCommandFactory(clusterTestRunner)

	blobStore := dav_blob_store.New(config.BlobStore())

	dropletRunner := droplet_runner.New(appRunner, taskRunner, config, blobStore, targetVerifier, appExaminer)
	cfIgnore := cf_ignore.New()
	dropletRunnerCommandFactory := droplet_runner_command_factory.NewDropletRunnerCommandFactory(*appRunnerCommandFactory, taskExaminer, dropletRunner, cfIgnore)

	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, ui, targetVerifier, dav_blob_store.Verifier{}, exitHandler)

	helpCommand := cli.Command{
		Name:        "help",
		Aliases:     []string{"h"},
		Usage:       "Shows a list of commands or help for one command",
		Description: "ltc help",
		Action:      defaultAction,
	}

	return []cli.Command{
		appExaminerCommandFactory.MakeCellsCommand(),
		dockerRunnerCommandFactory.MakeCreateAppCommand(),
		appRunnerCommandFactory.MakeSubmitLrpCommand(),
		logsCommandFactory.MakeDebugLogsCommand(),
		appExaminerCommandFactory.MakeListAppCommand(),
		logsCommandFactory.MakeLogsCommand(),
		appRunnerCommandFactory.MakeRemoveAppCommand(),
		appRunnerCommandFactory.MakeScaleAppCommand(),
		appExaminerCommandFactory.MakeStatusCommand(),
		taskRunnerCommandFactory.MakeSubmitTaskCommand(),
		configCommandFactory.MakeTargetCommand(),
		taskExaminerCommandFactory.MakeTaskCommand(),
		taskRunnerCommandFactory.MakeDeleteTaskCommand(),
		taskRunnerCommandFactory.MakeCancelTaskCommand(),
		clusterTestCommandFactory.MakeClusterTestCommand(),
		appRunnerCommandFactory.MakeUpdateRoutesCommand(),
		appExaminerCommandFactory.MakeVisualizeCommand(),
		dropletRunnerCommandFactory.MakeBuildDropletCommand(),
		dropletRunnerCommandFactory.MakeListDropletsCommand(),
		dropletRunnerCommandFactory.MakeLaunchDropletCommand(),
		dropletRunnerCommandFactory.MakeRemoveDropletCommand(),
		dropletRunnerCommandFactory.MakeImportDropletCommand(),
		dropletRunnerCommandFactory.MakeExportDropletCommand(),
		helpCommand,
	}
}

func LoggregatorUrl(loggregatorTarget string) string {
	return "ws://" + loggregatorTarget
}

func defaultVersion(diegoVersion, latticeVersion string) string {
	if diegoVersion == "" {
		diegoVersion = "unknown"
	}

	if latticeVersion == "" {
		latticeVersion = "development (not versioned)"
	}

	return fmt.Sprintf("%s (diego %s)", latticeVersion, diegoVersion)
}

func appHelpTemplate() string {
	return `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

AUTHOR(S): 
   {{range .Authors}}{{.}}
   {{end}}

COMMANDS: 
   {{range .Commands}}
  {{.SubTitle .Name}}{{range .CommandSubGroups}}
   {{range .}} {{.Name}}   {{.Description}}
   {{end}}{{end}}{{end}}
GLOBAL OPTIONS:
   --version, -v        Print the version 
   --help, -h           Show help 
`
}
