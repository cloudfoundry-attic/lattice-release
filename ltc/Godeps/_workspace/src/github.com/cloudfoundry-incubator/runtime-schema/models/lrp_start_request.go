package models

type LRPStartRequest struct {
	DesiredLRP DesiredLRP `json:"desired_lrp"`
	Indices    []uint     `json:"indices"`
}

func NewLRPStartRequest(d DesiredLRP, indices ...uint) LRPStartRequest {
	return LRPStartRequest{
		DesiredLRP: d,
		Indices:    indices,
	}
}

func (lrpstart LRPStartRequest) Validate() error {
	var validationError ValidationError

	err := lrpstart.DesiredLRP.Validate()
	if err != nil {
		validationError = validationError.Append(err)
	}

	if len(lrpstart.Indices) == 0 {
		validationError = validationError.Append(ErrInvalidField{"indices"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
