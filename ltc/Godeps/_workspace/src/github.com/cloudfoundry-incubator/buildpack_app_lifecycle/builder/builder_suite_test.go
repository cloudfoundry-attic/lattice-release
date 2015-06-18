package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var builderPath string

func TestBuildpackLifecycleBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack-Lifecycle-Builder Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	builder, err := gexec.Build("github.com/cloudfoundry-incubator/buildpack_app_lifecycle/builder")
	Expect(err).NotTo(HaveOccurred())
	return []byte(builder)
}, func(builder []byte) {
	builderPath = string(builder)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
