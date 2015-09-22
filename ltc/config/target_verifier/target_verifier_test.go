package target_verifier_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/receptor_client/fake_receptor_client_creator"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
)

var _ = Describe("TargetVerifier", func() {
	Describe("VerifyTarget", func() {
		var (
			fakeReceptorClient        *fake_receptor.FakeClient
			fakeReceptorClientCreator *fake_receptor_client_creator.FakeCreator
			targetVerifier            target_verifier.TargetVerifier
		)

		BeforeEach(func() {
			fakeReceptorClient = &fake_receptor.FakeClient{}
			fakeReceptorClientCreator = &fake_receptor_client_creator.FakeCreator{}
			fakeReceptorClientCreator.CreateReceptorClientReturns(fakeReceptorClient)
			targetVerifier = target_verifier.New(fakeReceptorClientCreator)
		})

		It("returns up=true, auth=true if the receptor does not return an error", func() {
			up, auth, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(up).To(BeTrue())
			Expect(auth).To(BeTrue())
			Expect(fakeReceptorClientCreator.CreateReceptorClientCallCount()).To(Equal(1))
			Expect(fakeReceptorClientCreator.CreateReceptorClientArgsForCall(0)).To(Equal("http://receptor.mylattice.com"))
		})

		It("returns up=true, auth=false if the receptor returns an authorization error", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptor.Error{
				Type:    receptor.Unauthorized,
				Message: "Go home. You're not welcome here.",
			})

			up, auth, err := targetVerifier.VerifyTarget("http://receptor.mylattice.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(up).To(BeTrue())
			Expect(auth).To(BeFalse())
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
