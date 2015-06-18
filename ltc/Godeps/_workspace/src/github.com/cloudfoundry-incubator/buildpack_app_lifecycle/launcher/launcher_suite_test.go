package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var launcher string

func TestBuildpackLifecycleLauncher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack-Lifecycle-Launcher Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	launcherPath, err := gexec.Build("github.com/cloudfoundry-incubator/buildpack_app_lifecycle/launcher", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(launcherPath)
}, func(launcherPath []byte) {
	launcher = string(launcherPath)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
