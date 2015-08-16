package logs_test

import (
	"github.com/cloudfoundry/dropsonde/log_sender/fake"
	"github.com/cloudfoundry/dropsonde/logs"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	var fakeLogSender *fake.FakeLogSender

	BeforeEach(func() {
		fakeLogSender = fake.NewFakeLogSender()
		logs.Initialize(fakeLogSender)
	})

	It("delegates SendAppLog", func() {
		logs.SendAppLog("app-id", "custom-log-message", "App", "0")

		Expect(fakeLogSender.GetLogs()).To(HaveLen(1))
		Expect(fakeLogSender.GetLogs()[0]).To(Equal(fake.Log{AppId: "app-id", Message: "custom-log-message", SourceType: "App", SourceInstance: "0", MessageType: "OUT"}))
	})

	It("delegates SendAppErrorLog", func() {
		logs.SendAppErrorLog("app-id", "custom-log-error-message", "App", "0")

		Expect(fakeLogSender.GetLogs()).To(HaveLen(1))
		Expect(fakeLogSender.GetLogs()[0]).To(Equal(fake.Log{AppId: "app-id", Message: "custom-log-error-message", SourceType: "App", SourceInstance: "0", MessageType: "ERR"}))
	})

	Context("when errors occur", func() {
		BeforeEach(func() {
			fakeLogSender.ReturnError = errors.New("error occurred")
		})

		It("SendAppLog returns error", func() {
			err := logs.SendAppLog("app-id", "custom-log-message", "App", "0")
			Expect(err).To(HaveOccurred())
		})

		It("SendAppErrorLog returns error", func() {
			err := logs.SendAppErrorLog("app-id", "custom-log-error-message", "App", "0")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when Metric Sender is not initialized", func() {
		BeforeEach(func() {
			logs.Initialize(nil)
		})

		It("SendAppLog is a no-op", func() {
			err := logs.SendAppLog("app-id", "custom-log-message", "App", "0")
			Expect(err).ToNot(HaveOccurred())
		})

		It("SendAppErrorLog is a no-op", func() {
			err := logs.SendAppErrorLog("app-id", "custom-log-error-message", "App", "0")
			Expect(err).ToNot(HaveOccurred())
		})

		It("ScanLogStream is a no-op", func() {
			Expect(func() { logs.ScanLogStream("app-id", "src-type", "src-instance", nil) }).ShouldNot(Panic())
		})

		It("ScanErrorLogStream is a no-op", func() {
			Expect(func() { logs.ScanErrorLogStream("app-id", "src-type", "src-instance", nil) }).ShouldNot(Panic())
		})

	})
})
