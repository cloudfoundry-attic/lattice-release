package presentation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory/presentation"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
)

var _ = Describe("Presentation", func() {
	Describe("ColorInstanceState", func() {
		It("colors RUNNING green", func() {
			Expect(presentation.ColorInstanceState(string(receptor.ActualLRPStateRunning))).To(Equal(colors.Green(string(receptor.ActualLRPStateRunning))))
		})

		It("colors CLAIMED yellow", func() {
			Expect(presentation.ColorInstanceState(string(receptor.ActualLRPStateClaimed))).To(Equal(colors.Yellow(string(receptor.ActualLRPStateClaimed))))
		})

		It("colors UNCLAIMED cyan", func() {
			Expect(presentation.ColorInstanceState(string(receptor.ActualLRPStateUnclaimed))).To(Equal(colors.Cyan(string(receptor.ActualLRPStateUnclaimed))))
		})

		It("colors INVALID red", func() {
			Expect(presentation.ColorInstanceState(string(receptor.ActualLRPStateInvalid))).To(Equal(colors.Red(string(receptor.ActualLRPStateInvalid))))
		})

	})
})
