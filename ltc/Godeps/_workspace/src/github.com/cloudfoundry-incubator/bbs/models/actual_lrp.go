package models

import (
	"errors"
	"strings"
	"time"
)

const (
	ActualLRPStateUnclaimed = "UNCLAIMED"
	ActualLRPStateClaimed   = "CLAIMED"
	ActualLRPStateRunning   = "RUNNING"
	ActualLRPStateCrashed   = "CRASHED"
)

var ActualLRPStates = []string{
	ActualLRPStateUnclaimed,
	ActualLRPStateClaimed,
	ActualLRPStateRunning,
	ActualLRPStateCrashed,
}

type ActualLRPChange struct {
	Before *ActualLRPGroup
	After  *ActualLRPGroup
}

type ActualLRPFilter struct {
	Domain string
	CellID string
}

func NewActualLRPKey(processGuid string, index int32, domain string) ActualLRPKey {
	return ActualLRPKey{processGuid, index, domain}
}

func NewActualLRPInstanceKey(instanceGuid string, cellId string) ActualLRPInstanceKey {
	return ActualLRPInstanceKey{instanceGuid, cellId}
}

func NewActualLRPNetInfo(address string, ports ...*PortMapping) ActualLRPNetInfo {
	if ports == nil {
		ports = []*PortMapping{}
	}
	return ActualLRPNetInfo{address, ports}
}

func EmptyActualLRPNetInfo() ActualLRPNetInfo {
	return NewActualLRPNetInfo("")
}

func (info ActualLRPNetInfo) Empty() bool {
	return info.Address == "" && len(info.Ports) == 0
}

func NewPortMapping(hostPort, containerPort uint32) *PortMapping {
	return &PortMapping{
		HostPort:      hostPort,
		ContainerPort: containerPort,
	}
}

func (key ActualLRPInstanceKey) Empty() bool {
	return key.InstanceGuid == "" && key.CellId == ""
}

func (actual ActualLRP) ShouldRestartCrash(now time.Time, calc RestartCalculator) bool {
	if actual.State != ActualLRPStateCrashed {
		return false
	}

	return calc.ShouldRestart(now.UnixNano(), actual.Since, actual.CrashCount)
}

func (before ActualLRP) AllowsTransitionTo(lrpKey ActualLRPKey, instanceKey ActualLRPInstanceKey, newState string) bool {
	if !before.ActualLRPKey.Equal(&lrpKey) {
		return false
	}

	if before.State == ActualLRPStateClaimed && newState == ActualLRPStateRunning {
		return true
	}

	if (before.State == ActualLRPStateClaimed || before.State == ActualLRPStateRunning) &&
		(newState == ActualLRPStateClaimed || newState == ActualLRPStateRunning) &&
		(!before.ActualLRPInstanceKey.Equal(&instanceKey)) {
		return false
	}

	return true
}

func NewRunningActualLRPGroup(actualLRP *ActualLRP) *ActualLRPGroup {
	return &ActualLRPGroup{
		Instance: actualLRP,
	}
}

func (group ActualLRPGroup) Resolve() (*ActualLRP, bool) {
	if group.Instance == nil && group.Evacuating == nil {
		panic(ErrActualLRPGroupInvalid)
	}

	if group.Instance == nil {
		return group.Evacuating, true
	}

	if group.Evacuating == nil {
		return group.Instance, false
	}

	if group.Instance.State == ActualLRPStateRunning || group.Instance.State == ActualLRPStateCrashed {
		return group.Instance, false
	} else {
		return group.Evacuating, true
	}
}

func NewUnclaimedActualLRP(lrpKey ActualLRPKey, since int64) *ActualLRP {
	return &ActualLRP{
		ActualLRPKey: lrpKey,
		State:        ActualLRPStateUnclaimed,
		Since:        since,
	}
}

func NewClaimedActualLRP(lrpKey ActualLRPKey, instanceKey ActualLRPInstanceKey, since int64) *ActualLRP {
	return &ActualLRP{
		ActualLRPKey:         lrpKey,
		ActualLRPInstanceKey: instanceKey,
		State:                ActualLRPStateClaimed,
		Since:                since,
	}
}

func NewRunningActualLRP(lrpKey ActualLRPKey, instanceKey ActualLRPInstanceKey, netInfo ActualLRPNetInfo, since int64) *ActualLRP {
	return &ActualLRP{
		ActualLRPKey:         lrpKey,
		ActualLRPInstanceKey: instanceKey,
		ActualLRPNetInfo:     netInfo,
		State:                ActualLRPStateRunning,
		Since:                since,
	}
}

func (request StartActualLRPRequest) Validate() error {
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

func (request FailActualLRPRequest) Validate() error {
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

func (actual ActualLRP) Validate() error {
	var validationError ValidationError

	err := actual.ActualLRPKey.Validate()
	if err != nil {
		validationError = validationError.Append(err)
	}

	if actual.Since == 0 {
		validationError = validationError.Append(ErrInvalidField{"since"})
	}

	switch actual.State {
	case ActualLRPStateUnclaimed:
		if !actual.ActualLRPInstanceKey.Empty() {
			validationError = validationError.Append(errors.New("instance key cannot be set when state is unclaimed"))
		}
		if !actual.ActualLRPNetInfo.Empty() {
			validationError = validationError.Append(errors.New("net info cannot be set when state is unclaimed"))
		}

	case ActualLRPStateClaimed:
		if err := actual.ActualLRPInstanceKey.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
		if !actual.ActualLRPNetInfo.Empty() {
			validationError = validationError.Append(errors.New("net info cannot be set when state is claimed"))
		}
		if strings.TrimSpace(actual.PlacementError) != "" {
			validationError = validationError.Append(errors.New("placement error cannot be set when state is claimed"))
		}

	case ActualLRPStateRunning:
		if err := actual.ActualLRPInstanceKey.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
		if err := actual.ActualLRPNetInfo.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
		if strings.TrimSpace(actual.PlacementError) != "" {
			validationError = validationError.Append(errors.New("placement error cannot be set when state is running"))
		}

	case ActualLRPStateCrashed:
		if !actual.ActualLRPInstanceKey.Empty() {
			validationError = validationError.Append(errors.New("instance key cannot be set when state is crashed"))
		}
		if !actual.ActualLRPNetInfo.Empty() {
			validationError = validationError.Append(errors.New("net info cannot be set when state is crashed"))
		}
		if strings.TrimSpace(actual.PlacementError) != "" {
			validationError = validationError.Append(errors.New("placement error cannot be set when state is crashed"))
		}

	default:
		validationError = validationError.Append(ErrInvalidField{"state"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (key *ActualLRPKey) Validate() error {
	var validationError ValidationError

	if key.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if key.Index < 0 {
		validationError = validationError.Append(ErrInvalidField{"index"})
	}

	if key.Domain == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (key *ActualLRPNetInfo) Validate() error {
	var validationError ValidationError

	if key.Address == "" {
		return validationError.Append(ErrInvalidField{"address"})
	}

	return nil
}

func (key *ActualLRPInstanceKey) Validate() error {
	var validationError ValidationError

	if key.CellId == "" {
		validationError = validationError.Append(ErrInvalidField{"cell_id"})
	}

	if key.InstanceGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"instance_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
