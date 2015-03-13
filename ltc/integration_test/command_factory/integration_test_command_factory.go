package command_factory

import (
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/codegangsta/cli"
)

type IntegrationTestCommandFactory struct {
	integrationTestRunner integration_test.IntegrationTestRunner
	output                *output.Output
}

func NewIntegrationTestCommandFactory(testRunner integration_test.IntegrationTestRunner, output *output.Output) *IntegrationTestCommandFactory {
	return &IntegrationTestCommandFactory{testRunner, output}
}

func (factory *IntegrationTestCommandFactory) MakeIntegrationTestCommand() cli.Command {

	testFlags := []cli.Flag{
		cli.DurationFlag{
			Name:  "timeout",
			Usage: "Duration of time tests will wait for lattice to respond",
			Value: time.Second * 30,
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Verbose mode",
		},
	}

	cliCommand := cli.Command{
		Name:        "test",
		Usage:       "Runs test suite against targeted lattice cluster",
		Description: "ltc test [-v] [--timeout=TIMEOUT]",
		Action:      factory.runIntegrationTests,
		Flags:       testFlags,
	}

	return cliCommand
}

func (factory *IntegrationTestCommandFactory) runIntegrationTests(context *cli.Context) {
	factory.integrationTestRunner.Run(context.Duration("timeout"), context.Bool("verbose"))
}
