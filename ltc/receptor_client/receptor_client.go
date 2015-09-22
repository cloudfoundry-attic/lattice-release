package receptor_client

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
)

//go:generate counterfeiter -o fake_receptor_client_creator/fake_receptor_client_creator.go . Creator
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
