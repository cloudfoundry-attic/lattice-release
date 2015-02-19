package target_verifier_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"

	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier"
)

var _ = Describe("targetVerifier", func() {
	Describe("ValidateAuthorization", func() {
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

		It("returns receptorUp=true, authorized=true if the receptor does not return an error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)
			targetVerifier := target_verifier.New(fakeReceptorClientFactory)

			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")
			Expect(receptorUp).To(BeTrue())
			Expect(authorized).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(targets).To(Equal([]string{"http://receptor.mylattice.com"}))
		})

		It("returns receptorUp=true, authorized=false if the receptor returns an authorization error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptor.Error{
				Type:    receptor.Unauthorized,
				Message: "Go home. You're not welcome here.",
			})
			targetVerifier := target_verifier.New(fakeReceptorClientFactory)

			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")
			Expect(receptorUp).To(BeTrue())
			Expect(authorized).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns receptorUp=true, authorized=false, err=(the bubbled up error) if there is a receptor error other than unauthorized", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptor.Error{
				Type:    receptor.UnknownError,
				Message: "It all happened so fast... I just dunno what went wrong.",
			})
			targetVerifier := target_verifier.New(fakeReceptorClientFactory)

			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")
			Expect(receptorUp).To(BeTrue())
			Expect(authorized).To(BeFalse())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("It all happened so fast... I just dunno what went wrong."))
		})

		It("returns receptorUp=false, authorized=false, err=(the bubbled up error) if there is a non-receptor error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, errors.New("Couldn't connect to the receptor."))
			targetVerifier := target_verifier.New(fakeReceptorClientFactory)

			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.my-borked-lattice.com")
			Expect(receptorUp).To(BeFalse())
			Expect(authorized).To(BeFalse())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Couldn't connect to the receptor."))

		})
	})
})
