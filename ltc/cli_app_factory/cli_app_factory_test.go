package cli_app_factory_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/lager"

	config_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/config/command_factory"
)

var _ = Describe("CliAppFactory", func() {

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

	Describe("MakeCliApp", func() {
		It("makes an app", func() {
			Expect(cliApp).ToNot(BeNil())
			Expect(cliApp.Name).To(Equal("ltc"))
			Expect(cliApp.Author).To(Equal("Pivotal"))
			Expect(cliApp.Version).To(Equal("v0.2.Test"))
			Expect(cliApp.Email).To(Equal("cf-lattice@lists.cloudfoundry.org"))
			Expect(cliApp.Usage).To(Equal(cli_app_factory.LtcUsage))
			Expect(cliApp.Commands).NotTo(BeEmpty())
		})

		It("runs the help command and checks a sample string in each line", func() {
			cliApp.Run([]string{"ltc"})

			//Ignore bolds and formatting lets check a sample content from each string
			helpText := []string{
				"NAME:",
				"ltc - Command line interface for Lattice.",
				"USAGE:",
				"global options",
				"VERSION:",
				"v0.2.Test",
				"AUTHOR",
				"Pivotal <lattice@cloudfoundry.org>",
				"",
				"GLOBAL OPTIONS:",
				"show help",
				"print the version",
				"",
				"COMMANDS",
				"Target Lattice",
				"Targets a lattice cluster",
				"",
				"Create and Modify Apps",
				"Creates a docker app on lattice",
				"Updates the routes for a running app",
				"Stops and removes docker app",
				"Scales a docker app on lattice",
				"",
				"Stream Logs",
				"Streams logs from the specified application",
				"",
				"See Whats Running",
				"Lists applications running on lattice",
				"Shows details about a running app on lattice",
				"Shows a visualization of the workload distribution across the lattice cells",
				"",
				"Advanced",
				"Creates a docker app from JSON on lattice",
				"",
				"Help and Debug",
				"Streams logs from the lattice cluster components",
				"Runs test suite against targeted lattice cluster",
			}

			for _, line := range helpText {
				Expect(outputBuffer).To(gbytes.Say(line))
			}

		})

		It("lists the subcommands", func() {
			cliCommands := cliApp.Commands
			cliExpectedCommands := []string{"target", "create", "update-routes", "remove", "scale", "logs", "list", "status", "visualize", "create-lrp", "debug-logs", "test"}
			Expect(cliCommands).NotTo(BeEmpty())

			var commandNames []string
			for _, cmd := range cliCommands {
				commandNames = append(commandNames, cmd.Name)
			}
			Expect(len(commandNames)).Should(Equal(len(cliExpectedCommands))) //Check the lenght first as we will know if a command is added or removed
			for i, cmd := range commandNames {
				Expect(cliExpectedCommands[i]).Should(Equal(cmd))
			}
		})

		Context("when invoked without latticeVersion set", func() {
			BeforeEach(func() {
				latticeVersion = ""
			})

			It("defaults the version", func() {
				Expect(cliApp).ToNot(BeNil())
				Expect(cliApp.Version).To(Equal("development (not versioned)"))
			})
		})

		Describe("App's Before Action", func() {
			Context("when running the target command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false

					cliApp.Commands = []cli.Command{
						cli.Command{
							Name: config_command_factory.TargetCommandName,
							Action: func(ctx *cli.Context) {
								commandRan = true
							},
						},
					}

					cliAppArgs := []string{"ltc", config_command_factory.TargetCommandName}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(BeZero())
					Expect(commandRan).To(BeTrue())
				})
			})

			Context("when running the help command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false

					cliApp.Commands = []cli.Command{
						cli.Command{
							Name: "help",
							Action: func(ctx *cli.Context) {
								commandRan = true
							},
						},
					}

					cliAppArgs := []string{"ltc", "help"}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(BeZero())
					Expect(commandRan).To(BeTrue())
				})
			})

			Context("when running the bare ltc command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false
					cliApp.Action = func(context *cli.Context) {
						commandRan = true
					}

					cliAppArgs := []string{"ltc"}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(BeZero())
					Expect(commandRan).To(BeTrue())
				})
			})

			Context("when we cannot find the subcommand", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false
					cliApp.Action = func(context *cli.Context) {
						commandRan = true
					}

					cliAppArgs := []string{"ltc", "buy-me-a-pony"}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(BeZero())
					Expect(commandRan).To(BeTrue())
				})
			})

			Context("Any other command", func() {
				Context("when targeted receptor is up and we are authorized", func() {
					It("executes the command", func() {
						fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

						cliConfig.SetTarget("my-lattice.example.com")
						cliConfig.Save()

						commandRan := false
						cliApp.Commands = []cli.Command{cli.Command{Name: "print-a-unicorn", Action: func(ctx *cli.Context) {
							commandRan = true
						}}}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)

						Expect(err).ToNot(HaveOccurred())
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-lattice.example.com"))
						Expect(commandRan).To(BeTrue())
					})
				})

				Context("when we are unauthorized for the targeted receptor", func() {
					It("Prints an error message and does not execute the command", func() {
						fakeTargetVerifier.VerifyTargetReturns(true, false, nil)
						cliConfig.SetTarget("my-borked-lattice.example.com")
						cliConfig.Save()

						commandRan := false
						cliApp.Commands = []cli.Command{cli.Command{Name: "print-a-unicorn", Action: func(ctx *cli.Context) { commandRan = true }}}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)

						Expect(err).To(HaveOccurred())
						Expect(outputBuffer).To(test_helpers.Say("Could not authenticate with the receptor. Please run ltc target with the correct credentials."))
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-borked-lattice.example.com"))
						Expect(commandRan).To(BeFalse())
					})
				})

				Context("when the receptor is down", func() {
					It("prints a helpful error", func() {
						fakeTargetVerifier.VerifyTargetReturns(false, false, errors.New("oopsie!"))

						cliConfig.SetTarget("my-borked-lattice.example.com")
						cliConfig.Save()

						commandRan := false
						cliApp.Commands = []cli.Command{cli.Command{Name: "print-a-unicorn", Action: func(ctx *cli.Context) { commandRan = true }}}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)

						Expect(err).To(HaveOccurred())
						Expect(outputBuffer).To(test_helpers.Say("Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.\n\tUnderlying error: oopsie!"))
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-borked-lattice.example.com"))
						Expect(commandRan).To(BeFalse())
					})
				})
			})
		})

	})

	Describe("LoggregatorUrl", func() {
		It("returns loggregator url with the websocket scheme added", func() {
			loggregatorUrl := cli_app_factory.LoggregatorUrl("doppler.diego.io")
			Expect(loggregatorUrl).To(Equal("ws://doppler.diego.io"))
		})
	})

})
