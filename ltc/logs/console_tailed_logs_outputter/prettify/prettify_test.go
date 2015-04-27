package prettify_test

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/prettify"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry/noaa/events"
)

var _ = Describe("Prettify", func() {

	var (
		input []byte
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

	It("pretties the text for lager message", func() {
		input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","session":"7.1.10","trace":"trace-me-now"}}`)
		lagerTimestamp := "1429296198.620077372"
		lagerUnixTime, err := strconv.ParseFloat(lagerTimestamp, 64)
		Expect(err).ToNot(HaveOccurred())
		logMessage := buildLogMessage("rep", "cell-77", time.Time{}, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).ToNot(BeEmpty())
		Expect(prettyLog).To(ContainSubstring(`rep`))
		Expect(prettyLog).To(ContainSubstring(`cell-77`))
		Expect(prettyLog).To(ContainSubstring(`INFO`))
		Expect(prettyLog).To(ContainSubstring(time.Unix(0, int64(lagerUnixTime*1e9)).Format("01/02 15:04:05.00")))
		Expect(prettyLog).To(ContainSubstring(`7.1.10`))
		Expect(prettyLog).To(ContainSubstring(`rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container`))
		Expect(prettyLog).To(ContainSubstring(`{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298"}`))
	})

	It("pretties the text for non-lager message", func() {
		now := time.Now()
		input = []byte(`ABC 123`)
		logMessage := buildLogMessage("rep", "cell-77", now, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).ToNot(BeEmpty())
		Expect(prettyLog).To(ContainSubstring(`rep`))
		Expect(prettyLog).To(ContainSubstring(`cell-77`))
		Expect(prettyLog).To(ContainSubstring(`ABC 123`))
		Expect(prettyLog).To(ContainSubstring(now.Format("01/02 15:04:05.00")))
	})

	It("prints a newline for non-empty data", func() {
		input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298"}}`)
		logMessage := buildLogMessage("rep", "cell-77", time.Time{}, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).To(ContainSubstring("\n"))
		Expect(prettyLog).To(ContainSubstring(`{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298"}`))
	})

	It("does not print newline for empty data", func() {
		input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{}}`)
		logMessage := buildLogMessage("rep", "cell-77", time.Time{}, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).ToNot(ContainSubstring("{}"))
		Expect(prettyLog).ToNot(ContainSubstring("\n"))
	})

	It("highlights the source type column with app-specific color", func() {
		input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{}}`)
		logMessage := buildLogMessage("rep", "cell-77", time.Time{}, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).To(MatchRegexp(strings.Replace(colors.Colorize("\x1b[34m", "rep"), "[", `\[`, -1)))
	})

	Context("when the source type is unknown", func() {
		It("highlights the source type column with no color", func() {
			input = []byte(`{"timestamp":"1429296198.620077372","source":"happyjoy","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","session":"7.1.10"}}`)
			logMessage := buildLogMessage("happyjoy", "", time.Time{}, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).To(MatchRegexp(strings.Replace(colors.Colorize("\x1b[0m", "happyjoy"), "[", `\[`, -1)))
		})
	})

	It("pads and spaces the output for lager", func() {
		now := time.Now()
		lagerTimestamp := "1429296198.620077372"
		lagerUnixTime, err := strconv.ParseFloat(lagerTimestamp, 64)
		Expect(err).ToNot(HaveOccurred())
		input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","session":"7.1.10"}}`)
		logMessage := buildLogMessage("rep", "cell-77", now, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).ToNot(BeEmpty())

		Expect(prettyLog).To(MatchRegexp(`\S{4}rep\S{4}\s{9}`))
		Expect(prettyLog).To(MatchRegexp(`^.{22}cell-77\s{2}`))
		Expect(prettyLog).To(MatchRegexp(`^.{34}\S{4}[INFO]\S{4}`))
		Expect(prettyLog).To(MatchRegexp(`^.{48}` + time.Unix(0, int64(lagerUnixTime*1e9)).Format("01/02 15:04:05.00")))
		Expect(prettyLog).To(MatchRegexp(`^.{66}7.1.10`))
		Expect(prettyLog).To(MatchRegexp(`^.{81}rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container`))

		Expect(prettyLog).To(MatchRegexp("^.*\\n\\s{63}\\{\"container-guid\""))
	})

	It("pads and spaces the output for non-lager", func() {
		now := time.Now()
		input = []byte(`ABC 123`)
		logMessage := buildLogMessage("rep", "cell-77", now, input)

		prettyLog := prettify.Prettify(logMessage)

		Expect(prettyLog).To(MatchRegexp(`\S{4}rep\S{4}\s{9}`))
		Expect(prettyLog).To(MatchRegexp(`^.{22}cell-77\s{2}`))
		Expect(prettyLog).To(MatchRegexp(fmt.Sprintf(`^.{39}%s`, now.Format("01/02 15:04:05.00"))))
		Expect(prettyLog).To(MatchRegexp(`^.{72}ABC 123`))
	})

	Context("for the various log levels", func() {

		It("colors the INFO with SourceType-specific color", func() {
			input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{}}`)
			logMessage := buildLogMessage("rep", "", time.Time{}, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).To(MatchRegexp(strings.Replace(colors.Colorize("\x1b[34m", "[INFO]"), "[", `\[`, -1)))
		})

		It("colors the DEBUG as Gray", func() {
			input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":0,"data":{}}`)
			logMessage := buildLogMessage("", "", time.Time{}, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).To(ContainSubstring(colors.Gray("[DEBUG]")))
		})

		It("colors the ERROR as Red", func() {
			input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":2,"data":{}}`)
			logMessage := buildLogMessage("", "", time.Time{}, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).To(ContainSubstring(colors.Red("[ERROR]")))
		})

		It("colors the FATAL as Red", func() {
			input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":3,"data":{}}`)
			logMessage := buildLogMessage("", "", time.Time{}, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).To(ContainSubstring(colors.Red("[FATAL]")))
		})
	})
})
