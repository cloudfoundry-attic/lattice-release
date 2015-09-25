package models

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

func (req *CompleteTaskRequest) Validate() error {
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

func (req *FailTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}
	if req.FailureReason == "" {
		validationError = validationError.Append(ErrInvalidField{"failure_reason"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (req *TasksRequest) Validate() error {
	return nil
}

func (request *TaskByGuidRequest) Validate() error {
	var validationError ValidationError

	if request.TaskGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *TaskGuidRequest) Validate() error {
	var validationError ValidationError

	if request.TaskGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ConvergeTasksRequest) Validate() error {
	return nil
}
