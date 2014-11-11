package cli_config

import (
	"os"

	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/app_runner"
	"github.com/pivotal-golang/lager"

	start_app_command_factory "github.com/pivotal-cf-experimental/diego-edge-cli/app_runner/command_factory"
)

func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Diego"
	app.Usage = "Command line interface for diego."

	etcdUrl := os.Getenv("DIEGO_EDGE_ETCD")

	if etcdUrl == "" {
		panic("DIEGO_EDGE_ETCD environment variable was not set")
	}

	adapter := etcdstoreadapter.NewETCDStoreAdapter([]string{etcdUrl}, workpool.NewWorkPool(20))
	err := adapter.Connect()
	if err != nil {
		panic(err)
	}

	myBbs := bbs.NewBBS(adapter, timeprovider.NewTimeProvider(), lager.NewLogger("diego-edge"))

	appRunner := app_runner.NewDiegoAppRunner(myBbs)

	app.Commands = []cli.Command{
		start_app_command_factory.NewStartAppCommandFactory(appRunner, os.Stdout).MakeCommand(),
	}
	return app
}
