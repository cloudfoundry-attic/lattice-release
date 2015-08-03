package models

import "bytes"

type Validator interface {
	Validate() error
}

type ValidationError []error

func (ve ValidationError) Append(err error) ValidationError {
	switch err := err.(type) {
	case ValidationError:
		return append(ve, err...)
	default:
		return append(ve, err)
	}
}

func (ve ValidationError) Error() string {
	var buffer bytes.Buffer

	for i, err := range ve {
		if err == nil {
			continue
		}
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(err.Error())
	}

	return buffer.String()
}

func (ve ValidationError) Empty() bool {
	return len(ve) == 0
}
