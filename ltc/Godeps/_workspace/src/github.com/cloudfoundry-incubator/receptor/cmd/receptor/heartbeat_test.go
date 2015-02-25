package main_test

import (
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Heartbeating", func() {
	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	It("heartbeats its presence to the BBS with the task handler URL", func() {
		var presence models.ReceptorPresence
		Eventually(func() error {
			var err error
			presence, err = bbs.Receptor()
			return err
		}).ShouldNot(HaveOccurred())

		Ω(presence.ReceptorID).ShouldNot(BeEmpty())
		Ω(presence.ReceptorURL).Should(Equal("http://" + receptorTaskHandlerAddress))
	})
})
