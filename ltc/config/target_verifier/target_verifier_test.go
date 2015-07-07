package target_verifier_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
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

		It("returns up=true, auth=true if the receptor does not return an error", func() {
			up, auth, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")

			Expect(up).To(BeTrue())
			Expect(auth).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(targets).To(ConsistOf("http://receptor.mylattice.com"))
		})

		It("returns up=true, auth=false if the receptor returns an authorization error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptor.Error{
				Type:    receptor.Unauthorized,
				Message: "Go home. You're not welcome here.",
			})

			up, auth, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")

			Expect(up).To(BeTrue())
			Expect(auth).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns up=true, auth=false, err=(the bubbled up error) if there is a receptor error other than unauthorized", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptor.Error{
				Type:    receptor.UnknownError,
				Message: "It all happened so fast... I just dunno what went wrong.",
			})

			up, auth, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")

			Expect(err).To(BeAssignableToTypeOf(receptor.Error{}))
			Expect(err).To(MatchError("It all happened so fast... I just dunno what went wrong."))
			Expect(up).To(BeTrue())
			Expect(auth).To(BeFalse())
		})

		// TODO: receptor really shouldn't give us non-receptor.Error
		It("returns up=false, auth=false, err=(the bubbled up error) if there is a non-receptor error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, errors.New("Couldn't connect to the receptor."))

			up, auth, err := targetVerifier.VerifyTarget("http://receptor.my-borked-lattice.com")

			Expect(err).To(MatchError("Couldn't connect to the receptor."))
			Expect(up).To(BeFalse())
			Expect(auth).To(BeFalse())
		})
	})
})
