package command_factory_test

import (
	_ "errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/diego-edge-cli/test_helpers"

	"github.com/pivotal-cf-experimental/diego-edge-cli/logs/command_factory"
)

var _ = Describe("CommandFactory", func() {
	Describe("logsCommand", func() {
		var (
			output *gbytes.Buffer
		)

		BeforeEach(func() {
			output = gbytes.NewBuffer()
		})

		It("Tails logs", func() {
			args := []string{
				"my-app-guid",
			}

			logReader := &fakeLogReader{}
			commandFactory := command_factory.NewLogsCommandFactory(logReader, output)
			tailLogsCommand := commandFactory.MakeLogsCommand()

			context := test_helpers.ContextFromArgsAndCommand(args, tailLogsCommand)

			logReader.addLog("First log")

			go tailLogsCommand.Action(context)

			Eventually(logReader.calledAppGuidsBeingTailed).Should(Equal([]string{"my-app-guid"}))
			Eventually(output).Should(gbytes.Say("First log\n"))
		})

		It("Handles invalid appguids", func() {
			args := []string{}

			logReader := &fakeLogReader{}
			commandFactory := command_factory.NewLogsCommandFactory(logReader, output)
			tailLogsCommand := commandFactory.MakeLogsCommand()

			context := test_helpers.ContextFromArgsAndCommand(args, tailLogsCommand)

			logReader.addLog("First log \n")

			tailLogsCommand.Action(context)

			Expect(output).To(gbytes.Say("Invalid Usage\n"))
		})

	})
})

type fakeLogReader struct {
	calledAppGuids []string
	logs           []string
}

func (f *fakeLogReader) TailLogs(appGuid string, callback func(string)) {
	f.calledAppGuids = append(f.calledAppGuids, appGuid)

	for _, log := range f.logs {
		callback(log)
	}
}

func (f *fakeLogReader) addLog(log string) {
	f.logs = append(f.logs, log)
}

func (f *fakeLogReader) calledAppGuidsBeingTailed() []string {
	return f.calledAppGuids
}
