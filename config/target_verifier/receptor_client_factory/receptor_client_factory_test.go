package receptor_client_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/receptor"

	"github.com/pivotal-cf-experimental/lattice-cli/config/target_verifier/receptor_client_factory"
)

var _ = Describe("receptorClientFactory", func() {
	Describe("BuildReceptorClient", func() {
		It("returns a receptor with the target specified", func() {
			Expect(receptor_client_factory.BuildReceptorClient("mylattice.com")).To(Equal(receptor.NewClient("mylattice.com")))
		})
	})
})
