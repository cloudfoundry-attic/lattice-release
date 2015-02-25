package main_test

import (
	"time"

	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain API", func() {
	BeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Describe("PUT /v1/domains/:domain", func() {
		var postErr error

		Context("with a ttl > 0", func() {
			BeforeEach(func() {
				postErr = client.UpsertDomain("domain-0", 100*time.Second)
				Ω(postErr).ShouldNot(HaveOccurred())
			})

			It("has the correct data from the bbs", func() {
				domains, err := bbs.Domains()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(domains).Should(ConsistOf([]string{"domain-0"}))
			})
		})

		Context("with an infinite ttl (0)", func() {
			BeforeEach(func() {
				postErr = client.UpsertDomain("domain-0", 0)
				Ω(postErr).ShouldNot(HaveOccurred())

				postErr = client.UpsertDomain("domain-1", 1*time.Second)
				Ω(postErr).ShouldNot(HaveOccurred())
			})

			It("has the correct data from the bbs", func() {
				domains, err := bbs.Domains()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(domains).Should(ConsistOf([]string{"domain-0", "domain-1"}))

				time.Sleep(2 * time.Second)

				domains, err = bbs.Domains()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(domains).Should(ConsistOf([]string{"domain-0"}))
			})
		})
	})

	Describe("GET /v1/domains", func() {
		var expectedDomains []string
		var actualDomains []string
		var getErr error

		BeforeEach(func() {
			expectedDomains = []string{"domain-0", "domain-1"}
			for i, d := range expectedDomains {
				err := bbs.UpsertDomain(d, 100*(i+1))
				Ω(err).ShouldNot(HaveOccurred())
			}

			actualDomains, getErr = client.Domains()
		})

		It("responds without error", func() {
			Ω(getErr).ShouldNot(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Ω(actualDomains).Should(HaveLen(2))
		})

		It("has the correct domains from the bbs", func() {
			Ω(expectedDomains).Should(ConsistOf(actualDomains))
		})
	})
})
