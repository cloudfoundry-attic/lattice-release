package console_tailed_logs_outputter_test

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/fake_log_reader"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/cloudfoundry/sonde-go/events"
)

var _ = Describe("ConsoleTailedLogsOutputter", func() {
	var (
		outputBuffer               *gbytes.Buffer
		terminalUI                 terminal.UI
		logReader                  *fake_log_reader.FakeLogReader
		consoleTailedLogsOutputter *console_tailed_logs_outputter.ConsoleTailedLogsOutputter
	)

	buildLogMessage := func(sourceType, sourceInstance string, timestamp time.Time, message []byte) *events.LogMessage {
		unixTime := timestamp.UnixNano()
		return &events.LogMessage{
			Message:        message,
			Timestamp:      &unixTime,
			SourceType:     &sourceType,
			SourceInstance: &sourceInstance,
		}
	}

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		logReader = fake_log_reader.NewFakeLogReader()
		consoleTailedLogsOutputter = console_tailed_logs_outputter.NewConsoleTailedLogsOutputter(terminalUI, logReader)
	})

	Describe("OutputTailedLogs", func() {
		It("Tails logs", func() {
			now := time.Now()
			logReader.AddLog(buildLogMessage("RTR", "1", now, []byte("First log")))
			logReader.AddError(errors.New("First Error"))

			go consoleTailedLogsOutputter.OutputTailedLogs("my-app-guid")

			Eventually(logReader.GetAppGuid).Should(Equal("my-app-guid"))

			logOutputBufferString := fmt.Sprintf("%s [%s|%s] First log\n", colors.Cyan(now.Format("01/02 15:04:05.00")), colors.Yellow("RTR"), colors.Yellow("1"))
			Eventually(outputBuffer).Should(test_helpers.Say(logOutputBufferString))
			Eventually(outputBuffer).Should(test_helpers.Say("First Error\n"))
		})
	})

	Describe("OutputDebugLogs", func() {
		It("tails logs with pretty formatting", func() {
			now := time.Now()
			logReader.AddLog(buildLogMessage("rep", "cell-1", now, []byte("First log")))
			logReader.AddError(errors.New("First Error"))

			go consoleTailedLogsOutputter.OutputDebugLogs(true)

			Eventually(logReader.GetAppGuid).Should(Equal(reserved_app_ids.LatticeDebugLogStreamAppId))

			Eventually(outputBuffer).Should(test_helpers.Say("rep"))
			Eventually(outputBuffer).Should(test_helpers.Say("cell-1"))
			Eventually(outputBuffer).Should(test_helpers.Say(now.Format("01/02 15:04:05.00")))
			Eventually(outputBuffer).Should(test_helpers.Say("First log"))
			Eventually(outputBuffer).Should(test_helpers.Say("First Error\n"))
		})

		It("tails logs without pretty formatting", func() {
			now := time.Now()
			logReader.AddLog(buildLogMessage("rep", "cell-1", now, []byte("First log")))
			logReader.AddError(errors.New("First Error"))

			go consoleTailedLogsOutputter.OutputDebugLogs(false)

			Eventually(logReader.GetAppGuid).Should(Equal(reserved_app_ids.LatticeDebugLogStreamAppId))

			Eventually(outputBuffer).Should(test_helpers.Say(now.Format("01/02 15:04:05.00")))
			Eventually(outputBuffer).Should(test_helpers.Say("rep"))
			Eventually(outputBuffer).Should(test_helpers.Say("cell-1"))
			Eventually(outputBuffer).Should(test_helpers.Say("First log\n"))
			Eventually(outputBuffer).Should(test_helpers.Say("First Error\n"))
		})
	})

	Describe("StopOutputting", func() {
		It("stops outputting logs", func() {
			go consoleTailedLogsOutputter.OutputTailedLogs("my-app-guid")

			consoleTailedLogsOutputter.StopOutputting()
			Expect(logReader.IsLogTailStopped()).To(BeTrue())
		})
	})
})
