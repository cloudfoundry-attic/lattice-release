package presentation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/presentation"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/receptor"
)

var _ = Describe("Presentation", func() {
	Describe("ColorInstanceState", func() {
		It("colors RUNNING green", func() {
			instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateRunning)}
			Expect(presentation.ColorInstanceState(instanceInfo)).To(Equal(colors.Green(string(receptor.ActualLRPStateRunning))))
		})

		It("colors CLAIMED yellow", func() {
			instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateClaimed)}
			Expect(presentation.ColorInstanceState(instanceInfo)).To(Equal(colors.Yellow(string(receptor.ActualLRPStateClaimed))))
		})

		Context("when there is a placement error", func() {
			It("colors UNCLAIMED red", func() {
				instanceInfo := app_examiner.InstanceInfo{
					State:          string(receptor.ActualLRPStateUnclaimed),
					PlacementError: "I misplaced my cells. Uh oh.",
				}

				Expect(presentation.ColorInstanceState(instanceInfo)).To(Equal(colors.Red(string(receptor.ActualLRPStateUnclaimed))))
			})
		})

		Context("when there is not a placement error", func() {
			It("colors UNCLAIMED cyan", func() {
				instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateUnclaimed)}
				Expect(presentation.ColorInstanceState(instanceInfo)).To(Equal(colors.Cyan(string(receptor.ActualLRPStateUnclaimed))))
			})
		})

		It("colors INVALID red", func() {
			instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateInvalid)}
			Expect(presentation.ColorInstanceState(instanceInfo)).To(Equal(colors.Red(string(receptor.ActualLRPStateInvalid))))
		})

		It("colors CRASHED red", func() {
			instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateCrashed)}
			Expect(presentation.ColorInstanceState(instanceInfo)).To(Equal(colors.Red(string(receptor.ActualLRPStateCrashed))))
		})
	})

	Describe("PadAndColorInstanceState", func() {
		It("pads and colors States shorter than UNCLAIMED", func() {
			instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateRunning)}
			Expect(presentation.PadAndColorInstanceState(instanceInfo)).To(Equal(colors.Green(string(receptor.ActualLRPStateRunning)) + "  "))
		})

		It("does not pad UNCLAIMED state", func() {
			instanceInfo := app_examiner.InstanceInfo{State: string(receptor.ActualLRPStateUnclaimed)}
			Expect(presentation.PadAndColorInstanceState(instanceInfo)).To(Equal(colors.Cyan(string(receptor.ActualLRPStateUnclaimed))))
		})
	})
})
