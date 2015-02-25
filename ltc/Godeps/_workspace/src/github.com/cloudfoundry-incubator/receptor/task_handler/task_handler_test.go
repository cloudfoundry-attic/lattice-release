package task_handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor/task_handler"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskHandler", func() {
	var (
		enqueue chan models.Task

		server *httptest.Server

		payload  []byte
		response *http.Response
	)

	BeforeEach(func() {
		logger := lager.NewLogger("task-watcher-test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.INFO))

		enqueue = make(chan models.Task, 100)

		server = httptest.NewServer(task_handler.NewHandler(enqueue, logger))
	})

	AfterEach(func() {
		server.Close()
	})

	JustBeforeEach(func() {
		var err error

		response, err = http.Post(server.URL, "application/json", bytes.NewBuffer(payload))
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("when the handler receives a task", func() {
		task := models.Task{
			TaskGuid:      "some-guid",
			Failed:        true,
			FailureReason: "'cause",
			Result:        "some result",
			Stack:         "java",
			Domain:        "some-domain",
			Action:        &models.RunAction{Path: "true"},
		}

		BeforeEach(func() {
			var err error
			payload, err = models.ToJSONArray(task)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("enqueues the task on the worker", func() {
			Ω(enqueue).Should(Receive(Equal(task)))
		})

		It("returns 202", func() {
			Ω(response.StatusCode).Should(Equal(http.StatusAccepted))
		})
	})

	Describe("when the handler receives a bogus payload", func() {
		BeforeEach(func() {
			payload = []byte("ß")
		})

		It("returns 400", func() {
			Ω(response.StatusCode).Should(Equal(http.StatusBadRequest))
		})

		It("does not enqueue anything somehow", func() {
			Ω(enqueue).ShouldNot(Receive())
		})
	})
})
