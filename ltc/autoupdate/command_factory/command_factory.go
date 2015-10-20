package command_factory

import (
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

type SyncCommandFactory struct {
	config      *config_package.Config
	ui          terminal.UI
	exitHandler exit_handler.ExitHandler

	arch    string
	ltcPath string

	sync Sync
}

//go:generate counterfeiter -o mocks/fake_sync.go . Sync
type Sync interface {
	SyncLTC(ltcPath string, arch string, config *config_package.Config) error
}

func NewSyncCommandFactory(config *config_package.Config, ui terminal.UI, exitHandler exit_handler.ExitHandler, arch string, ltcPath string, sync Sync) *SyncCommandFactory {
	return &SyncCommandFactory{config, ui, exitHandler, arch, ltcPath, sync}
}

func (f *SyncCommandFactory) MakeSyncCommand() cli.Command {
	return cli.Command{
		Name:        "sync",
		Usage:       "Updates ltc to the latest version available in the targeted Lattice cluster",
		Description: "ltc sync",
		Action:      f.syncLTC,
	}
}

func (f *SyncCommandFactory) syncLTC(context *cli.Context) {
	var architecture string
	switch f.arch {
	case "darwin":
		architecture = "osx"
	case "linux":
		architecture = "linux"
	default:
		f.ui.SayLine("Error: Unknown architecture %s. Sync not supported.", f.arch)
		f.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	if f.ltcPath == "" {
		f.ui.SayLine("Error: Unable to locate the ltc binary. Sync not supported.")
		f.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	if f.config.Target() == "" {
		f.ui.SayLine("Error: Must be targeted to sync.")
		f.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	err := f.sync.SyncLTC(f.ltcPath, architecture, f.config)
	if err != nil {
		f.ui.SayLine("Error: " + err.Error())
		f.exitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	f.ui.SayLine("Updated ltc to the latest version.")
}
