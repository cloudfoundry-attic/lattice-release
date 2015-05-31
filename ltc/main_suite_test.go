package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var ltcPath string

func TestLatticeCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LatticeCli Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	cliPath, err := gexec.Build("github.com/cloudfoundry-incubator/lattice/ltc")
	Expect(err).NotTo(HaveOccurred())
	return []byte(cliPath)
}, func(cliPath []byte) {
	ltcPath = string(cliPath)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
