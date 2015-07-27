package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var ErrInvalidActionType = errors.New("invalid action type")

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

type Action interface {
	ActionType() string
	Validator
}

type DownloadAction struct {
	Artifact string `json:"artifact,omitempty"`
	From     string `json:"from"`
	To       string `json:"to"`
	CacheKey string `json:"cache_key"`

	User string `json:"user"`

	LogSource string `json:"log_source,omitempty"`
}

func (a *DownloadAction) ActionType() string {
	return ActionTypeDownload
}

func (a DownloadAction) Validate() error {
	var validationError ValidationError

	if a.From == "" {
		validationError = validationError.Append(ErrInvalidField{"from"})
	}

	if a.To == "" {
		validationError = validationError.Append(ErrInvalidField{"to"})
	}

	if a.User == "" {
		validationError = validationError.Append(ErrInvalidField{"user"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

type UploadAction struct {
	Artifact string `json:"artifact,omitempty"`
	To       string `json:"to"`
	From     string `json:"from"`

	User string `json:"user"`

	LogSource string `json:"log_source,omitempty"`
}

func (a *UploadAction) ActionType() string {
	return ActionTypeUpload
}

func (a UploadAction) Validate() error {
	var validationError ValidationError

	if a.To == "" {
		validationError = validationError.Append(ErrInvalidField{"to"})
	}

	if a.From == "" {
		validationError = validationError.Append(ErrInvalidField{"from"})
	}

	if a.User == "" {
		validationError = validationError.Append(ErrInvalidField{"user"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

type RunAction struct {
	Path           string                `json:"path"`
	Args           []string              `json:"args,omitempty"`
	Dir            string                `json:"dir,omitempty"`
	Env            []EnvironmentVariable `json:"env,omitempty"`
	ResourceLimits ResourceLimits        `json:"resource_limits"`
	User           string                `json:"user"`

	LogSource string `json:"log_source,omitempty"`
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

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResourceLimits struct {
	Nofile *uint64 `json:"nofile,omitempty"`
}

type TimeoutAction struct {
	Action  Action
	Timeout time.Duration

	LogSource string
}

func (a *TimeoutAction) ActionType() string {
	return ActionTypeTimeout
}

func (a TimeoutAction) Validate() error {
	var validationError ValidationError

	if a.Action == nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
	} else {
		err := a.Action.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if a.Timeout <= 0 {
		validationError = validationError.Append(ErrInvalidField{"timeout"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (a *TimeoutAction) MarshalJSON() ([]byte, error) {
	bytes, err := MarshalAction(a.Action)
	if err != nil {
		return nil, err
	}
	j := json.RawMessage(bytes)

	return json.Marshal(&mTimeoutAction{
		Action:    &j,
		Timeout:   a.Timeout,
		LogSource: a.LogSource,
	})
}

func (a *TimeoutAction) UnmarshalJSON(data []byte) error {
	m := mTimeoutAction{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	var action Action
	if m.Action == nil {
		action = nil
	} else {
		action, err = UnmarshalAction([]byte(*m.Action))
		if err != nil {
			return err
		}
	}

	a.Action = action
	a.Timeout = m.Timeout
	a.LogSource = m.LogSource

	return nil
}

type mTimeoutAction struct {
	Action  *json.RawMessage `json:"action"`
	Timeout time.Duration    `json:"timeout"`

	LogSource string `json:"log_source,omitempty"`
}

type TryAction struct {
	Action

	LogSource string
}

func (a *TryAction) ActionType() string {
	return ActionTypeTry
}

func (a TryAction) Validate() error {
	var validationError ValidationError

	if a.Action == nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
	} else {
		err := a.Action.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (a *TryAction) MarshalJSON() ([]byte, error) {
	bytes, err := MarshalAction(a.Action)
	if err != nil {
		return nil, err
	}
	j := json.RawMessage(bytes)

	return json.Marshal(&mTryAction{
		Action:    &j,
		LogSource: a.LogSource,
	})
}

func (a *TryAction) UnmarshalJSON(data []byte) error {
	m := mTryAction{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	var action Action
	if m.Action == nil {
		action = nil
	} else {
		action, err = UnmarshalAction([]byte(*m.Action))
		if err != nil {
			return err
		}
	}

	a.Action = action
	a.LogSource = m.LogSource

	return nil
}

type mTryAction struct {
	Action *json.RawMessage `json:"action"`

	LogSource string `json:"log_source,omitempty"`
}

type ParallelAction struct {
	Actions []Action

	LogSource string
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
				err := action.Validate()
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

func (a *ParallelAction) MarshalJSON() ([]byte, error) {
	mActions, err := marshalActions(a.Actions)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&mParallelAction{
		Actions:   mActions,
		LogSource: a.LogSource,
	})
}

func (a *ParallelAction) UnmarshalJSON(data []byte) error {
	m := mParallelAction{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	actions, err := unmarshalActions(m.Actions)
	if err != nil {
		return err
	}

	a.Actions = actions
	a.LogSource = m.LogSource

	return nil
}

type mParallelAction struct {
	Actions []*json.RawMessage `json:"actions"`

	LogSource string `json:"log_source,omitempty"`
}

type CodependentAction struct {
	Actions []Action

	LogSource string
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
				err := action.Validate()
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

func (a *CodependentAction) MarshalJSON() ([]byte, error) {
	mActions, err := marshalActions(a.Actions)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&mCodependentAction{
		Actions:   mActions,
		LogSource: a.LogSource,
	})
}

func (a *CodependentAction) UnmarshalJSON(data []byte) error {
	m := mCodependentAction{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	actions, err := unmarshalActions(m.Actions)
	if err != nil {
		return err
	}

	a.Actions = actions
	a.LogSource = m.LogSource

	return nil
}

type mCodependentAction struct {
	Actions []*json.RawMessage `json:"actions"`

	LogSource string `json:"log_source,omitempty"`
}

type SerialAction struct {
	Actions []Action

	LogSource string
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
				err := action.Validate()
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

func (a *SerialAction) MarshalJSON() ([]byte, error) {
	mActions, err := marshalActions(a.Actions)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&mSerialAction{
		Actions:   mActions,
		LogSource: a.LogSource,
	})
}

func (a *SerialAction) UnmarshalJSON(data []byte) error {
	m := mSerialAction{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	actions, err := unmarshalActions(m.Actions)
	if err != nil {
		return err
	}

	a.Actions = actions
	a.LogSource = m.LogSource

	return nil
}

type mSerialAction struct {
	Actions []*json.RawMessage `json:"actions"`

	LogSource string `json:"log_source,omitempty"`
}

type EmitProgressAction struct {
	Action               Action `json:"-"`
	StartMessage         string `json:"start_message"`
	SuccessMessage       string `json:"success_message"`
	FailureMessagePrefix string `json:"failure_message_prefix"`

	LogSource string `json:"log_source,omitempty"`
}

type InnerEmitProgressAction EmitProgressAction

type mEmitProgressAction struct {
	*InnerEmitProgressAction
	ActionRaw *json.RawMessage `json:"action"`
}

func (a *EmitProgressAction) ActionType() string {
	return ActionTypeEmitProgress
}

func (a EmitProgressAction) Validate() error {
	var validationError ValidationError

	if a.Action == nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
	} else {
		err := a.Action.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (a *EmitProgressAction) MarshalJSON() ([]byte, error) {
	bytes, err := MarshalAction(a.Action)
	if err != nil {
		return nil, err
	}
	j := json.RawMessage(bytes)

	return json.Marshal(&mEmitProgressAction{
		InnerEmitProgressAction: (*InnerEmitProgressAction)(a),
		ActionRaw:               &j,
	})
}

func (a *EmitProgressAction) UnmarshalJSON(data []byte) error {
	m := &mEmitProgressAction{InnerEmitProgressAction: (*InnerEmitProgressAction)(a)}
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	var action Action
	if m.ActionRaw == nil {
		action = nil
	} else {
		action, err = UnmarshalAction(*m.ActionRaw)
		if err != nil {
			return err
		}
	}
	a.Action = action

	return nil
}

func EmitProgressFor(action Action, startMessage string, successMessage string, failureMessagePrefix string) *EmitProgressAction {
	return &EmitProgressAction{
		Action:               action,
		StartMessage:         startMessage,
		SuccessMessage:       successMessage,
		FailureMessagePrefix: failureMessagePrefix,
	}
}

func Timeout(action Action, timeout time.Duration) *TimeoutAction {
	return &TimeoutAction{
		Action:  action,
		Timeout: timeout,
	}
}

func Try(action Action) *TryAction {
	return &TryAction{
		Action: action,
	}
}

func Parallel(actions ...Action) *ParallelAction {
	return &ParallelAction{
		Actions: actions,
	}
}

func Codependent(actions ...Action) *CodependentAction {
	return &CodependentAction{
		Actions: actions,
	}
}

func Serial(actions ...Action) *SerialAction {
	return &SerialAction{
		Actions: actions,
	}
}

var actionMap = map[string]Action{
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

func marshalActions(actions []Action) ([]*json.RawMessage, error) {
	if actions == nil {
		return nil, nil
	}

	mActions := make([]*json.RawMessage, len(actions))
	for i, action := range actions {
		bytes, err := MarshalAction(action)
		if err != nil {
			return nil, err
		}
		j := json.RawMessage(bytes)

		mActions[i] = &j
	}

	return mActions, nil
}

func MarshalAction(a Action) ([]byte, error) {
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

func unmarshalActions(mActions []*json.RawMessage) ([]Action, error) {
	if mActions == nil {
		return nil, nil
	}

	actions := make([]Action, len(mActions))
	for i, mAction := range mActions {
		if mAction == nil {
			actions[i] = nil
		} else {
			action, err := UnmarshalAction([]byte(*mAction))
			if err != nil {
				return nil, err
			}
			actions[i] = action
		}
	}

	return actions, nil
}

func UnmarshalAction(data []byte) (Action, error) {
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
			return p.Interface().(Action), err
		}
	}

	return nil, ErrInvalidField{"Invalid action"}
}
