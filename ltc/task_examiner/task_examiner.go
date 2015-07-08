package task_examiner

import (
	"errors"

	"github.com/cloudfoundry-incubator/receptor"
)

const TaskNotFoundErrorMessage = "Task not found."

type TaskInfo struct {
	TaskGuid      string
	State         string
	CellID        string
	Failed        bool
	FailureReason string
	Result        string
}

//go:generate counterfeiter -o fake_task_examiner/fake_task_examiner.go . TaskExaminer
type TaskExaminer interface {
	TaskStatus(taskName string) (TaskInfo, error)
	ListTasks() ([]TaskInfo, error)
}

type taskExaminer struct {
	receptorClient receptor.Client
}

func New(receptorClient receptor.Client) TaskExaminer {
	return &taskExaminer{receptorClient}
}

func (e *taskExaminer) TaskStatus(taskName string) (TaskInfo, error) {
	taskResponse, err := e.receptorClient.GetTask(taskName)
	if err != nil {
		if receptorError, ok := err.(receptor.Error); ok {
			if receptorError.Type == receptor.TaskNotFound {
				return TaskInfo{}, errors.New(TaskNotFoundErrorMessage)
			}
		}
		return TaskInfo{}, err
	}

	return TaskInfo{
		TaskGuid:      taskResponse.TaskGuid,
		State:         taskResponse.State,
		CellID:        taskResponse.CellID,
		Failed:        taskResponse.Failed,
		FailureReason: taskResponse.FailureReason,
		Result:        taskResponse.Result,
	}, nil
}

func (e *taskExaminer) ListTasks() ([]TaskInfo, error) {
	taskList, err := e.receptorClient.Tasks()
	if err != nil {
		return nil, err
	}
	taskInfoList := make([]TaskInfo, 0, len(taskList))
	for _, task := range taskList {
		taskInfo := TaskInfo{
			TaskGuid:      task.TaskGuid,
			CellID:        task.CellID,
			Failed:        task.Failed,
			FailureReason: task.FailureReason,
			Result:        task.Result,
			State:         task.State,
		}
		taskInfoList = append(taskInfoList, taskInfo)
	}
	return taskInfoList, err
}
