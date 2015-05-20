package command_factory

import (
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test"
	"github.com/codegangsta/cli"
)

type IntegrationTestCommandFactory struct {
	integrationTestRunner integration_test.IntegrationTestRunner
}

func NewIntegrationTestCommandFactory(testRunner integration_test.IntegrationTestRunner) *IntegrationTestCommandFactory {
	return &IntegrationTestCommandFactory{testRunner}
}

func (factory *IntegrationTestCommandFactory) MakeIntegrationTestCommand() cli.Command {

	testFlags := []cli.Flag{
		cli.DurationFlag{
			Name:  "timeout, t",
			Usage: "Duration of time tests will wait for lattice to respond",
			Value: time.Minute * 2,
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Verbose mode",
		},
		cli.BoolFlag{
			Name:  "cli-help",
			Usage: "Cli Help Tests",
		},
	}

	cliCommand := cli.Command{
		Name:        "test",
		Aliases:     []string{"te"},
		Usage:       "Runs test suite against targeted lattice cluster",
		Description: "ltc test [-v] [--timeout=TIMEOUT] [--cli-help]",
		Action:      factory.runIntegrationTests,
		Flags:       testFlags,
	}

	return cliCommand
}

func (factory *IntegrationTestCommandFactory) runIntegrationTests(context *cli.Context) {
	factory.integrationTestRunner.Run(context.Duration("timeout"), context.Bool("verbose"), context.Bool("cli-help"))
}
