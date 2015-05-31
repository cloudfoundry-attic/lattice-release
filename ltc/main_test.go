package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("LatticeCli Main", func() {
	It("compiles and displays help text", func() {
		command := exec.Command(ltcPath)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("ltc - Command line interface for Lattice."))
	})
})
