package test_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
)

var _ = Describe("RegexSafeSay", func() {
	var outputBuffer *gbytes.Buffer

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
	})

	Describe("Say", func() {
		BeforeEach(func() {
			outputBuffer.Write([]byte(`match this \|?-^$.(){}`))
		})
		It("matches with regex-escaped characters", func() {
			Expect(outputBuffer).To(test_helpers.Say(`match this \|?-^$.(){}`))
		})
		It("negated match", func() {
			Expect(outputBuffer).NotTo(test_helpers.Say("match that"))
		})
		Context("when format string is passed with arguments", func() {
			It("matches with regex-escaped characters", func() {
				Expect(outputBuffer).To(test_helpers.Say(`match %s \|?-^$.(){}`, "this"))
			})
		})
	})

	Describe("SayLine", func() {
		BeforeEach(func() {
			outputBuffer.Write([]byte(`match this \|?-^$.(){}` + "\n"))
		})
		It("matches with regex-escaped characters", func() {
			Expect(outputBuffer).To(test_helpers.SayLine(`match this \|?-^$.(){}`))
		})
		It("negated match", func() {
			Expect(outputBuffer).NotTo(test_helpers.SayLine("match that"))
		})
		Context("when format string is passed with arguments", func() {
			It("matches with regex-escaped characters", func() {
				Expect(outputBuffer).To(test_helpers.SayLine(`match %s \|?-^$.(){}`, "this"))
			})
		})
	})

	Describe("SayIncorrectUsage", func() {
		It("matches", func() {
			outputBuffer.Write([]byte("Incorrect Usage"))
			Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
		})
		It("negated match", func() {
			outputBuffer.Write([]byte("say that"))
			Expect(outputBuffer).NotTo(test_helpers.SayIncorrectUsage())
		})
	})

	Describe("SayNewLine", func() {
		It("matches a newline character", func() {
			outputBuffer.Write([]byte("\n"))
			Expect(outputBuffer).To(test_helpers.SayNewLine())
		})
		It("negated match", func() {
			outputBuffer.Write([]byte("say that"))
			Expect(outputBuffer).NotTo(test_helpers.SayNewLine())
		})
	})
})
