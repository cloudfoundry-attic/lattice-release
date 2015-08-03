package models

import "errors"

func (err Error) Error() string {
	return err.GetMessage()
}

const (
	InvalidDomain = "InvalidDomain"

	InvalidRecord          = "InvalidRecord"
	InvalidRequest         = "InvalidRequest"
	InvalidResponse        = "InvalidResponse"
	InvalidProtobufMessage = "InvalidProtobufMessage"
	InvalidJSON            = "InvalidJSON"

	UnknownError = "UnknownError"
	Unauthorized = "Unauthorized"

	ResourceConflict = "ResourceConflict"
	ResourceNotFound = "ResourceNotFound"
	RouterError      = "RouterError"

	ActualLRPCannotBeClaimed = "ActualLRPCannotBeClaimed"
	ActualLRPCannotBeStarted = "ActualLRPCannotBeStarted"
	ActualLRPCannotBeFailed  = "ActualLRPCannotBeFailed"
	ActualLRPCannotBeRemoved = "ActualLRPCannotBeRemoved"
)

var (
	ErrResourceNotFound = &Error{
		Type:    ResourceNotFound,
		Message: "the requested resource could not be found",
	}

	ErrBadRequest = &Error{
		Type:    InvalidRequest,
		Message: "the request received is invalid",
	}

	ErrUnknownError = &Error{
		Type:    UnknownError,
		Message: "the request failed for an unknown reason",
	}

	ErrSerializeJSON = &Error{
		Type:    InvalidJSON,
		Message: "could not serialize JSON",
	}

	ErrDeserializeJSON = &Error{
		Type:    InvalidJSON,
		Message: "could not deserialize JSON",
	}

	ErrActualLRPCannotBeClaimed = &Error{
		Type:    ActualLRPCannotBeClaimed,
		Message: "cannot claim actual LRP",
	}

	ErrActualLRPCannotBeStarted = &Error{
		Type:    ActualLRPCannotBeStarted,
		Message: "cannot start actual LRP",
	}

	ErrActualLRPCannotBeFailed = &Error{
		Type:    ActualLRPCannotBeFailed,
		Message: "cannot fail actual LRP",
	}

	ErrActualLRPCannotBeRemoved = &Error{
		Type:    ActualLRPCannotBeRemoved,
		Message: "cannot remove actual LRP",
	}
)

func (err *Error) Equal(other *Error) bool {
	return err.GetType() == other.GetType()
}

type ErrInvalidField struct {
	Field string
}

func (err ErrInvalidField) Error() string {
	return "Invalid field: " + err.Field
}

type ErrInvalidModification struct {
	InvalidField string
}

func (err ErrInvalidModification) Error() string {
	return "attempt to make invalid change to field: " + err.InvalidField
}

var ErrActualLRPGroupInvalid = errors.New("ActualLRPGroup invalid")
