package command_factory

import (
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/cluster_test"
	"github.com/codegangsta/cli"
)

type ClusterTestCommandFactory struct {
	clusterTestRunner cluster_test.ClusterTestRunner
}

func NewClusterTestCommandFactory(testRunner cluster_test.ClusterTestRunner) *ClusterTestCommandFactory {
	return &ClusterTestCommandFactory{testRunner}
}

func (factory *ClusterTestCommandFactory) MakeClusterTestCommand() cli.Command {

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

func (factory *ClusterTestCommandFactory) runIntegrationTests(context *cli.Context) {
	factory.clusterTestRunner.Run(context.Duration("timeout"), context.Bool("verbose"))
}
