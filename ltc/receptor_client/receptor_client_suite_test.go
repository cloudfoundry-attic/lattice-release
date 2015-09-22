package receptor_client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestReceptorClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReceptorClient Suite")
}
