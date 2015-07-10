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
				Expect(postErr).NotTo(HaveOccurred())
			})

			It("has the correct data from the bbs", func() {
				domains, err := bbsClient.Domains()
				Expect(err).NotTo(HaveOccurred())

				Expect(domains).To(ConsistOf([]string{"domain-0"}))
			})
		})

		Context("with an infinite ttl (0)", func() {
			BeforeEach(func() {
				postErr = client.UpsertDomain("domain-0", 0)
				Expect(postErr).NotTo(HaveOccurred())

				postErr = client.UpsertDomain("domain-1", 1*time.Second)
				Expect(postErr).NotTo(HaveOccurred())
			})

			It("has the correct data from the bbs", func() {
				domains, err := bbsClient.Domains()
				Expect(err).NotTo(HaveOccurred())

				Expect(domains).To(ConsistOf([]string{"domain-0", "domain-1"}))

				time.Sleep(2 * time.Second)

				domains, err = bbsClient.Domains()
				Expect(err).NotTo(HaveOccurred())

				Expect(domains).To(ConsistOf([]string{"domain-0"}))
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
				err := client.UpsertDomain(d, time.Duration(100*(i+1))*time.Second)
				Expect(err).NotTo(HaveOccurred())
			}

			actualDomains, getErr = client.Domains()
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(actualDomains).To(HaveLen(2))
		})

		It("has the correct domains from the bbs", func() {
			Expect(expectedDomains).To(ConsistOf(actualDomains))
		})
	})
})
