package task_runner

import (
	"encoding/json"
	"errors"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/receptor"
)

const (
	AttemptedToCreateLatticeDebugErrorMessage = reserved_app_ids.LatticeDebugLogStreamAppId + " is a reserved app name. It is used internally to stream debug logs for lattice components."
)

//go:generate counterfeiter -o fake_task_runner/fake_task_runner.go . TaskRunner
type TaskRunner interface {
	SubmitTask(submitTaskJson []byte) (string, error)
}

type taskRunner struct {
	receptorClient receptor.Client
}

func New(receptorClient receptor.Client) TaskRunner {
	return &taskRunner{receptorClient}
}

func (taskRunner *taskRunner) SubmitTask(submitTaskJson []byte) (string, error) {
	task := receptor.TaskCreateRequest{}
	if err := json.Unmarshal(submitTaskJson, &task); err != nil {
		return "", err
	}

	if task.TaskGuid == reserved_app_ids.LatticeDebugLogStreamAppId {
		return task.TaskGuid, errors.New(AttemptedToCreateLatticeDebugErrorMessage)
	}

	submittedTasks, err := taskRunner.receptorClient.Tasks()
	if err != nil {
		return task.TaskGuid, err
	}
	for _, submittedTask := range submittedTasks {
		if task.TaskGuid == submittedTask.TaskGuid {
			return task.TaskGuid, errors.New(task.TaskGuid + " has already been submitted")
		}
	}

	if err := taskRunner.receptorClient.UpsertDomain("lattice", 0); err != nil {
		return task.TaskGuid, err
	}

	return task.TaskGuid, taskRunner.receptorClient.CreateTask(task)
}
