package models

type ReceptorPresence struct {
	ReceptorID  string `json:"id"`
	ReceptorURL string `json:"address"`
}

func NewReceptorPresence(id, url string) ReceptorPresence {
	return ReceptorPresence{
		ReceptorID:  id,
		ReceptorURL: url,
	}
}

func (r ReceptorPresence) Validate() error {
	var validationError ValidationError
	if r.ReceptorID == "" {
		validationError = validationError.Append(ErrInvalidField{"id"})
	}

	if r.ReceptorURL == "" {
		validationError = validationError.Append(ErrInvalidField{"address"})
	}
	if !validationError.Empty() {
		return validationError
	}
	return nil
}
