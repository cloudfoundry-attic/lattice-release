package task_runner

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type CreateTaskParams struct {
	receptorRequest receptor.TaskCreateRequest
}

func NewCreateTaskParams(action models.Action, taskGuid, rootFS, domain, logSource string, egressRules []models.SecurityGroupRule) CreateTaskParams {
	return CreateTaskParams{
		receptor.TaskCreateRequest{
			Action:      action,
			LogGuid:     taskGuid,
			MetricsGuid: taskGuid,
			TaskGuid:    taskGuid,
			RootFS:      rootFS,
			Domain:      domain,
			LogSource:   logSource,
			EgressRules: egressRules,
		},
	}
}

func (c *CreateTaskParams) GetReceptorRequest() receptor.TaskCreateRequest {
	return c.receptorRequest
}
