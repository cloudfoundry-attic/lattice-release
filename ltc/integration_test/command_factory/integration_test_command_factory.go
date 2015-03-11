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
		Usage:       "ltc test",
		Description: `ltc test verifies that the targeted lattice deployment is up and running.`,
		Action:      factory.runIntegrationTests,
		Flags:       testFlags,
	}

	return cliCommand
}

func (factory *IntegrationTestCommandFactory) runIntegrationTests(context *cli.Context) {
	factory.integrationTestRunner.Run(context.Duration("timeout"), context.Bool("verbose"))
}
