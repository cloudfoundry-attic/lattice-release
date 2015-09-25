package models

func (request *DesiredLRPsRequest) Validate() error {
	return nil
}

func (request *DesiredLRPByProcessGuidRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *DesireLRPRequest) Validate() error {
	var validationError ValidationError

	if request.DesiredLrp == nil {
		validationError = validationError.Append(ErrInvalidField{"desired_lrp"})
	} else if err := request.DesiredLrp.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *UpdateDesiredLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if request.Update != nil {
		if err := request.Update.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *RemoveDesiredLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
