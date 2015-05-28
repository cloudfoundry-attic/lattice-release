package setup_cli_test

import (
	. "github.com/cloudfoundry-incubator/lattice/ltc/setup_cli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("HelpHelpers", func() {
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
		cliApp = cli_app_factory.MakeCliApp(
			latticeVersion,
			"~/",
			&fake_exit_handler.FakeExitHandler{},
			cliConfig,
			lager.NewLogger("test"),
			fakeTargetVerifier,
			terminalUI,
		)
		cliApp.Writer = terminalUI
	})

	Describe("MatchArgAndFlags", func() {
		It("Checks for badflag", func() {
			cliAppArgs := []string{"ltc", "create", "--badflag"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])

			Expect(badFlags).To(Equal("Unknown flag \"--badflag\""))

		})

		It("returns if multiple bad flags are passed", func() {
			cliAppArgs := []string{"ltc", "create", "--badflag1", "--badflag2"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			Expect(badFlags).To(Equal("Unknown flags: \"--badflag1\", \"--badflag2\""))

		})
	})

	Describe("GetCommandFlags", func() {
		It("returns list of type Flag", func() {
			flaglist := GetCommandFlags(cliApp, "create")
			cmd := cliApp.Command("create")
			for _, flag := range cmd.Flags {
				switch t := flag.(type) {
				default:
				case cli.StringSliceFlag:
					Expect(flaglist).Should(ContainElement(t.Name))
				case cli.IntFlag:
					Expect(flaglist).Should(ContainElement(t.Name))
				case cli.StringFlag:
					Expect(flaglist).Should(ContainElement(t.Name))
				case cli.BoolFlag:
					Expect(flaglist).Should(ContainElement(t.Name))
				case cli.DurationFlag:
					Expect(flaglist).Should(ContainElement(t.Name))
				}
			}
		})
	})

	Describe("GetByCmdName", func() {
		It("returns command not found error", func() {
			_, err := GetByCmdName(cliApp, "zz")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Command not found"))

		})
	})

	Describe("RequestHelp", func() {
		It("checks for the flag -h", func() {
			cliAppArgs := []string{"ltc", "-h"}
			boolVal := RequestHelp(cliAppArgs[1:])
			Expect(boolVal).To(BeTrue())
		})

		It("checks for the flag --help", func() {
			cliAppArgs := []string{"ltc", "--help"}
			boolVal := RequestHelp(cliAppArgs[1:])
			Expect(boolVal).To(BeTrue())
		})

		It("checks for the unknown flag", func() {
			cliAppArgs := []string{"ltc", "--unknownFlag"}
			boolVal := RequestHelp(cliAppArgs[1:])
			Expect(boolVal).To(BeFalse())
		})

	})

	Describe("Flag verification", func() {

		BeforeEach(func() {
			commandRan := false
			cliApp.Commands = []cli.Command{
				cli.Command{Name: "print-a-unicorn",
					Action: func(ctx *cli.Context) { commandRan = true },
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "flag1, f1",
							Usage: "flag for print-a-unicorn command",
							Value: 10,
						},
						cli.BoolFlag{
							Name:  "flag2, f2",
							Usage: "flag for print-a-unicorn command",
						},
					},
				},
			}

		})

		It("informs user for any incorrect provided flags", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "--bad-flag=10"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("Unknown flag \"--bad-flag\""))

		})

		It("checks flags with prefix '--'", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "not-a-flag", "--invalid-flag"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("Unknown flag \"--invalid-flag\""))
			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"not-a-flag\""))

		})

		It("checks flags with prefix '-'", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "not-a-flag", "-invalid-flag"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"-invalid-flag\""))
			Expect(outputBuffer).NotTo(test_helpers.Say("\"not-a-flag\""))
		})

		It("checks flags but ignores the value after '=' ", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "-f1=1", "-invalid-flag=blarg"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"-invalid-flag\""))
			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"-p\""))
		})

		It("outputs all unknown flags in single sentence", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "--bad-flag1", "--bad-flag2", "--bad-flag3"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"--bad-flag1\", \"--bad-flag2\", \"--bad-flag3\""))
		})

		It("only checks input flags against flags from the provided command", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "--instances", "--skip-ssl-validation"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"--skip-ssl-validation\""))
		})

		It("accepts -h and --h flags for all commands", func() {
			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()

			cliAppArgs := []string{"ltc", "print-a-unicorn", "-h"}
			flags := GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"-h\""))

			cliAppArgs = []string{"ltc", "print-a-unicorn", "--h"}
			flags = GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags = MatchArgAndFlags(flags, cliAppArgs[2:])
			InjectHelpTemplate(badFlags)
			err = cliApp.Run(cliAppArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"--h\""))
		})

		Context("When a negative integer is preceeded by a valid flag", func() {
			It("skips validation for negative integer flag values", func() {
				fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
				cliConfig.SetTarget("my-lattice.example.com")
				cliConfig.Save()

				cliAppArgs := []string{"ltc", "print-a-unicorn", "-f1", "-10"}
				flags := GetCommandFlags(cliApp, cliAppArgs[1])
				badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
				InjectHelpTemplate(badFlags)
				err := cliApp.Run(cliAppArgs)

				Expect(err).ToNot(HaveOccurred())
				Expect(outputBuffer).NotTo(test_helpers.Say("\"-10\""))
			})
		})

		Context("When a negative integer is preceeded by a invalid flag", func() {
			It("validates the negative integer as a flag", func() {
				fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

				cliConfig.SetTarget("my-lattice.example.com")
				cliConfig.Save()

				cliAppArgs := []string{"ltc", "print-a-unicorn", "-badflag", "-10"}
				flags := GetCommandFlags(cliApp, cliAppArgs[1])
				badFlags := MatchArgAndFlags(flags, cliAppArgs[2:])
				InjectHelpTemplate(badFlags)
				err := cliApp.Run(cliAppArgs)

				Expect(err).To(HaveOccurred())
				Expect(outputBuffer).To(test_helpers.Say("\"-badflag\""))
				Expect(outputBuffer).To(test_helpers.Say("\"-10\""))
			})
		})
	})
})
