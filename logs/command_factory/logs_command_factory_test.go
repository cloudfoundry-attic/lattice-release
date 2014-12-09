package command_factory_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry/noaa/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/diego-edge-cli/colors"
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

			appGuidChan := make(chan string)
			logReader := &fakeLogReader{appGuidChan: appGuidChan}
			commandFactory := command_factory.NewLogsCommandFactory(logReader, output)
			tailLogsCommand := commandFactory.MakeLogsCommand()

			time := time.Now()
			sourceType := "RTR"
			sourceInstance := "1"

			unixTime := time.UnixNano()
			logReader.addLog(&events.LogMessage{
				Message:        []byte("First log"),
				Timestamp:      &unixTime,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			})
			logReader.addError(errors.New("First Error"))

			go test_helpers.ExecuteCommandWithArgs(tailLogsCommand, args)

			Eventually(appGuidChan).Should(Receive(Equal("my-app-guid")))

			logOutputString := fmt.Sprintf("%s [%s|%s] First log\n", colors.Cyan(time.Format("02 Jan 15:04")), colors.Yellow(sourceType), colors.Yellow(sourceInstance))
			Eventually(string(output.Contents())).Should(ContainSubstring(logOutputString))

			Eventually(output).Should(gbytes.Say("First Error\n"))
		})

		It("Handles invalid appguids", func() {
			args := []string{}

			logReader := &fakeLogReader{}
			commandFactory := command_factory.NewLogsCommandFactory(logReader, output)
			tailLogsCommand := commandFactory.MakeLogsCommand()

			err := test_helpers.ExecuteCommandWithArgs(tailLogsCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(gbytes.Say("Incorrect Usage"))
		})

	})
})

type fakeLogReader struct {
	appGuidChan chan string
	logs        []*events.LogMessage
	errors      []error
}

func (f *fakeLogReader) TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error)) {
	for _, log := range f.logs {
		logCallback(log)
	}

	for _, err := range f.errors {
		errorCallback(err)
	}

	f.appGuidChan <- appGuid
}

func (f *fakeLogReader) addLog(log *events.LogMessage) {
	f.logs = append(f.logs, log)
}

func (f *fakeLogReader) addError(err error) {
	f.errors = append(f.errors, err)
}
