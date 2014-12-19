package target_verifier_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"

	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
)

type fakereceptorClientBuilder struct {
	receptorClient receptor.Client
}

func (f *fakereceptorClientBuilder) Build(target string) receptor.Client {
	return f.receptorClient
}

var _ = Describe("targetVerifier", func() {
	Describe("RequiresAuth", func() {
		var fakeReceptorClient *fake_receptor.FakeClient
		var targets []string

		var fakeReceptorClientFactory = func(target string) receptor.Client {
			targets = append(targets, target)
			return fakeReceptorClient
		}

		BeforeEach(func() {
			fakeReceptorClient = &fake_receptor.FakeClient{}
			targets = []string{}
		})

		It("returns true if the receptor returns a authorization error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, errors.New("Unauthorized"))
			targetVerifier := target_verifier.New(fakeReceptorClientFactory)

			Expect(targetVerifier.RequiresAuth("http://receptor.mylattice.com")).To(BeTrue())
			Expect(targets).To(Equal([]string{"http://receptor.mylattice.com"}))
		})

		It("returns false if the receptor does not return an authorization error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)
			targetVerifier := target_verifier.New(fakeReceptorClientFactory)

			Expect(targetVerifier.RequiresAuth("http://receptor.mylattice.com")).To(BeFalse())
			Expect(targets).To(Equal([]string{"http://receptor.mylattice.com"}))
		})
	})
})
