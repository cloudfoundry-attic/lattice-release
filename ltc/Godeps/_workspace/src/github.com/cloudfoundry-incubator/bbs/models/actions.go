package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
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
	Validator
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

func (a *ParallelAction) ActionType() string {
	return ActionTypeParallel
}

func (a ParallelAction) Validate() error {
	var validationError ValidationError

	if a.Actions == nil {
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

func (a *CodependentAction) ActionType() string {
	return ActionTypeCodependent
}

func (a CodependentAction) Validate() error {
	var validationError ValidationError

	if a.Actions == nil {
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

func (a *SerialAction) ActionType() string {
	return ActionTypeSerial
}

func (a SerialAction) Validate() error {
	var validationError ValidationError

	if a.Actions == nil {
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

func EmitProgressFor(action interface{}, startMessage string, successMessage string, failureMessagePrefix string) *Action {
	return WrapAction(&EmitProgressAction{
		Action:               WrapAction(action),
		StartMessage:         startMessage,
		SuccessMessage:       successMessage,
		FailureMessagePrefix: failureMessagePrefix,
	})
}

func Timeout(action interface{}, timeout time.Duration) *Action {
	return WrapAction(&TimeoutAction{
		Action:  WrapAction(action),
		Timeout: (int64)(timeout),
	})
}

func Try(action interface{}) *Action {
	return WrapAction(&TryAction{Action: WrapAction(action)})
}

func Parallel(actions ...interface{}) *Action {
	return WrapAction(&ParallelAction{Actions: WrapActions(actions)})
}

func Codependent(actions ...interface{}) *Action {
	return WrapAction(&CodependentAction{Actions: WrapActions(actions)})
}

func Serial(actions ...interface{}) *Action {
	return WrapAction(&SerialAction{Actions: WrapActions(actions)})
}

func UnwrapAction(action interface{}) ActionInterface {
	if action == nil {
		return nil
	}

	if action, ok := action.(*Action); ok {
		return action.GetValue().(ActionInterface)
	}

	if action, ok := action.(ActionInterface); ok {
		return action
	}

	return nil
}

func WrapActions(actions []interface{}) []*Action {
	wrappedActions := make([]*Action, len(actions))
	for i, action := range actions {
		wrappedActions[i] = WrapAction(action)
	}
	return wrappedActions
}

func WrapAction(action interface{}) *Action {
	if action, ok := action.(*Action); ok {
		return action
	}

	if action, ok := action.(ActionInterface); ok {
		a := &Action{}
		a.SetValue(action)
		return a
	}

	return nil
}

var actionMap = map[string]ActionInterface{
	ActionTypeDownload:     &DownloadAction{},
	ActionTypeEmitProgress: &EmitProgressAction{},
	ActionTypeRun:          &RunAction{},
	ActionTypeUpload:       &UploadAction{},
	ActionTypeTimeout:      &TimeoutAction{},
	ActionTypeTry:          &TryAction{},
	ActionTypeParallel:     &ParallelAction{},
	ActionTypeSerial:       &SerialAction{},
	ActionTypeCodependent:  &CodependentAction{},
}

func MarshalAction(a ActionInterface) ([]byte, error) {
	if a == nil {
		return json.Marshal(a)
	}
	payload, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	j := json.RawMessage(payload)

	wrapped := map[string]*json.RawMessage{
		a.ActionType(): &j,
	}

	return json.Marshal(wrapped)
}

func UnmarshalAction(data []byte) (ActionInterface, error) {
	wrapped := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &wrapped)
	if err != nil {
		return nil, err
	}
	if wrapped == nil {
		return nil, nil
	}

	if len(wrapped) == 1 {
		for k, v := range wrapped {
			action := actionMap[k]
			if action == nil {
				return nil, errors.New("Unknown action: " + string(k))
			}
			st := reflect.TypeOf(action).Elem()
			p := reflect.New(st)
			err = json.Unmarshal(v, p.Interface())
			return p.Interface().(ActionInterface), err
		}
	}

	return nil, ErrInvalidField{"Invalid action"}
}
