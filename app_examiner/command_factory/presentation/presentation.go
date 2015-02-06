package presentation

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
)

func ColorInstanceState(instanceInfo app_examiner.InstanceInfo) string {
	colorFunc := colors.NoColor

	state := receptor.ActualLRPState(instanceInfo.State)
	switch {
	case state == receptor.ActualLRPStateRunning:
		colorFunc = colors.Green
	case state == receptor.ActualLRPStateClaimed:
		colorFunc = colors.Yellow
	case state == receptor.ActualLRPStateUnclaimed && instanceInfo.PlacementError == "":
		colorFunc = colors.Cyan
	case state == receptor.ActualLRPStateUnclaimed && instanceInfo.PlacementError != "":
		colorFunc = colors.Red
	case state == receptor.ActualLRPStateInvalid:
		colorFunc = colors.Red
	}

	return colorFunc(string(instanceInfo.State))
}
