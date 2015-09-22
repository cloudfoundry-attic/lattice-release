package receptor_client_test

import (
	"net/http"
	"reflect"

	"github.com/cloudfoundry-incubator/lattice/ltc/receptor_client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReceptorClientCreator", func() {
	var (
		receptorClientCreator *receptor_client.ProxyAwareCreator
	)

	BeforeEach(func() {
		receptorClientCreator = &receptor_client.ProxyAwareCreator{}
	})

	Context("#CreateReceptorClient", func() {
		It("should add Proxy func back to the http client", func() {
			client := receptorClientCreator.CreateReceptorClient("targethost")
			actualProxyFunc := client.GetClient().Transport.(*http.Transport).Proxy
			expectedProxyFunc := http.ProxyFromEnvironment
			Expect(reflect.ValueOf(actualProxyFunc).Pointer()).To(Equal(reflect.ValueOf(expectedProxyFunc).Pointer()))
		})
	})
})
