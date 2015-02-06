package console_tailed_logs_outputter_test

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/noaa/events"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/logs/console_tailed_logs_outputter"
	"github.com/pivotal-cf-experimental/lattice-cli/logs/fake_log_reader"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"
)

var _ = Describe("ConsoleTailedLogsOutputter", func() {
	var (
		outputBuffer *gbytes.Buffer
	)

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
	})
	Describe("OutputTailedLogs", func() {
		It("Tails logs", func() {
			logReader := fake_log_reader.NewFakeLogReader()
			consoleTailedLogsOutputter := console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(output.New(outputBuffer), logReader)

			time := time.Now()
			sourceType := "RTR"
			sourceInstance := "1"

			unixTime := time.UnixNano()
			logReader.AddLog(&events.LogMessage{
				Message:        []byte("First log"),
				Timestamp:      &unixTime,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			})
			logReader.AddError(errors.New("First Error"))

			go consoleTailedLogsOutputter.OutputTailedLogs("my-app-guid")

			Eventually(logReader.GetAppGuid).Should(Equal("my-app-guid"))

			logOutputBufferString := fmt.Sprintf("%s [%s|%s] First log\n", colors.Cyan(time.Format("02 Jan 15:04")), colors.Yellow(sourceType), colors.Yellow(sourceInstance))
			Eventually(outputBuffer).Should(test_helpers.Say(logOutputBufferString))

			Eventually(outputBuffer).Should(test_helpers.Say("First Error\n"))
		})
	})

	Describe("StopOutputting", func() {
		It("stops outputting logs", func() {
			logReader := fake_log_reader.NewFakeLogReader()
			consoleTailedLogsOutputter := console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(output.New(outputBuffer), logReader)

			go consoleTailedLogsOutputter.OutputTailedLogs("my-app-guid")

			consoleTailedLogsOutputter.StopOutputting()

			Expect(logReader.IsLogTailStopped()).To(BeTrue())
		})
	})
})
