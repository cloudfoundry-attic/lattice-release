package main_test

import (
	"github.com/apcera/nats"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("registrations", func() {
	JustBeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	Context("when the server attempts to connect to NATS", func() {
		AfterEach(func() {
			ginkgomon.Kill(receptorProcess)
		})

		Context("and NATS is accessible", func() {
			var registrations chan *nats.Msg

			BeforeEach(func() {
				registrations = make(chan *nats.Msg, 1)
				natsClient.Subscribe("router.register", func(msg *nats.Msg) {
					registrations <- msg
				})
			})

			It("announces it's location", func() {
				Eventually(registrations).Should(Receive())
			})
		})

		Context("and NATS is inaccessible", func() {
			BeforeEach(func() {
				ginkgomon.Kill(natsGroupProcess)
			})

			It("does not exit", func() {
				Consistently(receptorRunner).ShouldNot(gexec.Exit())
			})
		})

		Context("and NATS becomes accessible later", func() {
			var registrations chan *nats.Msg
			BeforeEach(func() {
				ginkgomon.Kill(natsGroupProcess)
			})

			JustBeforeEach(func() {
				Eventually(receptorRunner).Should(gbytes.Say("connecting-to-nats-failed"))
				natsGroupProcess = ginkgomon.Invoke(newNatsGroup())

				registrations = make(chan *nats.Msg, 1)
				natsClient.Subscribe("router.register", func(msg *nats.Msg) {
					registrations <- msg
				})
			})

			It("starts announcing its location again", func() {
				Eventually(registrations).Should(Receive())
			})
		})
	})

	Context("when a server is sent SIGINT after the hearbeat has started", func() {
		JustBeforeEach(func() {
			Eventually(receptorRunner).Should(gbytes.Say("nats-heartbeat.started"))
			ginkgomon.Interrupt(receptorProcess)
		})

		Context("and NATS is accessible", func() {
			var unregistrations chan *nats.Msg

			BeforeEach(func() {
				unregistrations = make(chan *nats.Msg, 1)
				natsClient.Subscribe("router.unregister", func(msg *nats.Msg) {
					unregistrations <- msg
				})
			})

			It("broadcasts an unregister message", func() {
				Eventually(unregistrations).Should(Receive())
			})
		})
	})
})
