package models

import (
	"errors"
	"fmt"
)

func NewError(errType Error_Type, msg string) *Error {
	return &Error{
		Type:    errType,
		Message: msg,
	}
}

func ConvertError(err error) *Error {
	if err == nil {
		return nil
	}

	modelErr, ok := err.(*Error)
	if !ok {
		modelErr = NewError(Error_UnknownError, err.Error())
	}
	return modelErr
}

func (err *Error) ToError() error {
	if err == nil {
		return nil
	}
	return err
}

func (err *Error) Equal(other error) bool {
	if e, ok := other.(*Error); ok {
		if err == nil && e != nil {
			return false
		}
		return e.GetType() == err.GetType()
	}
	return false
}

func (err *Error) Error() string {
	return err.GetMessage()
}

var (
	ErrResourceNotFound = &Error{
		Type:    Error_ResourceNotFound,
		Message: "the requested resource could not be found",
	}

	ErrResourceExists = &Error{
		Type:    Error_ResourceExists,
		Message: "the requested resource already exists",
	}

	ErrResourceConflict = &Error{
		Type:    Error_ResourceConflict,
		Message: "the requested resource is in a conflicting state",
	}

	ErrBadRequest = &Error{
		Type:    Error_InvalidRequest,
		Message: "the request received is invalid",
	}

	ErrUnknownError = &Error{
		Type:    Error_UnknownError,
		Message: "the request failed for an unknown reason",
	}

	ErrSerializeJSON = &Error{
		Type:    Error_InvalidJSON,
		Message: "could not serialize JSON",
	}

	ErrDeserializeJSON = &Error{
		Type:    Error_InvalidJSON,
		Message: "could not deserialize JSON",
	}

	ErrFailedToOpenEnvelope = &Error{
		Type:    Error_FailedToOpenEnvelope,
		Message: "could not open envelope",
	}

	ErrActualLRPCannotBeClaimed = &Error{
		Type:    Error_ActualLRPCannotBeClaimed,
		Message: "cannot claim actual LRP",
	}

	ErrActualLRPCannotBeStarted = &Error{
		Type:    Error_ActualLRPCannotBeStarted,
		Message: "cannot start actual LRP",
	}

	ErrActualLRPCannotBeCrashed = &Error{
		Type:    Error_ActualLRPCannotBeCrashed,
		Message: "cannot crash actual LRP",
	}

	ErrActualLRPCannotBeFailed = &Error{
		Type:    Error_ActualLRPCannotBeFailed,
		Message: "cannot fail actual LRP",
	}

	ErrActualLRPCannotBeRemoved = &Error{
		Type:    Error_ActualLRPCannotBeRemoved,
		Message: "cannot remove actual LRP",
	}

	ErrActualLRPCannotBeStopped = &Error{
		Type:    Error_ActualLRPCannotBeStopped,
		Message: "cannot stop actual LRP",
	}

	ErrActualLRPCannotBeUnclaimed = &Error{
		Type:    Error_ActualLRPCannotBeUnclaimed,
		Message: "cannot unclaim actual LRP",
	}

	ErrActualLRPCannotBeEvacuated = &Error{
		Type:    Error_ActualLRPCannotBeEvacuated,
		Message: "cannot evacuate actual LRP",
	}

	ErrDesiredLRPCannotBeUpdated = &Error{
		Type:    Error_DesiredLRPCannotBeUpdated,
		Message: "cannot update desired LRP",
	}
)

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

func NewTaskTransitionError(from, to Task_State) *Error {
	return &Error{
		Type:    Error_InvalidStateTransition,
		Message: fmt.Sprintf("Cannot transition from %s to %s", from.String(), to.String()),
	}
}

func NewRunningOnDifferentCellError(expectedCellId, actualCellId string) *Error {
	return &Error{
		Type:    Error_RunningOnDifferentCell,
		Message: fmt.Sprintf("Running on cell %s not %s", actualCellId, expectedCellId),
	}
}
