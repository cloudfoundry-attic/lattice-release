package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	cli string
)

var _ = BeforeSuite(func() {
	var err error
	cli, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/ltc")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("lattice-cli", func() {
	It("compiles and displays help text", func() {
		command := exec.Command(cli)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("ltc - Command line interface for Lattice."))
	})
})
