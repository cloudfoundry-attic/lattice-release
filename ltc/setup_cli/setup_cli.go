package setup_cli

import (
	"os"
	"os/signal"

	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/config_helpers"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/lager"
)

const latticeCliHomeVar = "LATTICE_CLI_HOME"

var latticeVersion, diegoVersion string // provided by linker argument at compile-time

func NewCliApp() *cli.App {
	config := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(ltcConfigRoot())))

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	exitHandler := exit_handler.New(signalChan, os.Exit)
	go exitHandler.Run()

	targetVerifier := target_verifier.New(func(target string) receptor.Client {
		return receptor.NewClient(target)
	})

	return cli_app_factory.MakeCliApp(
		diegoVersion,
		latticeVersion,
		ltcConfigRoot(),
		exitHandler,
		config,
		logger(),
		targetVerifier,
		os.Stdout,
	)
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

func ltcConfigRoot() string {
	if os.Getenv(latticeCliHomeVar) != "" {
		return os.Getenv(latticeCliHomeVar)
	}

	return os.Getenv("HOME")
}
