package receptor_client

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
)

type Creator interface {
	CreateReceptorClient(target string) receptor.Client
}

type ProxyAwareCreator struct{}

func (ProxyAwareCreator) CreateReceptorClient(target string) receptor.Client {
	receptorClient := receptor.NewClient(target)

	transport := receptorClient.GetClient().Transport
	httpTransport := transport.(*http.Transport)
	httpTransport.Proxy = http.ProxyFromEnvironment
	receptorClient.GetClient().Transport = httpTransport

	return receptorClient
}
