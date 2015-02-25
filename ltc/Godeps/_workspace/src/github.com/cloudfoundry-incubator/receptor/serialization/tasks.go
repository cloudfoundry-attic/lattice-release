package serialization

import (
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func TaskFromRequest(req receptor.TaskCreateRequest) (models.Task, error) {
	var u *url.URL
	if req.CompletionCallbackURL != "" {
		var err error
		u, err = url.ParseRequestURI(req.CompletionCallbackURL)
		if err != nil {
			return models.Task{}, err
		}
	}

	task := models.Task{
		Action:                req.Action,
		Annotation:            req.Annotation,
		CompletionCallbackURL: u,
		CPUWeight:             req.CPUWeight,
		DiskMB:                req.DiskMB,
		Domain:                req.Domain,
		EnvironmentVariables:  EnvironmentVariablesToModel(req.EnvironmentVariables),
		LogGuid:               req.LogGuid,
		LogSource:             req.LogSource,
		MetricsGuid:           req.MetricsGuid,
		MemoryMB:              req.MemoryMB,
		ResultFile:            req.ResultFile,
		RootFSPath:            req.RootFSPath,
		Stack:                 req.Stack,
		TaskGuid:              req.TaskGuid,
		Privileged:            req.Privileged,
		EgressRules:           req.EgressRules,
	}

	return task, nil
}

func TaskToResponse(task models.Task) receptor.TaskResponse {
	url := ""
	if task.CompletionCallbackURL != nil {
		url = task.CompletionCallbackURL.String()
	}

	return receptor.TaskResponse{
		Action:                task.Action,
		Annotation:            task.Annotation,
		CompletionCallbackURL: url,
		CPUWeight:             task.CPUWeight,
		DiskMB:                task.DiskMB,
		Domain:                task.Domain,
		EnvironmentVariables:  EnvironmentVariablesFromModel(task.EnvironmentVariables),
		CellID:                task.CellID,
		LogGuid:               task.LogGuid,
		LogSource:             task.LogSource,
		MetricsGuid:           task.MetricsGuid,
		MemoryMB:              task.MemoryMB,
		Privileged:            task.Privileged,
		RootFSPath:            task.RootFSPath,
		Stack:                 task.Stack,
		TaskGuid:              task.TaskGuid,

		CreatedAt:     task.CreatedAt,
		Failed:        task.Failed,
		FailureReason: task.FailureReason,
		Result:        task.Result,
		State:         taskStateToResponseState(task.State),
		EgressRules:   task.EgressRules,
	}
}

func taskStateToResponseState(state models.TaskState) string {
	switch state {
	case models.TaskStateInvalid:
		return receptor.TaskStateInvalid
	case models.TaskStatePending:
		return receptor.TaskStatePending
	case models.TaskStateRunning:
		return receptor.TaskStateRunning
	case models.TaskStateCompleted:
		return receptor.TaskStateCompleted
	case models.TaskStateResolving:
		return receptor.TaskStateResolving
	}

	return ""
}
