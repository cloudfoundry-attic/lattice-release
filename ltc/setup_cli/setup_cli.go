package setup_cli

import (
	"os"
	"os/signal"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/config_helpers"
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler"
	"github.com/pivotal-golang/lager"

	"github.com/pivotal-cf-experimental/lattice-cli/cli_app_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier/receptor_client_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
)

func NewCliApp() *cli.App {
	config := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(userHome())))

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	exitHandler := exit_handler.New(signalChan, os.Exit)
	go exitHandler.Run()

	targetVerifier := target_verifier.New(receptor_client_factory.MakeReceptorClient)
	app := cli_app_factory.MakeCliApp(exitHandler, config, logger(), targetVerifier, output.New(os.Stdout))
	return app
}

func logger() lager.Logger {
	logger := lager.NewLogger("ltc")
	var logLevel lager.LogLevel

	if os.Getenv("LTC_LOG_LEVEL") == "DEBUG" {
		logLevel = lager.DEBUG
	} else {
		logLevel = lager.INFO
	}

	logger.RegisterSink(lager.NewWriterSink(os.Stderr, logLevel))
	return logger
}

func userHome() string {
	if os.Getenv("LATTICE_CLI_HOME") != "" {
		return os.Getenv("LATTICE_CLI_HOME")
	}

	return os.Getenv("HOME")
}
