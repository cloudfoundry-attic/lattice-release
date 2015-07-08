package prettify_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/prettify"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Prettify", func() {
	var (
		input []byte
		regex = regexp.MustCompile("[-/\\\\^$*+?.()|[\\]{}]")
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

	regexSafe := func(matcher string) string {
		return regex.ReplaceAllStringFunc(matcher, func(s string) string {
			return fmt.Sprintf("\\%s", s)
		})
	}

	Context("when the message has lager to chug", func() {
		It("pretties the text for lager message", func() {
			input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","session":"7.1.10","trace":"trace-me-now"}}`)
			lagerTimestamp := "1429296198.620077372"
			lagerUnixTime, err := strconv.ParseFloat(lagerTimestamp, 64)
			Expect(err).ToNot(HaveOccurred())
			logMessage := buildLogMessage("rep", "cell-77", time.Time{}, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).ToNot(BeEmpty())

			var outputExpects []string
			outputExpects = append(outputExpects, regexSafe(""))
			outputExpects = append(outputExpects, regexSafe("rep"))
			outputExpects = append(outputExpects, regexSafe("cell-77"))
			outputExpects = append(outputExpects, regexSafe("INFO"))
			outputExpects = append(outputExpects, regexSafe(time.Unix(0, int64(lagerUnixTime*1e9)).Format("01/02 15:04:05.00")))
			outputExpects = append(outputExpects, regexSafe("7.1.10"))
			outputExpects = append(outputExpects, regexSafe("rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container"))
			outputExpects = append(outputExpects, regexSafe("\n"))
			outputExpects = append(outputExpects, regexSafe(`{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298"}`))
			outputExpects = append(outputExpects, regexSafe(""))
			regexPattern := strings.Join(outputExpects, ".*")

			Expect(prettyLog).To(MatchRegexp(regexPattern))
		})

		Describe("output formatting", func() {
			It("prints a separate line for the error contents", func() {
				input = []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":2,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","error":"unicorns can fly"}}`)
				logMessage := buildLogMessage("rep", "cell-77", time.Time{}, input)

				prettyLog := prettify.Prettify(logMessage)

				var outputExpects []string
				outputExpects = append(outputExpects, regexSafe("\n"))
				outputExpects = append(outputExpects, regexSafe("unicorns can fly"))
				outputExpects = append(outputExpects, regexSafe("\n"))
				regexPattern := strings.Join(outputExpects, ".*")

				Expect(prettyLog).To(MatchRegexp(regexPattern))
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
		})

		Describe("output coloring", func() {
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

			Context("for the various log levels", func() {

				buildInputByLevel := func(logLevel lager.LogLevel) []byte {
					inputPrefix := `{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":`
					inputSuffix := `,"data":{}}`

					var inputMessage []string
					inputMessage = append(inputMessage, inputPrefix)
					inputMessage = append(inputMessage, fmt.Sprint(logLevel))
					inputMessage = append(inputMessage, inputSuffix)
					return []byte(strings.Join(inputMessage, ""))
				}

				It("colors the INFO with SourceType-specific color", func() {
					logMessage := buildLogMessage("generic", "instance", time.Time{}, buildInputByLevel(lager.INFO))

					prettyLog := prettify.Prettify(logMessage)

					var outputExpects []string
					outputExpects = append(outputExpects, regexSafe(""))
					outputExpects = append(outputExpects, regexSafe(colors.ColorDefault))
					outputExpects = append(outputExpects, regexSafe("[INFO]"))
					outputExpects = append(outputExpects, regexSafe(colors.ColorDefault))
					outputExpects = append(outputExpects, regexSafe(""))
					regexPattern := strings.Join(outputExpects, ".*")

					// TODO: there are other default color tokens in this string, improve test
					Expect(prettyLog).To(MatchRegexp(regexPattern))
				})

				It("colors the DEBUG as Gray", func() {
					logMessage := buildLogMessage("", "", time.Time{}, buildInputByLevel(lager.DEBUG))

					prettyLog := prettify.Prettify(logMessage)

					var outputExpects []string
					outputExpects = append(outputExpects, regexSafe(""))
					outputExpects = append(outputExpects, regexSafe(colors.ColorGray))
					outputExpects = append(outputExpects, regexSafe("[DEBUG]"))
					outputExpects = append(outputExpects, regexSafe(colors.ColorDefault))
					outputExpects = append(outputExpects, regexSafe(""))
					regexPattern := strings.Join(outputExpects, ".*")

					Expect(prettyLog).To(MatchRegexp(regexPattern))
				})

				It("colors the ERROR as Red", func() {
					logMessage := buildLogMessage("", "", time.Time{}, buildInputByLevel(lager.ERROR))

					prettyLog := prettify.Prettify(logMessage)

					var outputExpects []string
					outputExpects = append(outputExpects, regexSafe(""))
					outputExpects = append(outputExpects, regexSafe(colors.ColorRed))
					outputExpects = append(outputExpects, regexSafe("[ERROR]"))
					outputExpects = append(outputExpects, regexSafe(colors.ColorDefault))
					outputExpects = append(outputExpects, regexSafe(""))
					regexPattern := strings.Join(outputExpects, ".*")

					Expect(prettyLog).To(MatchRegexp(regexPattern))
				})

				It("colors the FATAL as Red", func() {
					logMessage := buildLogMessage("", "", time.Time{}, buildInputByLevel(lager.FATAL))

					prettyLog := prettify.Prettify(logMessage)

					var outputExpects []string
					outputExpects = append(outputExpects, regexSafe(""))
					outputExpects = append(outputExpects, regexSafe(colors.ColorRed))
					outputExpects = append(outputExpects, regexSafe("[FATAL]"))

					outputExpects = append(outputExpects, regexSafe(colors.ColorDefault))
					outputExpects = append(outputExpects, regexSafe(""))
					regexPattern := strings.Join(outputExpects, ".*")
					Expect(prettyLog).To(MatchRegexp(regexPattern))
				})
			})
		})
	})

	Context("when the messages are not lager-formatted", func() {
		It("pads and spaces the output for non-lager", func() {
			now := time.Now()
			input = []byte("ABC 123")
			logMessage := buildLogMessage("rep", "cell-77", now, input)

			prettyLog := prettify.Prettify(logMessage)

			Expect(prettyLog).ToNot(BeEmpty())

			var outputExpects []string
			outputExpects = append(outputExpects, regexSafe(""))
			outputExpects = append(outputExpects, regexSafe("rep"))
			outputExpects = append(outputExpects, regexSafe("cell-77"))
			outputExpects = append(outputExpects, regexSafe(now.Format("01/02 15:04:05.00")))
			outputExpects = append(outputExpects, regexSafe("ABC 123"))
			outputExpects = append(outputExpects, regexSafe(""))
			regexPattern := strings.Join(outputExpects, ".*")

			Expect(prettyLog).To(MatchRegexp(regexPattern))
		})
	})
})
