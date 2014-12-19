package receptor_client_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestReceptorBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReceptorBuilder Suite")
}
