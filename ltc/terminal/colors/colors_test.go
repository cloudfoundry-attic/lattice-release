package colors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
)

var _ = Describe("colors", func() {

	itShouldNotColorizeWhitespace := func(colorizer func(text string) string) {
		It("returns a string without color codes when only whitespace is passed in", func() {
			Expect(colorizer("  ")).To(Equal("  "))
			Expect(colorizer("\n")).To(Equal("\n"))
			Expect(colorizer("\t")).To(Equal("\t"))
			Expect(colorizer("\r")).To(Equal("\r"))
		})
	}

	Describe("Red", func() {
		It("adds the red color code", func() {
			Expect(colors.Red("ERROR NOT GOOD")).To(Equal("\x1b[91mERROR NOT GOOD\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.Red)
	})

	Describe("Green", func() {
		It("adds the green color code", func() {
			Expect(colors.Green("TOO GOOD")).To(Equal("\x1b[32mTOO GOOD\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.Green)
	})

	Describe("Cyan", func() {
		It("adds the cyan color code", func() {
			Expect(colors.Cyan("INFO")).To(Equal("\x1b[36mINFO\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.Cyan)
	})

	Describe("Yellow", func() {
		It("adds the yellow color code", func() {
			Expect(colors.Yellow("INFO")).To(Equal("\x1b[33mINFO\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.Yellow)
	})

	Describe("Gray", func() {
		It("adds the gray color code", func() {
			Expect(colors.Gray("INFO")).To(Equal("\x1b[90mINFO\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.Gray)
	})

	Describe("Bold", func() {
		It("adds the bold color code", func() {
			Expect(colors.Bold("Bold")).To(Equal("\x1b[1mBold\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.Bold)
	})

	Describe("PurpleUnderline", func() {
		It("adds the purple underlined color code", func() {
			Expect(colors.PurpleUnderline("PURPLE UNDERLINE")).To(Equal("\x1b[35;4mPURPLE UNDERLINE\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.PurpleUnderline)
	})

	Describe("NoColor", func() {
		It("adds no color code", func() {
			Expect(colors.NoColor("None")).To(Equal("\x1b[0mNone\x1b[0m"))
		})

		itShouldNotColorizeWhitespace(colors.NoColor)
	})

})
