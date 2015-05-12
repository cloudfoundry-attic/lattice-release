package cli_app_factory

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/graphical"
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

type CmdWithTheme struct {
	cmd   cli.Command
	theme int
}

func getCommands(cmds []CmdWithTheme) []cli.Command {

	var retCmd []cli.Command
	for _, c := range cmds {
		retCmd = append(retCmd, c.cmd)
	}
	return retCmd
}

//Some variables for thememed help
var (
	AppHelpThemedTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]
VERSION:
   {{.Version}}
AUTHOR(S): 
   {{range .Authors}}{{ . }}
   {{end}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`
	listOfCmds   []CmdWithTheme
	listOfThemes = []string{"Target Lattice", "Create and Modify Apps", "Stream Logs", "See Whats Running", "Advanced", "Help and Debug"}
	cliappWriter io.Writer
)

const (
	themeLatticeTarget = iota
	themeCreateModify
	themeLogs
	themeSeeWhatsRunning
	themeAdvanced
	themeHelpDebug
	themeCOUNT //This should always be the last entry in this const section
)

func MakeCliApp(latticeVersion, ltcConfigRoot string, exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, cliStdout io.Writer) *cli.App {
	config.Load()
	app := cli.NewApp()
	app.Name = AppName
	app.Author = latticeCliAuthor
	app.Version = defaultVersion(latticeVersion)
	app.Usage = LtcUsage
	app.Email = "cf-lattice@lists.cloudfoundry.org"

	ui := terminal.NewUI(os.Stdin, cliStdout, password_reader.NewPasswordReader(exitHandler))

	listOfCmds = cliCommands(ltcConfigRoot, exitHandler, config, logger, targetVerifier, ui)

	app.Commands = getCommands(listOfCmds)

	//Over-write app.Writer to the supplied terminal Writer cli defautls it to stdout
	app.Writer = cliStdout
	cliappWriter = app.Writer
	//Assign a custom helper function
	cli.HelpPrinter = PrintThemedHelp

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

func cliCommands(ltcConfigRoot string, exitHandler exit_handler.ExitHandler, config *config.Config, logger lager.Logger, targetVerifier target_verifier.TargetVerifier, ui terminal.UI) []CmdWithTheme {

	receptorClient := receptor.NewClient(config.Receptor())
	noaaConsumer := noaa.NewConsumer(LoggregatorUrl(config.Loggregator()), nil, nil)
	appRunner := docker_app_runner.New(receptorClient, config.Target())

	clock := clock.NewClock()

	logReader := logs.NewLogReader(noaaConsumer)
	tailedLogsOutputter := console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(ui, logReader)

	appExaminer := app_examiner.New(receptorClient, app_examiner.NewNoaaConsumer(noaaConsumer))
	graphicalVisualizer := graphical.NewGraphicalVisualizer(appExaminer)
	appExaminerCommandFactory := app_examiner_command_factory.NewAppExaminerCommandFactory(appExaminer, ui, clock, exitHandler, graphicalVisualizer)

	appRunnerCommandFactoryConfig := app_runner_command_factory.AppRunnerCommandFactoryConfig{
		AppRunner:             appRunner,
		AppExaminer:           appExaminer,
		DockerMetadataFetcher: docker_metadata_fetcher.New(docker_metadata_fetcher.NewDockerSessionFactory()),
		UI:                  ui,
		Domain:              config.Target(),
		Env:                 os.Environ(),
		Clock:               clock,
		Logger:              logger,
		TailedLogsOutputter: tailedLogsOutputter,
		ExitHandler:         exitHandler,
	}

	appRunnerCommandFactory := app_runner_command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)

	logsCommandFactory := logs_command_factory.NewLogsCommandFactory(appExaminer, ui, tailedLogsOutputter, exitHandler)

	configCommandFactory := config_command_factory.NewConfigCommandFactory(config, ui, targetVerifier, exitHandler)

	testRunner := integration_test.NewIntegrationTestRunner(config, ltcConfigRoot)
	integrationTestCommandFactory := integration_test_command_factory.NewIntegrationTestCommandFactory(testRunner)

	return []CmdWithTheme{
		{configCommandFactory.MakeTargetCommand(), themeLatticeTarget},
		{appRunnerCommandFactory.MakeCreateAppCommand(), themeCreateModify},
		{appRunnerCommandFactory.MakeUpdateRoutesCommand(), themeCreateModify},
		{appRunnerCommandFactory.MakeRemoveAppCommand(), themeCreateModify},
		{appRunnerCommandFactory.MakeScaleAppCommand(), themeCreateModify},
		{logsCommandFactory.MakeLogsCommand(), themeLogs},
		{appExaminerCommandFactory.MakeListAppCommand(), themeSeeWhatsRunning},
		{appExaminerCommandFactory.MakeStatusCommand(), themeSeeWhatsRunning},
		{appExaminerCommandFactory.MakeVisualizeCommand(), themeSeeWhatsRunning},
		{appRunnerCommandFactory.MakeCreateLrpCommand(), themeAdvanced},
		{logsCommandFactory.MakeDebugLogsCommand(), themeHelpDebug},
		{integrationTestCommandFactory.MakeIntegrationTestCommand(), themeHelpDebug},
	}
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

func PrintThemedHelp(templ string, data interface{}) {
	if strings.Contains(templ, "GLOBAL OPTIONS") {
		funcMap := template.FuncMap{
			"join": strings.Join,
		}
		w := tabwriter.NewWriter(cliappWriter, 0, 8, 1, '\t', 0)
		t := template.Must(template.New("helptheam").Funcs(funcMap).Parse(AppHelpThemedTemplate))
		err := t.Execute(w, data)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "COMMANDS:\t\n")
		for i := 0; i < themeCOUNT; i++ {
			fmt.Fprintf(w, "%s:\t\n", listOfThemes[i])
			for _, c := range listOfCmds {
				if c.theme == i {
					fmt.Fprintf(w, "   %s\t%s\n", strings.Join(c.cmd.Names(), ", "), c.cmd.Usage)

				}
			}
			fmt.Fprintln(w, "\t")

		}
		w.Flush()

	} else {
		// default to codegangsta's help screen for others commands
		funcMap := template.FuncMap{
			"join": strings.Join,
		}
		w := tabwriter.NewWriter(cliappWriter, 0, 8, 1, '\t', 0)
		t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))
		err := t.Execute(w, data)
		if err != nil {
			panic(err)
		}
		w.Flush()
	}
}
