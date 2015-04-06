package presentation

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/receptor"
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
	case state == receptor.ActualLRPStateCrashed:
		colorFunc = colors.Red
	}

	return colorFunc(string(instanceInfo.State))
}

func PadAndColorInstanceState(instanceInfo app_examiner.InstanceInfo) string {
	padLength := 0
	if len(ColorInstanceState(instanceInfo)) < len(colors.NoColor("UNCLAIMED")) {
		padLength = len("UNCLAIMED") - len(instanceInfo.State)
	}

	return fmt.Sprintf("%s%s", ColorInstanceState(instanceInfo), strings.Repeat(" ", padLength))
}
