package metrics_test

import (
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Metrics", func() {
	var fakeMetricSender *fake.FakeMetricSender

	BeforeEach(func() {
		fakeMetricSender = fake.NewFakeMetricSender()
		metricBatcher := metricbatcher.New(fakeMetricSender, time.Millisecond)
		metrics.Initialize(fakeMetricSender, metricBatcher)
	})

	It("delegates SendValue", func() {
		metrics.SendValue("metric", 42.42, "answers")

		Expect(fakeMetricSender.GetValue("metric").Value).To(Equal(42.42))
		Expect(fakeMetricSender.GetValue("metric").Unit).To(Equal("answers"))
	})

	It("delegates IncrementCounter", func() {
		metrics.IncrementCounter("count")

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(1))

		metrics.IncrementCounter("count")

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(2))
	})

	It("delegates BatchIncrementCounter", func() {
		metrics.BatchIncrementCounter("count")
		time.Sleep(2 * time.Millisecond)
		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(1))

		metrics.BatchIncrementCounter("count")
		time.Sleep(2 * time.Millisecond)
		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(2))
	})

	It("delegates AddToCounter", func() {
		metrics.AddToCounter("count", 5)

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(5))

		metrics.AddToCounter("count", 10)

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(15))
	})

	It("delegates BatchAddCounter", func() {
		metrics.BatchAddCounter("count", 3)
		time.Sleep(2 * time.Millisecond)
		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(3))

		metrics.BatchAddCounter("count", 7)
		time.Sleep(2 * time.Millisecond)
		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(10))
	})

	It("delegates SendContainerMetric", func() {
		appGuid := "some_app_guid"
		metrics.SendContainerMetric(appGuid, 7, 42.42, 1234, 123412341234)

		Expect(fakeMetricSender.GetContainerMetric(appGuid).ApplicationId).To(Equal(appGuid))
		Expect(fakeMetricSender.GetContainerMetric(appGuid).InstanceIndex).To(BeEquivalentTo(7))
		Expect(fakeMetricSender.GetContainerMetric(appGuid).CpuPercentage).To(BeEquivalentTo(42.42))
		Expect(fakeMetricSender.GetContainerMetric(appGuid).MemoryBytes).To(BeEquivalentTo(1234))
		Expect(fakeMetricSender.GetContainerMetric(appGuid).DiskBytes).To(BeEquivalentTo(123412341234))
	})

	Context("when Metric Sender is not initialized", func() {

		BeforeEach(func() {
			metrics.Initialize(nil, nil)
		})

		It("SendValue is a no-op", func() {
			err := metrics.SendValue("metric", 42.42, "answers")

			Expect(err).ToNot(HaveOccurred())
		})

		It("IncrementCounter is a no-op", func() {
			err := metrics.IncrementCounter("count")

			Expect(err).ToNot(HaveOccurred())
		})

		It("AddToCounter is a no-op", func() {
			err := metrics.AddToCounter("count", 10)

			Expect(err).ToNot(HaveOccurred())
		})

		It("SendContainerMetric is a no-op", func() {
			appGuid := "some_app_guid"
			err := metrics.SendContainerMetric(appGuid, 0, 42.42, 1234, 123412341234)

			Expect(err).ToNot(HaveOccurred())
		})

	})
})
