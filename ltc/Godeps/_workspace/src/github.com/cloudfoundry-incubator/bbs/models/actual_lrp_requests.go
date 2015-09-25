package models

func (request *StartActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpNetInfo == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_net_info"})
	} else if err := request.ActualLrpNetInfo.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ClaimActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *CrashActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *FailActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ErrorMessage == "" {
		validationError = validationError.Append(ErrInvalidField{"error_message"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *RetireActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *RemoveEvacuatingActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *EvacuateClaimedActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *EvacuateCrashedActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ErrorMessage == "" {
		validationError = validationError.Append(ErrInvalidField{"error_message"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *EvacuateStoppedActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *EvacuateRunningActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpNetInfo == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_net_info"})
	} else if err := request.ActualLrpNetInfo.Validate(); err != nil {
		validationError = validationError.Append(err)
	}
	if !validationError.Empty() {
		return validationError
	}

	return nil
}
