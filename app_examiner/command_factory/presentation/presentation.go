package presentation

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
)

func ColorInstanceState(state string) string {
	colorFunc := colors.NoColor

	switch receptor.ActualLRPState(state) {
	case receptor.ActualLRPStateRunning:
		colorFunc = colors.Green
	case receptor.ActualLRPStateClaimed:
		colorFunc = colors.Yellow
	case receptor.ActualLRPStateUnclaimed:
		colorFunc = colors.Cyan
	case receptor.ActualLRPStateInvalid:
		colorFunc = colors.Red
	}

	return colorFunc(string(state))
}
