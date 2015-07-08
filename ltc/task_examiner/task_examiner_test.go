package task_examiner_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
)

var _ = Describe("TaskExaminer", func() {
	var (
		fakeReceptorClient *fake_receptor.FakeClient
		taskExaminer       task_examiner.TaskExaminer
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		taskExaminer = task_examiner.New(fakeReceptorClient)
	})

	Describe("TaskStatus", func() {
		BeforeEach(func() {
			getTaskResponse := receptor.TaskResponse{
				TaskGuid:      "boop",
				State:         receptor.TaskStateCompleted,
				CellID:        "cell-01",
				Failed:        false,
				FailureReason: "",
				Result:        "some-result",
			}
			fakeReceptorClient.GetTaskReturns(getTaskResponse, nil)
		})

		It("returns a task status", func() {
			taskInfo, err := taskExaminer.TaskStatus("boop")
			Expect(err).ToNot(HaveOccurred())
			Expect(taskInfo.TaskGuid).To(Equal("boop"))
			Expect(taskInfo.State).To(Equal(receptor.TaskStateCompleted))
			Expect(taskInfo.CellID).To(Equal("cell-01"))
			Expect(taskInfo.Failed).To(BeFalse())
			Expect(taskInfo.FailureReason).To(BeEmpty())
			Expect(taskInfo.Result).To(Equal("some-result"))

			Expect(fakeReceptorClient.GetTaskCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.GetTaskArgsForCall(0)).To(Equal("boop"))
		})

		Context("when the receptor returns errors", func() {
			It("returns exists false for TaskNotFound", func() {
				receptorError := receptor.Error{Type: receptor.TaskNotFound, Message: "could not locate this"}
				fakeReceptorClient.GetTaskReturns(receptor.TaskResponse{}, receptorError)

				_, err := taskExaminer.TaskStatus("boop1")
				Expect(err).To(MatchError(task_examiner.TaskNotFoundErrorMessage))
			})

			It("bubbles up error for receptor Error anything but TaskNotFound", func() {
				receptorError := receptor.Error{Type: receptor.TaskGuidAlreadyExists, Message: "could not locate this"}
				fakeReceptorClient.GetTaskReturns(receptor.TaskResponse{}, receptorError)

				_, err := taskExaminer.TaskStatus("boop1")
				Expect(err).To(MatchError(receptorError))
			})

			It("bubbles up error for non-receptor error", func() {
				fakeReceptorClient.GetTaskReturns(receptor.TaskResponse{}, errors.New("you done goofed"))

				_, err := taskExaminer.TaskStatus("boop")
				Expect(err).To(MatchError("you done goofed"))
			})
		})
	})

	Describe("ListTasks", func() {
		It("returns the list of task", func() {
			taskListReturns := []receptor.TaskResponse{
				receptor.TaskResponse{
					TaskGuid:      "task-guid-1",
					CellID:        "cell-01",
					Failed:        false,
					FailureReason: "",
					Result:        "Finished",
					State:         "COMPLETED",
				},
				receptor.TaskResponse{
					TaskGuid:      "task-guid-2",
					CellID:        "cell-02",
					Failed:        true,
					FailureReason: "failed",
					Result:        "Failed",
					State:         "COMPLETED",
				},
			}
			fakeReceptorClient.TasksReturns(taskListReturns, nil)

			taskList, err := taskExaminer.ListTasks()
			Expect(err).ToNot(HaveOccurred())
			Expect(taskList).To(HaveLen(2))

			task1 := taskList[0]
			Expect(task1.TaskGuid).To(Equal("task-guid-1"))
			Expect(task1.CellID).To(Equal("cell-01"))
			Expect(task1.FailureReason).To(Equal(""))
			Expect(task1.Result).To(Equal("Finished"))
			Expect(task1.State).To(Equal("COMPLETED"))

			task2 := taskList[1]
			Expect(task2.TaskGuid).To(Equal("task-guid-2"))
			Expect(task2.CellID).To(Equal("cell-02"))
			Expect(task2.Result).To(Equal("Failed"))
		})

		It("when receptor returns error", func() {
			fakeReceptorClient.TasksReturns(nil, errors.New("Client not reachable."))

			_, err := taskExaminer.ListTasks()
			Expect(err).To(MatchError("Client not reachable."))
		})

		It("when receptor returns empty list", func() {
			fakeReceptorClient.TasksReturns([]receptor.TaskResponse{}, nil)

			taskList, err := taskExaminer.ListTasks()
			Expect(err).ToNot(HaveOccurred())
			Expect(taskList).To(HaveLen(0))
		})
	})
})
