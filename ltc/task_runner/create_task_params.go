package task_runner

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type CreateTaskParams struct {
	receptorRequest receptor.TaskCreateRequest
}

func NewCreateTaskParams(action models.Action, taskGuid, rootFS, domain, logSource string, env map[string]string, egressRules []models.SecurityGroupRule, memoryMB, cpuWeight, diskMB int) CreateTaskParams {
	return CreateTaskParams{
		receptor.TaskCreateRequest{
			Action:               action,
			LogGuid:              taskGuid,
			MetricsGuid:          taskGuid,
			TaskGuid:             taskGuid,
			RootFS:               rootFS,
			Domain:               domain,
			LogSource:            logSource,
			EnvironmentVariables: buildReceptorEnvironment(env),
			EgressRules:          egressRules,
			Privileged:           true,
			MemoryMB:             memoryMB,
			CPUWeight:            uint(cpuWeight),
			DiskMB:               diskMB,
		},
	}
}

func (c *CreateTaskParams) GetReceptorRequest() receptor.TaskCreateRequest {
	return c.receptorRequest
}

func buildReceptorEnvironment(env map[string]string) []receptor.EnvironmentVariable {
	renv := []receptor.EnvironmentVariable{}
	for name, value := range env {
		renv = append(renv, receptor.EnvironmentVariable{Name: name, Value: value})
	}
	return renv
}
