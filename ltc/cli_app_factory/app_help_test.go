package cli_app_factory_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/io"
	"github.com/codegangsta/cli"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("AppHelp", func() {
	var (
		fakeTargetVerifier *fake_target_verifier.FakeTargetVerifier
		memPersister       persister.Persister
		outputBuffer       *gbytes.Buffer
		terminalUI         terminal.UI
		cliApp             *cli.App
		cliConfig          *config.Config
		latticeVersion     string
	)

	BeforeEach(func() {
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		memPersister = persister.NewMemPersister()
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		cliConfig = config.New(memPersister)
		latticeVersion = "v0.2.Test"
	})

	JustBeforeEach(func() {
		cliApp = cli_app_factory.MakeCliApp(
			latticeVersion,
			"~/",
			&fake_exit_handler.FakeExitHandler{},
			cliConfig,
			lager.NewLogger("test"),
			fakeTargetVerifier,
			terminalUI,
		)
	})

	It("shows help for all commands", func() {

		dummyTemplate := `
{{range .Commands}}{{range .CommandSubGroups}}{{range .}}
{{.Name}}
{{end}}{{end}}{{end}}
`

		cliCommands := cliApp.Commands
		Expect(cliCommands).NotTo(BeEmpty())

		output := io.CaptureOutput(func() {
			cli_app_factory.ShowHelp(dummyTemplate, cliApp)
		})

		for _, command := range cliCommands {
			Expect(commandInOutput(command.Names(), output)).To(BeTrue(), command.Name+" not in help")
		}
	})

})

func commandInOutput(cmdName []string, output []string) bool {
	commandName := strings.Join(cmdName, ", ")
	for _, line := range output {
		if strings.TrimSpace(line) == strings.TrimSpace(commandName) {
			return true
		}
	}
	return false
}
