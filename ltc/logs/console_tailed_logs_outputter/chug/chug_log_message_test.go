package chug_test

import (
	"errors"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/chug"
	"github.com/cloudfoundry/noaa/events"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("ChugLogMessage", func() {

	var (
		entry chug.Entry
		input []byte
	)

	itReturnsRawData := func() {
		Expect(entry.IsLager).To(BeFalse())
		Expect(entry.Log).To(BeZero())
		Expect(entry.Raw).To(Equal(input))
	}

	It("chugs a message that is not lager", func() {
		now := time.Now()
		unixTime := now.UnixNano()
		sourceType := "rep"
		sourceInstance := "cell-77"
		input = []byte(`ABC 123`)
		logMessage := &events.LogMessage{
			Message:        input,
			Timestamp:      &unixTime,
			SourceType:     &sourceType,
			SourceInstance: &sourceInstance,
		}

		entry = chug.ChugLogMessage(logMessage)

		Expect(entry).ToNot(BeNil())
		Expect(entry.LogMessage).ToNot(BeNil())
		Expect(entry.LogMessage).To(Equal(logMessage))

		itReturnsRawData()
	})

	It("chugs a message that is lager", func() {
		now := time.Now()
		unixTime := now.UnixNano()
		sourceType := "rep"
		sourceInstance := "cell-77"
		logPayload := []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":1,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","session":"7.1.10","trace":"trace-me-now"}}`)

		logMessage := &events.LogMessage{
			Message:        logPayload,
			Timestamp:      &unixTime,
			SourceType:     &sourceType,
			SourceInstance: &sourceInstance,
		}

		entry := chug.ChugLogMessage(logMessage)

		Expect(entry).ToNot(BeNil())
		Expect(entry.LogMessage).ToNot(BeNil())
		Expect(entry.LogMessage).To(Equal(logMessage))

		Expect(entry.Raw).ToNot(BeEmpty())
		Expect(entry.Raw).To(Equal(logPayload))

		Expect(entry.IsLager).To(BeTrue())
	})

	It("chugs a message that has invalid json", func() {
		input = []byte(`{"timestamp`)
		logMessage := &events.LogMessage{
			Message: input,
		}

		entry = chug.ChugLogMessage(logMessage)

		Expect(entry).ToNot(BeNil())
		Expect(entry.LogMessage).ToNot(BeNil())
		Expect(entry.LogMessage).To(Equal(logMessage))

		itReturnsRawData()
	})

	It("populates Entry.Log with a lager message", func() {

		logPayload := []byte(`{"timestamp":"1429296198.620077372","source":"rep","message":"rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container","log_level":2,"data":{"container-guid":"app-9eb203ad-72f3-4f26-6424-48f20dc12298","session":"7.1.10","trace":"trace-me-now","error":"your the man now dog"}}`)
		logMessage := &events.LogMessage{
			Message: logPayload,
		}

		entry := chug.ChugLogMessage(logMessage)

		Expect(entry).ToNot(BeNil())
		Expect(entry.IsLager).To(BeTrue())

		timestamp, err := strconv.ParseFloat("1429296198.620077372", 64)
		Expect(err).ToNot(HaveOccurred())
		Expect(entry.Log.Timestamp).To(Equal(time.Unix(0, int64(timestamp*1e9))))

		Expect(entry.Log.LogLevel).To(Equal(lager.LogLevel(2)))

		Expect(entry.Log.Source).To(Equal("rep"))
		Expect(entry.Log.Message).To(Equal("rep.event-consumer.operation-stream.executing-container-operation.succeeded-fetch-container"))
		Expect(entry.Log.Session).To(Equal("7.1.10"))
		Expect(entry.Log.Trace).To(Equal("trace-me-now"))
		Expect(entry.Log.Error).To(Equal(errors.New("your the man now dog")))
		Expect(entry.Log.Data).To(Equal(lager.Data{"container-guid": "app-9eb203ad-72f3-4f26-6424-48f20dc12298"}))
	})

	Context("handling malformed/non-lager data", func() {

		Context("when the timestamp is invalid", func() {
			It("returns raw data", func() {
				input = []byte(`{"timestamp":"tomorrow","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","error":7}}`)
				logMessage := &events.LogMessage{
					Message: input,
				}

				entry = chug.ChugLogMessage(logMessage)

				itReturnsRawData()
			})
		})

		Context("when the error does not parse", func() {
			It("returns raw data", func() {
				input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","error":7}}`)
				logMessage := &events.LogMessage{
					Message: input,
				}

				entry = chug.ChugLogMessage(logMessage)

				itReturnsRawData()
			})
		})

		Context("when the trace does not parse", func() {
			It("returns raw data", func() {
				input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","trace":7}}`)
				logMessage := &events.LogMessage{
					Message: input,
				}

				entry = chug.ChugLogMessage(logMessage)

				itReturnsRawData()
			})
		})

		Context("when the session does not parse", func() {
			It("returns raw data", func() {
				input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","session":7}}`)
				logMessage := &events.LogMessage{
					Message: input,
				}

				entry = chug.ChugLogMessage(logMessage)

				itReturnsRawData()
			})
		})

	})

})
