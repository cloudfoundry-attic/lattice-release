package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc"
	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("HelpHelpers", func() {
	var (
		fakeTargetVerifier *fake_target_verifier.FakeTargetVerifier
		outputBuffer       *gbytes.Buffer
		terminalUI         terminal.UI
		cliConfig          *config.Config
		cliApp             *cli.App
	)

	BeforeEach(func() {
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		cliConfig = config.New(persister.NewMemPersister())
		cliApp = cli_app_factory.MakeCliApp(
			"",
			"~/",
			nil,
			cliConfig,
			nil,
			fakeTargetVerifier,
			terminalUI,
		)
	})

	Describe("MatchArgAndFlags", func() {
		It("Checks for badflag", func() {
			cliAppArgs := []string{"ltc", "create", "--badflag"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])

			Expect(badFlags).To(Equal("Unknown flag \"--badflag\""))
		})

		It("returns if multiple bad flags are passed", func() {
			cliAppArgs := []string{"ltc", "create", "--badflag1", "--badflag2"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)

			Expect(badFlags).To(Equal("Unknown flags: \"--badflag1\", \"--badflag2\""))
		})
	})

	Describe("GetCommandFlags", func() {
		It("returns flags string slice for a given command list of type Flag", func() {
			flagList := main.GetCommandFlags(cliApp, "create")
			cmd := cliApp.Command("create")
			for _, flag := range cmd.Flags {
				switch t := flag.(type) {
				default:
					Fail("unhandled flag type")
				case cli.StringSliceFlag:
					Expect(flagList).Should(ContainElement(t.Name))
				case cli.IntFlag:
					Expect(flagList).Should(ContainElement(t.Name))
				case cli.StringFlag:
					Expect(flagList).Should(ContainElement(t.Name))
				case cli.BoolFlag:
					Expect(flagList).Should(ContainElement(t.Name))
				case cli.DurationFlag:
					Expect(flagList).Should(ContainElement(t.Name))
				}
			}
		})

		It("does not append unexpected flag types", func() {
			cliApp = cli.NewApp()
			cliGenericFlag := cli.GenericFlag{Name: "capture-the-flag"}
			cliApp.Commands = []cli.Command{
				cli.Command{
					Name: "commander-keen",
					Flags: []cli.Flag{
						cliGenericFlag,
					},
				},
			}

			flagList := main.GetCommandFlags(cliApp, "commander-keen")
			Expect(flagList).ToNot(ContainElement(cliGenericFlag.Name))
		})
	})

	Describe("GetByCmdName", func() {
		It("returns command not found error", func() {
			_, err := main.GetByCmdName(cliApp, "zz")

			Expect(err).To(MatchError("Command not found"))
		})
	})

	Describe("RequestHelp", func() {
		It("checks for the flag -h", func() {
			cliAppArgs := []string{"ltc", "-h"}
			boolVal := main.RequestHelp(cliAppArgs[1:])

			Expect(boolVal).To(BeTrue())
		})

		It("checks for the flag --help", func() {
			cliAppArgs := []string{"ltc", "--help"}
			boolVal := main.RequestHelp(cliAppArgs[1:])

			Expect(boolVal).To(BeTrue())
		})

		It("checks for the unknown flag", func() {
			cliAppArgs := []string{"ltc", "--unknownFlag"}
			boolVal := main.RequestHelp(cliAppArgs[1:])

			Expect(boolVal).To(BeFalse())
		})

	})

	Describe("Flag verification", func() {

		BeforeEach(func() {
			cliApp.Commands = []cli.Command{
				cli.Command{Name: "print-a-unicorn",
					Action: func(ctx *cli.Context) {},
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

			fakeTargetVerifier.VerifyTargetReturns(true, true, nil)
			cliConfig.SetTarget("my-lattice.example.com")
			cliConfig.Save()
		})

		It("informs user for any incorrect provided flags", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "--bad-flag=10"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("Unknown flag \"--bad-flag\""))
		})

		It("checks flags with prefix '--'", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "not-a-flag", "--invalid-flag"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("Unknown flag \"--invalid-flag\""))
			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"not-a-flag\""))
		})

		It("checks flags with prefix '-'", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "not-a-flag", "-invalid-flag"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"-invalid-flag\""))
			Expect(outputBuffer).NotTo(test_helpers.Say("\"not-a-flag\""))
		})

		It("checks flags but ignores the value after '=' ", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "-f1=1", "-invalid-flag=blarg"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"-invalid-flag\""))
			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"-p\""))
		})

		It("outputs all unknown flags in single sentence", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "--bad-flag1", "--bad-flag2", "--bad-flag3"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"--bad-flag1\", \"--bad-flag2\", \"--bad-flag3\""))
		})

		It("only checks input flags against flags from the provided command", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "--instances", "--skip-ssl-validation"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).To(HaveOccurred())

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage."))
			Expect(outputBuffer).To(test_helpers.Say("\"--skip-ssl-validation\""))
		})

		It("accepts -h and --h flags for all commands", func() {
			cliAppArgs := []string{"ltc", "print-a-unicorn", "-h"}
			flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err := cliApp.Run(cliAppArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"-h\""))

			cliAppArgs = []string{"ltc", "print-a-unicorn", "--h"}
			flags = main.GetCommandFlags(cliApp, cliAppArgs[1])
			badFlags = main.MatchArgAndFlags(flags, cliAppArgs[2:])
			main.InjectHelpTemplate(badFlags)
			err = cliApp.Run(cliAppArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(outputBuffer).NotTo(test_helpers.Say("Unknown flag \"--h\""))
		})

		Context("When a negative integer is preceeded by a valid flag", func() {
			It("skips validation for negative integer flag values", func() {
				cliAppArgs := []string{"ltc", "print-a-unicorn", "-f1", "-10"}
				flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
				badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
				main.InjectHelpTemplate(badFlags)

				err := cliApp.Run(cliAppArgs)

				Expect(err).ToNot(HaveOccurred())
				Expect(outputBuffer).NotTo(test_helpers.Say("\"-10\""))
			})
		})

		Context("When a negative integer is preceeded by a invalid flag", func() {
			It("validates the negative integer as a flag", func() {
				cliAppArgs := []string{"ltc", "print-a-unicorn", "-badflag", "-10"}
				flags := main.GetCommandFlags(cliApp, cliAppArgs[1])
				badFlags := main.MatchArgAndFlags(flags, cliAppArgs[2:])
				main.InjectHelpTemplate(badFlags)
				err := cliApp.Run(cliAppArgs)

				Expect(err).To(HaveOccurred())
				Expect(outputBuffer).To(test_helpers.Say("\"-badflag\""))
				Expect(outputBuffer).To(test_helpers.Say("\"-10\""))
			})
		})

	})
})
