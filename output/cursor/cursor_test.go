package cursor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/lattice-cli/output/cursor"
)

var _ = Describe("cursor", func() {
	Describe("Up", func() {
		It("moves the cursor up N lines", func() {
			Expect(cursor.Up(5)).To(Equal("\033[5A"))
		})
	})

	Describe("ClearToEndOfLine", func() {
		It("clears the line after the cursor", func() {
			Expect(cursor.ClearToEndOfLine()).To(Equal("\033[0K"))
		})
	})

	Describe("ClearToEndOfDisplay", func() {
		It("clears everything below the cursor", func() {
			Expect(cursor.ClearToEndOfDisplay()).To(Equal("\033[0J"))
		})
	})

	Describe("Show", func() {
		It("shows the cursor", func() {
			Expect(cursor.Show()).To(Equal("\033[?25h"))
		})
	})

	Describe("Hide", func() {
		It("hides the cursor", func() {
			Expect(cursor.Hide()).To(Equal("\033[?25l"))
		})
	})
})
