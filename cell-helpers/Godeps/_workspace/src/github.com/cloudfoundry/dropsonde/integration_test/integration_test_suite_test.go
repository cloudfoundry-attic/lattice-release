package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"log"
	"testing"
)

func TestIntegrationTest(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTest Suite")
}
