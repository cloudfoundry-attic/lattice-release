package lager_test

import (
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ReconfigurableSink", func() {
	var (
		testSink *lagertest.TestSink

		sink *lager.ReconfigurableSink
	)

	BeforeEach(func() {
		testSink = lagertest.NewTestSink()

		sink = lager.NewReconfigurableSink(testSink, lager.INFO)
	})

	It("returns the current level", func() {
		Ω(sink.GetMinLevel()).Should(Equal(lager.INFO))
	})

	Context("when logging above the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.INFO, []byte("hello world"))
		})

		It("writes to the given sink", func() {
			Ω(testSink.Buffer()).Should(gbytes.Say("hello world\n"))
		})
	})

	Context("when logging below the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
		})

		It("does not write to the given writer", func() {
			Ω(testSink.Buffer().Contents()).Should(BeEmpty())
		})
	})

	Context("when reconfigured to a new log level", func() {
		BeforeEach(func() {
			sink.SetMinLevel(lager.DEBUG)
		})

		It("writes logs above the new log level", func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
			Ω(testSink.Buffer()).Should(gbytes.Say("hello world\n"))
		})

		It("returns the newly updated level", func() {
			Ω(sink.GetMinLevel()).Should(Equal(lager.DEBUG))
		})
	})
})
