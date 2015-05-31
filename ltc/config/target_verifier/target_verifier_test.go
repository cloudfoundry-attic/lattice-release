package target_verifier_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
)

var _ = Describe("TargetVerifier", func() {
	Describe("VerifyTarget", func() {
		var (
			fakeReceptorClient *fake_receptor.FakeClient
			targetVerifier     target_verifier.TargetVerifier
			targets            []string
		)

		fakeReceptorClientFactory := func(target string) receptor.Client {
			targets = append(targets, target)
			return fakeReceptorClient
		}

		BeforeEach(func() {
			fakeReceptorClient = &fake_receptor.FakeClient{}
			targetVerifier = target_verifier.New(fakeReceptorClientFactory)
			targets = []string{}
		})

		It("returns receptorUp=true, authorized=true if the receptor does not return an error", func() {
			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")

			Expect(receptorUp).To(BeTrue())
			Expect(authorized).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(targets).To(ConsistOf("http://receptor.mylattice.com"))
		})

		It("returns receptorUp=true, authorized=false if the receptor returns an authorization error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptor.Error{
				Type:    receptor.Unauthorized,
				Message: "Go home. You're not welcome here.",
			})

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

			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")

			Expect(err).To(MatchError("It all happened so fast... I just dunno what went wrong."))
			Expect(receptorUp).To(BeTrue())
			Expect(authorized).To(BeFalse())
		})

		It("returns receptorUp=false, authorized=false, err=(the bubbled up error) if there is a non-receptor error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, errors.New("Couldn't connect to the receptor."))

			receptorUp, authorized, err := targetVerifier.VerifyTarget("http://receptor.my-borked-lattice.com")

			Expect(err).To(MatchError("Couldn't connect to the receptor."))
			Expect(receptorUp).To(BeFalse())
			Expect(authorized).To(BeFalse())
		})
	})
})
