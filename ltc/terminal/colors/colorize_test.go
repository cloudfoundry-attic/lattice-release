package colors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
)

var _ = Describe("Colorize", func() {
	It("colors the text with printf-style syntax", func() {
		Expect(colors.Colorize("\x1b[98m", "%dxyz%s", 23, "happy")).To(Equal("\x1b[98m23xyzhappy\x1b[0m"))
	})

	It("colors the text without printf-style syntax", func() {
		Expect(colors.Colorize("\x1b[98m", "happy")).To(Equal("\x1b[98mhappy\x1b[0m"))
	})
})
