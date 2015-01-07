package command_factory_test

import (
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/command_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/app_examiner/fake_app_examiner"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"
)

var _ = Describe("CommandFactory", func() {

	var (
		appExaminer  *fake_app_examiner.FakeAppExaminer
		outputBuffer *gbytes.Buffer
	)

	BeforeEach(func() {
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
	})

	Describe("ListAppsCommand", func() {
		var listAppsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppExaminerCommandFactory(appExaminer, output.New(outputBuffer))
			listAppsCommand = commandFactory.MakeListAppCommand()
		})

		It("displays all the existing apps", func() {
			listApps := []app_examiner.AppInfo{
				app_examiner.AppInfo{ProcessGuid: "process1", DesiredInstances: 21, ActualRunningInstances: 0, DiskMB: 100, MemoryMB: 50, Routes: []string{"alldaylong.com"}},
				app_examiner.AppInfo{ProcessGuid: "process2", DesiredInstances: 8, ActualRunningInstances: 9, DiskMB: 400, MemoryMB: 30, Routes: []string{"never.io"}},
				app_examiner.AppInfo{ProcessGuid: "process3", DesiredInstances: 5, ActualRunningInstances: 5, DiskMB: 600, MemoryMB: 90, Routes: []string{"allthetime.com", "herewego.org"}},
				app_examiner.AppInfo{ProcessGuid: "process4", DesiredInstances: 0, ActualRunningInstances: 0, DiskMB: 10, MemoryMB: 10, Routes: []string{}},
			}
			appExaminer.ListAppsReturns(listApps, nil)

			expectedOutput := gbytes.NewBuffer()
			w := new(tabwriter.Writer)
			w.Init(expectedOutput, 10+len(colors.NoColor("")), 8, 1, '\t', 0)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMb"), colors.Bold("MemoryMB"), colors.Bold("Routes"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process1"), colors.Red("0/21"), colors.NoColor("100"), colors.NoColor("50"), colors.Cyan("alldaylong.com"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process2"), colors.Yellow("9/8"), colors.NoColor("400"), colors.NoColor("30"), colors.Cyan("never.io"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process3"), colors.Green("5/5"), colors.NoColor("600"), colors.NoColor("90"), colors.Cyan("allthetime.com herewego.org"))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold("process4"), colors.Green("0/0"), colors.NoColor("10"), colors.NoColor("10"), colors.Cyan(""))
			w.Flush()

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer.Contents()).To(Equal(expectedOutput.Contents()))
		})

		It("alerts the user if there are no apps", func() {
			listApps := []app_examiner.AppInfo{}
			appExaminer.ListAppsReturns(listApps, nil)

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("No apps to display."))
		})

		It("alerts the user if fetching the list returns an error", func() {
			listApps := []app_examiner.AppInfo{}
			appExaminer.ListAppsReturns(listApps, errors.New("The list was lost"))

			test_helpers.ExecuteCommandWithArgs(listAppsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Error listing apps: The list was lost"))
		})
	})
})
