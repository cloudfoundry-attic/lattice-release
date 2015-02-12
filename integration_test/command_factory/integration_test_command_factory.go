package command_factory

import (
	"time"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/integration_test"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
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
			Usage: "How long whetstone will wait for docker apps to start",
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
		Usage:       "ltc test --domain=DOMAIN",
		Description: "The app formerly known as WHETSTONE",
		Action:      commandFactory.integrationTestRunnerCommand.runIntegrationTests,
		Flags:       testFlags,
	}

	return cliCommand
}

func (cmd *runIntegrationTestCommand) runIntegrationTests(ctx *cli.Context) {
	cmd.integrationTestRunner.Run(ctx.Duration("timeout"), ctx.Bool("verbose"))
}
