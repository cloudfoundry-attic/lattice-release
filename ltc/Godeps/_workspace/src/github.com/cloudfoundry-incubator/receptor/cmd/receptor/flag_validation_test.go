package main_test

import (
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Flag Validation", func() {
	JustBeforeEach(func() {
		receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
		receptorProcess = ifrit.Background(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Context("when registerWithRouter is set", func() {
		BeforeEach(func() {
			receptorArgs.RegisterWithRouter = true
		})

		Context("when all necessary router registration parameters are set", func() {
			BeforeEach(func() {
				receptorArgs.DomainNames = "domain-names"
				receptorArgs.NatsAddresses = "nats-addresses"
			})
			It("does not exit", func() {
				Consistently(receptorRunner).ShouldNot(gexec.Exit())
			})
		})

		Context("when domain names is missing", func() {
			BeforeEach(func() {
				receptorArgs.DomainNames = ""
				receptorArgs.NatsAddresses = "nats-addresses"
			})
			It("exits with a non-zero exitcode", func() {
				Eventually(receptorRunner).Should(gexec.Exit(1))
			})
		})

		Context("when nats addresses is missing", func() {
			BeforeEach(func() {
				receptorArgs.DomainNames = "domain-names"
				receptorArgs.NatsAddresses = ""
			})
			It("exits with a non-zero exitcode", func() {
				Eventually(receptorRunner).Should(gexec.Exit(1))
			})
		})
	})

	Context("when registerWithRouter is not set", func() {
		BeforeEach(func() {
			receptorArgs.RegisterWithRouter = false
		})
		It("does not exit", func() {
			Consistently(receptorRunner).ShouldNot(gexec.Exit())
		})
	})
})
