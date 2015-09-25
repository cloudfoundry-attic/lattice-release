package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/format"
)

const (
	ActionTypeDownload     = "download"
	ActionTypeEmitProgress = "emit_progress"
	ActionTypeRun          = "run"
	ActionTypeUpload       = "upload"
	ActionTypeTimeout      = "timeout"
	ActionTypeTry          = "try"
	ActionTypeParallel     = "parallel"
	ActionTypeSerial       = "serial"
	ActionTypeCodependent  = "codependent"
)

var ErrInvalidActionType = errors.New("invalid action type")

type ActionInterface interface {
	ActionType() string
	Validate() error
}

func (*Action) Version() format.Version {
	return format.V0
}

func (*Action) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *Action) Validate() error {
	if a == nil {
		return nil
	}

	if inner := UnwrapAction(a); inner != nil {
		err := inner.Validate()
		if err != nil {
			return err
		}
	} else {
		return ErrInvalidField{"inner-action"}
	}
	return nil
}

func (*DownloadAction) Version() format.Version {
	return format.V0
}

func (*DownloadAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *DownloadAction) ActionType() string {
	return ActionTypeDownload
}

func (a DownloadAction) Validate() error {
	var validationError ValidationError

	if a.GetFrom() == "" {
		validationError = validationError.Append(ErrInvalidField{"from"})
	}

	if a.GetTo() == "" {
		validationError = validationError.Append(ErrInvalidField{"to"})
	}

	if a.GetUser() == "" {
		validationError = validationError.Append(ErrInvalidField{"user"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*UploadAction) Version() format.Version {
	return format.V0
}

func (*UploadAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *UploadAction) ActionType() string {
	return ActionTypeUpload
}

func (a UploadAction) Validate() error {
	var validationError ValidationError

	if a.GetTo() == "" {
		validationError = validationError.Append(ErrInvalidField{"to"})
	}

	if a.GetFrom() == "" {
		validationError = validationError.Append(ErrInvalidField{"from"})
	}

	if a.GetUser() == "" {
		validationError = validationError.Append(ErrInvalidField{"user"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*RunAction) Version() format.Version {
	return format.V0
}

func (*RunAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *RunAction) ActionType() string {
	return ActionTypeRun
}

func (a RunAction) Validate() error {
	var validationError ValidationError

	if a.Path == "" {
		validationError = validationError.Append(ErrInvalidField{"path"})
	}

	if a.User == "" {
		validationError = validationError.Append(ErrInvalidField{"user"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*TimeoutAction) Version() format.Version {
	return format.V0
}

func (*TimeoutAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *TimeoutAction) ActionType() string {
	return ActionTypeTimeout
}

func (a TimeoutAction) Validate() error {
	var validationError ValidationError

	if a.Action == nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
	} else {
		err := UnwrapAction(a.Action).Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if a.GetTimeout() <= 0 {
		validationError = validationError.Append(ErrInvalidField{"timeout"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*TryAction) Version() format.Version {
	return format.V0
}

func (*TryAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *TryAction) ActionType() string {
	return ActionTypeTry
}

func (a TryAction) Validate() error {
	var validationError ValidationError

	if a.Action == nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
	} else {
		err := UnwrapAction(a.Action).Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*ParallelAction) Version() format.Version {
	return format.V0
}

func (*ParallelAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *ParallelAction) ActionType() string {
	return ActionTypeParallel
}

func (a ParallelAction) Validate() error {
	var validationError ValidationError

	if a.Actions == nil || len(a.Actions) == 0 {
		validationError = validationError.Append(ErrInvalidField{"actions"})
	} else {
		for index, action := range a.Actions {
			if action == nil {
				errorString := fmt.Sprintf("action at index %d", index)
				validationError = validationError.Append(ErrInvalidField{errorString})
			} else {
				err := UnwrapAction(action).Validate()
				if err != nil {
					validationError = validationError.Append(err)
				}
			}
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*CodependentAction) Version() format.Version {
	return format.V0
}

func (*CodependentAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *CodependentAction) ActionType() string {
	return ActionTypeCodependent
}

func (a CodependentAction) Validate() error {
	var validationError ValidationError

	if a.Actions == nil || len(a.Actions) == 0 {
		validationError = validationError.Append(ErrInvalidField{"actions"})
	} else {
		for index, action := range a.Actions {
			if action == nil {
				errorString := fmt.Sprintf("action at index %d", index)
				validationError = validationError.Append(ErrInvalidField{errorString})
			} else {
				err := UnwrapAction(action).Validate()
				if err != nil {
					validationError = validationError.Append(err)
				}
			}
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*SerialAction) Version() format.Version {
	return format.V0
}

func (*SerialAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *SerialAction) ActionType() string {
	return ActionTypeSerial
}

func (a SerialAction) Validate() error {
	var validationError ValidationError

	if a.Actions == nil || len(a.Actions) == 0 {
		validationError = validationError.Append(ErrInvalidField{"actions"})
	} else {
		for index, action := range a.Actions {
			if action == nil {
				errorString := fmt.Sprintf("action at index %d", index)
				validationError = validationError.Append(ErrInvalidField{errorString})
			} else {
				err := UnwrapAction(action).Validate()
				if err != nil {
					validationError = validationError.Append(err)
				}
			}
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (*EmitProgressAction) Version() format.Version {
	return format.V0
}

func (*EmitProgressAction) MigrateFromVersion(v format.Version) error {
	return nil
}

func (a *EmitProgressAction) ActionType() string {
	return ActionTypeEmitProgress
}

func (a EmitProgressAction) Validate() error {
	var validationError ValidationError

	if a.Action == nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
	} else {
		err := UnwrapAction(a.Action).Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func EmitProgressFor(action ActionInterface, startMessage string, successMessage string, failureMessagePrefix string) *EmitProgressAction {
	return &EmitProgressAction{
		Action:               WrapAction(action),
		StartMessage:         startMessage,
		SuccessMessage:       successMessage,
		FailureMessagePrefix: failureMessagePrefix,
	}
}

func Timeout(action ActionInterface, timeout time.Duration) *TimeoutAction {
	return &TimeoutAction{
		Action:  WrapAction(action),
		Timeout: (int64)(timeout),
	}
}

func Try(action ActionInterface) *TryAction {
	return &TryAction{Action: WrapAction(action)}
}

func Parallel(actions ...ActionInterface) *ParallelAction {
	return &ParallelAction{Actions: WrapActions(actions)}
}

func Codependent(actions ...ActionInterface) *CodependentAction {
	return &CodependentAction{Actions: WrapActions(actions)}
}

func Serial(actions ...ActionInterface) *SerialAction {
	return &SerialAction{Actions: WrapActions(actions)}
}

func UnwrapAction(action *Action) ActionInterface {
	if action == nil {
		return nil
	}
	a := action.GetValue()
	if a == nil {
		return nil
	}
	return a.(ActionInterface)
}

func WrapActions(actions []ActionInterface) []*Action {
	wrappedActions := make([]*Action, 0, len(actions))
	for _, action := range actions {
		wrappedActions = append(wrappedActions, WrapAction(action))
	}
	return wrappedActions
}

func WrapAction(action ActionInterface) *Action {
	if action == nil {
		return nil
	}
	a := &Action{}
	a.SetValue(action)
	return a
}
