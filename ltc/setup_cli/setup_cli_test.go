package setup_cli_test

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/setup_cli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/codegangsta/cli"
)

var _ = Describe("SetupCli", func() {

	var (
		cliApp *cli.App
	)

	Describe("NewCliApp", func() {
		It("Runs registered command without error", func() {
			cliApp = setup_cli.NewCliApp()

			Expect(cliApp).NotTo(BeNil())
		})
	})
})
