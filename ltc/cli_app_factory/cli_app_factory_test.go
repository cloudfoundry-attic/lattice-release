package cli_app_factory_test

import (
	"errors"
	"flag"

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
)

var _ = Describe("CliAppFactory", func() {

	var (
		fakeTargetVerifier           *fake_target_verifier.FakeTargetVerifier
		fakeExitHandler              *fake_exit_handler.FakeExitHandler
		outputBuffer                 *gbytes.Buffer
		terminalUI                   terminal.UI
		cliApp                       *cli.App
		cliConfig                    *config.Config
		latticeVersion, diegoVersion string
	)

	BeforeEach(func() {
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		fakeExitHandler = new(fake_exit_handler.FakeExitHandler)
		memPersister := persister.NewMemPersister()
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		cliConfig = config.New(memPersister)
		latticeVersion, diegoVersion = "v0.2.Test", "0.12345.0"
	})

	JustBeforeEach(func() {
		cliApp = cli_app_factory.MakeCliApp(
			diegoVersion,
			latticeVersion,
			"~/",
			fakeExitHandler,
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
			Expect(cliApp.Version).To(Equal("v0.2.Test (diego 0.12345.0)"))
			Expect(cliApp.Email).To(Equal("cf-lattice@lists.cloudfoundry.org"))
			Expect(cliApp.Usage).To(Equal(cli_app_factory.LtcUsage))
			Expect(cliApp.Commands).NotTo(BeEmpty())
			Expect(cliApp.Action).ToNot(BeNil())
			Expect(cliApp.CommandNotFound).ToNot(BeNil())

			By("writing to the App.Writer", func() {
				cliApp.Writer.Write([]byte("write_sample"))
				Expect(outputBuffer).To(test_helpers.Say("write_sample"))
			})

		})

		Context("when invoked without latticeVersion set", func() {
			BeforeEach(func() {
				diegoVersion = ""
				latticeVersion = ""
			})

			It("defaults the version", func() {
				Expect(cliApp).NotTo(BeNil())
				Expect(cliApp.Version).To(Equal("development (not versioned) (diego unknown)"))
			})
		})

		Describe("App.Action", func() {
			Context("when ltc is run without argument(s)", func() {
				It("prints app help", func() {
					cli.AppHelpTemplate = "HELP_TEMPLATE"
					flagSet := flag.NewFlagSet("flag_set", flag.ContinueOnError)
					flagSet.Parse([]string{})
					testContext := cli.NewContext(cliApp, flagSet, nil)

					cliApp.Action(testContext)

					Expect(outputBuffer).To(test_helpers.Say("ltc - Command line interface for Lattice."))
				})
			})

			Context("when ltc is run with argument(s)", func() {
				It("prints unknown command message", func() {
					flagSet := flag.NewFlagSet("flag_set", flag.ContinueOnError)
					flagSet.Parse([]string{"one_arg"})
					testContext := cli.NewContext(cliApp, flagSet, nil)

					cliApp.Action(testContext)

					Expect(outputBuffer).To(test_helpers.Say("ltc: 'one_arg' is not a registered command. See 'ltc help'"))
				})
			})
		})

		Describe("App.CommandNotFound", func() {
			It("prints unknown command message and exits nonzero", func() {
				testContext := cli.NewContext(cliApp, &flag.FlagSet{}, nil)

				cliApp.CommandNotFound(testContext, "do_it")

				Expect(outputBuffer).To(test_helpers.Say("ltc: 'do_it' is not a registered command. See 'ltc help'"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{1}))
			})
		})

		Describe("App.Before", func() {
			Context("when running the target command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					Expect(cliConfig.Save()).To(Succeed())

					commandRan := false

					cliApp.Commands = []cli.Command{
						cli.Command{
							Name: "target",
							Action: func(ctx *cli.Context) {
								commandRan = true
							},
						},
					}

					Expect(cliApp.Run([]string{"ltc", "target"})).To(Succeed())
					Expect(commandRan).To(BeTrue())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
				})
			})

			Context("when running the help command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					Expect(cliConfig.Save()).To(Succeed())

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
					Expect(err).NotTo(HaveOccurred())

					Expect(commandRan).To(BeTrue())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(BeZero())
				})
			})

			Context("when running the bare ltc command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					Expect(cliConfig.Save()).To(Succeed())

					commandRan := false
					cliApp.Action = func(context *cli.Context) {
						commandRan = true
					}

					cliAppArgs := []string{"ltc"}

					err := cliApp.Run(cliAppArgs)
					Expect(err).NotTo(HaveOccurred())

					Expect(commandRan).To(BeTrue())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
				})
			})

			Context("when we cannot find the subcommand", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					Expect(cliConfig.Save()).To(Succeed())

					commandRan := false
					cliApp.Action = func(context *cli.Context) {
						commandRan = true
					}

					cliAppArgs := []string{"ltc", "buy-me-a-pony"}

					err := cliApp.Run(cliAppArgs)
					Expect(err).NotTo(HaveOccurred())

					Expect(commandRan).To(BeTrue())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
				})
			})

			Context("Any other command", func() {
				Context("when targeted receptor is up and we are authorized", func() {
					It("executes the command", func() {
						fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

						cliConfig.SetTarget("my-lattice.example.com")
						Expect(cliConfig.Save()).To(Succeed())

						commandRan := false
						cliApp.Commands = []cli.Command{
							cli.Command{
								Name: "print-a-unicorn",
								Action: func(ctx *cli.Context) {
									commandRan = true
								},
							},
						}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)
						Expect(err).NotTo(HaveOccurred())

						Expect(commandRan).To(BeTrue())
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-lattice.example.com"))
					})
				})

				Context("when we are unauthorized for the targeted receptor", func() {
					It("Prints an error message and does not execute the command", func() {
						fakeTargetVerifier.VerifyTargetReturns(true, false, nil)
						cliConfig.SetTarget("my-borked-lattice.example.com")
						Expect(cliConfig.Save()).To(Succeed())

						commandRan := false
						cliApp.Commands = []cli.Command{
							cli.Command{
								Name:   "print-a-unicorn",
								Action: func(ctx *cli.Context) { commandRan = true },
							},
						}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)
						Expect(err).To(MatchError("Could not authenticate with the receptor."))

						Expect(outputBuffer).To(test_helpers.Say("Could not authenticate with the receptor. Please run ltc target with the correct credentials."))
						Expect(commandRan).To(BeFalse())

						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-borked-lattice.example.com"))
					})
				})

				Context("when the receptor is down", func() {
					It("prints a helpful error", func() {
						fakeTargetVerifier.VerifyTargetReturns(false, false, errors.New("oopsie!"))

						cliConfig.SetTarget("my-borked-lattice.example.com")
						Expect(cliConfig.Save()).To(Succeed())

						commandRan := false
						cliApp.Commands = []cli.Command{
							cli.Command{
								Name:   "print-a-unicorn",
								Action: func(ctx *cli.Context) { commandRan = true },
							},
						}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)
						Expect(err).To(MatchError("oopsie!"))

						Expect(outputBuffer).To(test_helpers.Say("Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.\n\tUnderlying error: oopsie!"))
						Expect(commandRan).To(BeFalse())

						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-borked-lattice.example.com"))
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
