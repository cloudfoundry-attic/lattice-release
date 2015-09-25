package models

import (
	"encoding/json"
	"net/url"
	"regexp"

	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

var taskGuidPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (t *Task) LagerData() lager.Data {
	return lager.Data{
		"task-guid": t.TaskGuid,
		"domain":    t.Domain,
		"state":     t.State,
		"cell-id":   t.CellId,
	}
}

func (task *Task) Validate() error {
	var validationError ValidationError

	if task.Domain == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if !taskGuidPattern.MatchString(task.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if task.TaskDefinition == nil {
		validationError = validationError.Append(ErrInvalidField{"task_definition"})
	} else if defErr := task.TaskDefinition.Validate(); defErr != nil {
		validationError = validationError.Append(defErr)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (def *TaskDefinition) Validate() error {
	var validationError ValidationError

	if def.RootFs == "" {
		validationError = validationError.Append(ErrInvalidField{"rootfs"})
	} else {
		rootFsURL, err := url.Parse(def.RootFs)
		if err != nil || rootFsURL.Scheme == "" {
			validationError = validationError.Append(ErrInvalidField{"rootfs"})
		}
	}

	action := UnwrapAction(def.Action)
	if action == nil {
		validationError = validationError.Append(ErrInvalidActionType)
	} else {
		err := action.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if def.CpuWeight > 100 {
		validationError = validationError.Append(ErrInvalidField{"cpu_weight"})
	}

	if len(def.Annotation) > maximumAnnotationLength {
		validationError = validationError.Append(ErrInvalidField{"annotation"})
	}

	for _, rule := range def.EgressRules {
		err := rule.Validate()
		if err != nil {
			validationError = validationError.Append(ErrInvalidField{"egress_rules"})
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (task *Task) MarshalJSON() ([]byte, error) {
	var taskDef *TaskDefinition
	if task.TaskDefinition != nil {
		taskDef = task.TaskDefinition
	} else {
		taskDef = &TaskDefinition{}
	}

	b, err := json.Marshal(taskDef.Action)
	if err != nil {
		return nil, err
	}
	var oldAction oldmodels.Action
	if UnwrapAction(taskDef.Action) != nil {
		oldAction, err = oldmodels.UnmarshalAction(b)
		if err != nil {
			return nil, err
		}
	}

	b, err = json.Marshal(taskDef.EgressRules)
	if err != nil {
		return nil, err
	}
	var oldEgress []oldmodels.SecurityGroupRule
	err = json.Unmarshal(b, &oldEgress)

	oldUrl, err := url.Parse(taskDef.CompletionCallbackUrl)
	if err != nil {
		return nil, err
	}

	oldtask := oldmodels.Task{
		TaskGuid:              task.TaskGuid,
		Domain:                task.Domain,
		RootFS:                taskDef.RootFs,
		EnvironmentVariables:  EnvironmentVariablesFromProto(taskDef.EnvironmentVariables),
		CellID:                task.CellId,
		Action:                oldAction,
		ResultFile:            taskDef.ResultFile,
		Result:                task.Result,
		Failed:                task.Failed,
		FailureReason:         task.FailureReason,
		MemoryMB:              int(taskDef.MemoryMb),
		DiskMB:                int(taskDef.DiskMb),
		CPUWeight:             uint(taskDef.CpuWeight),
		Privileged:            taskDef.Privileged,
		LogGuid:               taskDef.LogGuid,
		LogSource:             taskDef.LogSource,
		MetricsGuid:           taskDef.MetricsGuid,
		CreatedAt:             task.CreatedAt,
		UpdatedAt:             task.UpdatedAt,
		FirstCompletedAt:      task.FirstCompletedAt,
		State:                 oldmodels.TaskState(Task_State_value[task.State.String()]),
		Annotation:            taskDef.Annotation,
		EgressRules:           oldEgress,
		CompletionCallbackURL: oldUrl,
	}

	return json.Marshal(&oldtask)
}

func (task *Task) UnmarshalJSON(data []byte) error {
	var oldtask oldmodels.Task
	err := json.Unmarshal(data, &oldtask)
	if err != nil {
		return err
	}

	b, err := oldmodels.MarshalAction(oldtask.Action)
	if err != nil {
		return err
	}
	var newAction Action
	err = json.Unmarshal(b, &newAction)

	b, err = json.Marshal(oldtask.EgressRules)
	if err != nil {
		return err
	}

	var newEgressRules []*SecurityGroupRule
	err = json.Unmarshal(b, &newEgressRules)
	if err != nil {
		return err
	}

	task.TaskDefinition = &TaskDefinition{}

	task.TaskGuid = oldtask.TaskGuid
	task.Domain = oldtask.Domain
	task.RootFs = oldtask.RootFS
	task.EnvironmentVariables = EnvironmentVariablesFromModel(oldtask.EnvironmentVariables)
	task.CellId = oldtask.CellID
	task.Action = &newAction
	task.ResultFile = oldtask.ResultFile
	task.Result = oldtask.Result
	task.Failed = oldtask.Failed
	task.FailureReason = oldtask.FailureReason
	task.MemoryMb = int32(oldtask.MemoryMB)
	task.DiskMb = int32(oldtask.DiskMB)
	task.CpuWeight = uint32(oldtask.CPUWeight)
	task.Privileged = oldtask.Privileged
	task.LogGuid = oldtask.LogGuid
	task.LogSource = oldtask.LogSource
	task.MetricsGuid = oldtask.MetricsGuid
	task.CreatedAt = oldtask.CreatedAt
	task.UpdatedAt = oldtask.UpdatedAt
	task.FirstCompletedAt = oldtask.FirstCompletedAt
	task.State = Task_State(oldtask.State)
	task.Annotation = oldtask.Annotation
	task.EgressRules = newEgressRules
	if oldtask.CompletionCallbackURL != nil {
		task.CompletionCallbackUrl = oldtask.CompletionCallbackURL.String()
	}

	return nil
}

func (req *DesireTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if req.Domain == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if req.TaskDefinition == nil {
		validationError = validationError.Append(ErrInvalidField{"task_definition"})
	} else if defErr := req.TaskDefinition.Validate(); defErr != nil {
		validationError = validationError.Append(defErr)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (req *StartTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}
	if req.CellId == "" {
		validationError = validationError.Append(ErrInvalidField{"cell_id"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func EnvironmentVariablesFromProto(envVars []*EnvironmentVariable) []oldmodels.EnvironmentVariable {
	if envVars == nil {
		return nil
	}
	out := make([]oldmodels.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func EnvironmentVariablesFromModel(envVars []oldmodels.EnvironmentVariable) []*EnvironmentVariable {
	if envVars == nil {
		return nil
	}
	out := make([]*EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i] = &EnvironmentVariable{
			Name:  val.Name,
			Value: val.Value,
		}
	}
	return out
}
