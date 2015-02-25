package sse_test

import (
	"time"

	. "github.com/vito/go-sse/sse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Event", func() {
	Describe("Encode", func() {
		It("encodes to a dispatchable event", func() {
			Ω(Event{
				ID:   "some-id",
				Name: "some-name",
				Data: []byte("some-data"),
			}.Encode()).Should(Equal("id: some-id\nevent: some-name\ndata: some-data\n\n"))
		})

		It("splits lines across multiple data segments", func() {
			Ω(Event{
				ID:   "some-id",
				Name: "some-name",
				Data: []byte("some-data\nsome-more-data\n"),
			}.Encode()).Should(Equal("id: some-id\nevent: some-name\ndata: some-data\ndata: some-more-data\ndata\n\n"))
		})

		It("includes retry if present", func() {
			Ω(Event{
				ID:    "some-id",
				Name:  "some-name",
				Data:  []byte("some-data"),
				Retry: 123 * time.Millisecond,
			}.Encode()).Should(Equal("id: some-id\nevent: some-name\nretry: 123\ndata: some-data\n\n"))
		})
	})

	Describe("Write", func() {
		var destination *gbytes.Buffer

		BeforeEach(func() {
			destination = gbytes.NewBuffer()
		})

		It("writes the encoded event to the destination", func() {
			event := Event{
				ID:   "some-id",
				Name: "some-name",
				Data: []byte("some-data\nsome-more-data\n"),
			}

			err := event.Write(destination)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(destination.Contents()).Should(Equal([]byte(event.Encode())))
		})
	})
})
