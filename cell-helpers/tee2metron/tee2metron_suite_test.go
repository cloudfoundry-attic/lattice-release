package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestTee2metron(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tee2metron Suite")
}

var tee2MetronPath, chattyProcessPath string

var _ = BeforeSuite(func() {
	var err error
	tee2MetronPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/tee2metron")
	chattyProcessPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/tee2metron/test_helpers/chatty_process")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
