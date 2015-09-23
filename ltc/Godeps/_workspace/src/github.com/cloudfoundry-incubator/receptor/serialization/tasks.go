package serialization

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
)

func TaskFromRequest(req receptor.TaskCreateRequest) (*models.Task, error) {
	task := &models.Task{
		Action:                req.Action,
		Annotation:            req.Annotation,
		CompletionCallbackUrl: req.CompletionCallbackURL,
		CpuWeight:             uint32(req.CPUWeight),
		DiskMb:                int32(req.DiskMB),
		Domain:                req.Domain,
		EnvironmentVariables:  req.EnvironmentVariables,
		LogGuid:               req.LogGuid,
		LogSource:             req.LogSource,
		MetricsGuid:           req.MetricsGuid,
		MemoryMb:              int32(req.MemoryMB),
		ResultFile:            req.ResultFile,
		RootFs:                req.RootFS,
		TaskGuid:              req.TaskGuid,
		Privileged:            req.Privileged,
		EgressRules:           req.EgressRules,
	}

	return task, nil
}

func TaskToResponse(task *models.Task) receptor.TaskResponse {
	return receptor.TaskResponse{
		Action:                task.Action,
		Annotation:            task.Annotation,
		CompletionCallbackURL: task.CompletionCallbackUrl,
		CPUWeight:             uint(task.CpuWeight),
		DiskMB:                int(task.DiskMb),
		Domain:                task.Domain,
		EnvironmentVariables:  task.EnvironmentVariables,
		CellID:                task.CellId,
		LogGuid:               task.LogGuid,
		LogSource:             task.LogSource,
		MetricsGuid:           task.MetricsGuid,
		MemoryMB:              int(task.MemoryMb),
		Privileged:            task.Privileged,
		RootFS:                task.RootFs,
		TaskGuid:              task.TaskGuid,

		CreatedAt:     task.CreatedAt,
		Failed:        task.Failed,
		FailureReason: task.FailureReason,
		Result:        task.Result,
		State:         taskStateToResponseState(task.State),
		EgressRules:   task.EgressRules,
	}
}

func taskStateToResponseState(state models.Task_State) string {
	switch state {
	case models.Task_Invalid:
		return receptor.TaskStateInvalid
	case models.Task_Pending:
		return receptor.TaskStatePending
	case models.Task_Running:
		return receptor.TaskStateRunning
	case models.Task_Completed:
		return receptor.TaskStateCompleted
	case models.Task_Resolving:
		return receptor.TaskStateResolving
	}

	return ""
}
