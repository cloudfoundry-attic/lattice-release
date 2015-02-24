package command_factory

import (
	"time"

	"github.com/cloudfoundry-incubator/lattice/cli/integration_test"
	"github.com/cloudfoundry-incubator/lattice/cli/output"
	"github.com/codegangsta/cli"
)

type IntegrationTestCommandFactory struct {
	integrationTestRunnerCommand *runIntegrationTestCommand
}

type runIntegrationTestCommand struct {
	integrationTestRunner integration_test.IntegrationTestRunner
	output                *output.Output
}

func NewIntegrationTestCommandFactory(testRunner integration_test.IntegrationTestRunner, output *output.Output) *IntegrationTestCommandFactory {
	return &IntegrationTestCommandFactory{&runIntegrationTestCommand{testRunner, output}}
}

func (commandFactory *IntegrationTestCommandFactory) MakeIntegrationTestCommand() cli.Command {

	testFlags := []cli.Flag{
		cli.DurationFlag{
			Name:  "timeout",
			Usage: "How long tests will wait for docker apps to start",
			Value: time.Second * 30,
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Whether tests should run in verbose mode",
		},
	}

	cliCommand := cli.Command{
		Name:        "test",
		ShortName:   "t",
		Usage:       "ltc test",
		Description: `ltc test verifies that the targeted lattice deployment is up and running.`,
		Action:      commandFactory.integrationTestRunnerCommand.runIntegrationTests,
		Flags:       testFlags,
	}

	return cliCommand
}

func (cmd *runIntegrationTestCommand) runIntegrationTests(context *cli.Context) {
	cmd.integrationTestRunner.Run(context.Duration("timeout"), context.Bool("verbose"))
}
