package test_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
)

var _ = Describe("RegexSafeSay", func() {

	var gbytesBuffer *gbytes.Buffer

	BeforeEach(func() {
		gbytesBuffer = gbytes.NewBuffer()
	})

	Describe("Say", func() {
		It("matches with regex-escaped characters", func() {
			gbytesBuffer.Write([]byte(`match this \|?-^$.(){}`))

			Expect(gbytesBuffer).To(test_helpers.Say(`match this \|?-^$.(){}`))
		})

		It("negated match", func() {
			gbytesBuffer.Write([]byte("say that"))

			Expect(gbytesBuffer).ToNot(test_helpers.Say("different"))
		})
	})

	Describe("SayLine", func() {
		It("matches with regex-escaped characters", func() {
			gbytesBuffer.Write([]byte("sample\n"))

			Expect(gbytesBuffer).To(test_helpers.SayLine("sample"))
		})

		It("negated match", func() {
			gbytesBuffer.Write([]byte("no match"))

			Expect(gbytesBuffer).ToNot(test_helpers.SayLine("no match"))
		})
	})

	Describe("SayIncorrectUsage", func() {
		It("matches", func() {
			gbytesBuffer.Write([]byte("Incorrect Usage"))

			Expect(gbytesBuffer).To(test_helpers.SayIncorrectUsage())
		})

		It("negated match", func() {
			gbytesBuffer.Write([]byte("say that"))

			Expect(gbytesBuffer).ToNot(test_helpers.SayIncorrectUsage())
		})
	})

	Describe("SayNewLine", func() {
		It("match", func() {
			gbytesBuffer.Write([]byte("\n"))

			Expect(gbytesBuffer).To(test_helpers.SayNewLine())
		})

		It("negated match", func() {
			gbytesBuffer.Write([]byte("say that"))

			Expect(gbytesBuffer).ToNot(test_helpers.SayNewLine())
		})
	})
})
