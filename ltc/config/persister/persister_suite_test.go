package persister_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPersister(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Persister Suite")
}
