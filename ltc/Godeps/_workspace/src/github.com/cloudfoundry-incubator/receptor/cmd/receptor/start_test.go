package main_test

import (
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Starting Receptor", func() {
	Context("when etcd is down", func() {
		BeforeEach(func() {
			etcdRunner.Stop()
			receptorProcess = ginkgomon.Invoke(receptorRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(receptorProcess)
			etcdRunner.Start()
		})

		It("starts", func() {
			Eventually(receptorProcess.Ready()).Should(BeClosed())
		})
	})
})
