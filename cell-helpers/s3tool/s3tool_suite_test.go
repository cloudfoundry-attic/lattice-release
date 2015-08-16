package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestS3Tool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3Tool Suite")
}

var s3toolPath string

var _ = BeforeSuite(func() {
	var err error
	s3toolPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/s3tool")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
