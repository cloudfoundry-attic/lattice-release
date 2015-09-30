package task_runner

import (
	"encoding/json"
	"errors"
	"fmt"
        "time"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/receptor"
)

const (
	AttemptedToCreateLatticeDebugErrorMessage = reserved_app_ids.LatticeDebugLogStreamAppId + " is a reserved app name. It is used internally to stream debug logs for lattice components."

	taskDomain = "lattice"
)

//go:generate counterfeiter -o fake_task_runner/fake_task_runner.go . TaskRunner
type TaskRunner interface {
	CreateTask(createTaskParams CreateTaskParams) error
	SubmitTask(submitTaskJson []byte) (string, error)
	DeleteTask(taskGuid string) error
	CancelTask(taskGuid string) error
}

type taskRunner struct {
	receptorClient receptor.Client
	taskExaminer   task_examiner.TaskExaminer
}

func New(receptorClient receptor.Client, taskExaminer task_examiner.TaskExaminer) TaskRunner {
	return &taskRunner{receptorClient, taskExaminer}
}

func (taskRunner *taskRunner) CreateTask(createTaskParams CreateTaskParams) error {
	task := createTaskParams.GetReceptorRequest()

	if task.TaskGuid == reserved_app_ids.LatticeDebugLogStreamAppId {
		return errors.New(AttemptedToCreateLatticeDebugErrorMessage)
	}

	submittedTasks, err := taskRunner.receptorClient.Tasks()
	if err != nil {
		return err
	}
	for _, submittedTask := range submittedTasks {
		if task.TaskGuid == submittedTask.TaskGuid {
			return errors.New(task.TaskGuid + " has already been submitted")
		}
	}

	if err := taskRunner.receptorClient.UpsertDomain("lattice", 0); err != nil {
		return err
	}

	return taskRunner.receptorClient.CreateTask(task)
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

func (e *taskRunner) DeleteTask(taskGuid string) error {

        /*Ignoring the error of cancel task */
	e.receptorClient.CancelTask(taskGuid)
       ticker := time.NewTicker(time.Second * 1)
	count := 0
	defer ticker.Stop()

        for {
            select {
                case <- ticker.C:
                  count++;
		    if count == 30 {
			return errors.New("Delete not completed because the timer expired before to complete the cancel task sucessfully")
		    }

		    taskInfo, err := e.taskExaminer.TaskStatus(taskGuid)
                  if err == nil && taskInfo.State == receptor.TaskStateCompleted {
                        err:= e.receptorClient.DeleteTask(taskGuid)
                        return err
                }
            }
        }
}

func (e *taskRunner) CancelTask(taskGuid string) error {
	taskInfo, err := e.taskExaminer.TaskStatus(taskGuid)
	if err != nil {
		return err
	}
	if taskInfo.State != receptor.TaskStatePending && taskInfo.State != receptor.TaskStateRunning {
		return fmt.Errorf("Unable to cancel %s task", taskInfo.State)
	}
	return e.receptorClient.CancelTask(taskGuid)
}
