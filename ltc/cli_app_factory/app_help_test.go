package cli_app_factory_test

import (
	"errors"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
)

var _ = Describe("AppHelp", func() {
	var (
		cliApp       *cli.App
		outputBuffer *gbytes.Buffer
	)

	dummyTemplate := `
		{{range .Commands}}{{range .CommandSubGroups}}{{range .}}
		{{.Name}}
		{{end}}{{end}}{{end}}`

	subCommandTemplate := `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.Name}} command{{if .Flags}} [command options]{{end}} [arguments...]
`
	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()

		cliApp = cli_app_factory.MakeCliApp(
			"",
			"",
			"~/",
			nil,
			config.New(persister.NewMemPersister()),
			nil,
			nil,
			terminal.NewUI(nil, outputBuffer, nil),
		)
	})

	Describe("ShowHelp", func() {

		It("shows help for all commands", func() {
			Expect(cliApp.Commands).NotTo(BeEmpty())

			cli_app_factory.ShowHelp(outputBuffer, dummyTemplate, cliApp)

			outputBytes, err := ioutil.ReadAll(outputBuffer)
			Expect(err).NotTo(HaveOccurred())
			for _, command := range cliApp.Commands {
				commandName := strings.TrimSpace(strings.Join(command.Names(), ", "))
				Expect(string(outputBytes)).To(ContainSubstring(commandName))
			}
		})

		It("shows help for a specific command", func() {
			subCommand := cli.Command{Name: "print-a-command"}

			cli_app_factory.ShowHelp(outputBuffer, subCommandTemplate, subCommand)

			outputBytes, err := ioutil.ReadAll(outputBuffer)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(outputBytes)).To(ContainSubstring(subCommand.Name))
		})

		It("panics for other type", func() {
			showHelp := func() func() {
				return func() { cli_app_factory.ShowHelp(outputBuffer, dummyTemplate, struct{}{}) }
			}

			Consistently(showHelp).Should(Panic(), "unsupported type but help didn't panic")
		})

		Context("when writer is busted", func() {
			It("panics showing help for all commands", func() {
				showHelp := func() func() {
					return func() { cli_app_factory.ShowHelp(errorWriter{}, dummyTemplate, cliApp) }
				}

				Consistently(showHelp).Should(Panic(), "writer bailed but help didn't panic")
			})

			It("panics showing help for a specific command", func() {
				showHelp := func() func() {
					return func() { cli_app_factory.ShowHelp(errorWriter{}, dummyTemplate, cli.Command{}) }
				}

				Consistently(showHelp).Should(Panic(), "writer bailed but help didn't panic")
			})
		})

	})
})

type errorWriter struct{}

func (errorWriter) Write(p []byte) (n int, err error) {
	return -1, errors.New("no bueno")
}
