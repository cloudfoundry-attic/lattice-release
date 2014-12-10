package colors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/lattice-cli/colors"
)

var _ = Describe("colors", func() {
	Describe("Red", func() {
		It("adds the red color code", func() {
			Expect(colors.Red("ERROR NOT GOOD")).To(Equal("\x1b[91mERROR NOT GOOD\x1b[0m"))
		})
	})

	Describe("Green", func() {
		It("adds the green color code", func() {
			Expect(colors.Green("TOO GOOD")).To(Equal("\x1b[32mTOO GOOD\x1b[0m"))
		})
	})

	Describe("Cyan", func() {
		It("adds the cyan color code", func() {
			Expect(colors.Cyan("INFO")).To(Equal("\x1b[36mINFO\x1b[0m"))
		})
	})

	Describe("Yellow", func() {
		It("adds the yellow color code", func() {
			Expect(colors.Yellow("INFO")).To(Equal("\x1b[33mINFO\x1b[0m"))
		})
	})
})
