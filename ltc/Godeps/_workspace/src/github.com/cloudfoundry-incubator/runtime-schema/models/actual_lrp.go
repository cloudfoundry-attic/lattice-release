package models

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type ActualLRPsByIndex map[int]ActualLRP

func (actuals ActualLRPsByIndex) Slice() []ActualLRP {
	aSlice := make([]ActualLRP, 0, len(actuals))
	for _, actual := range actuals {
		aSlice = append(aSlice, actual)
	}
	return aSlice
}

type ActualLRPsByProcessGuidAndIndex map[string]ActualLRPsByIndex

func (set ActualLRPsByProcessGuidAndIndex) Add(actual ActualLRP) {
	actuals, found := set[actual.ProcessGuid]
	if !found {
		actuals = ActualLRPsByIndex{}
		set[actual.ProcessGuid] = actuals
	}

	actuals[actual.Index] = actual
}

func (set ActualLRPsByProcessGuidAndIndex) Each(predicate func(actual ActualLRP)) {
	for _, indexSet := range set {
		for _, actual := range indexSet {
			predicate(actual)
		}
	}
}

var ErrActualLRPGroupInvalid = errors.New("ActualLRPGroup invalid")

type ActualLRPGroup struct {
	Instance   *ActualLRP
	Evacuating *ActualLRP
}

func (group ActualLRPGroup) Resolve() (*ActualLRP, bool, error) {
	if group.Instance == nil && group.Evacuating == nil {
		return nil, false, ErrActualLRPGroupInvalid
	}

	if group.Instance == nil {
		return group.Evacuating, true, nil
	}

	if group.Evacuating == nil {
		return group.Instance, false, nil
	}

	if group.Instance.State == ActualLRPStateRunning || group.Instance.State == ActualLRPStateCrashed {
		return group.Instance, false, nil
	} else {
		return group.Evacuating, true, nil
	}
}

type ActualLRPGroupsByIndex map[int]ActualLRPGroup

func (actuals ActualLRPGroupsByIndex) Slice() []ActualLRPGroup {
	aSlice := make([]ActualLRPGroup, 0, len(actuals))
	for _, actual := range actuals {
		aSlice = append(aSlice, actual)
	}
	return aSlice
}

type ActualLRPState string

const (
	ActualLRPStateUnclaimed ActualLRPState = "UNCLAIMED"
	ActualLRPStateClaimed   ActualLRPState = "CLAIMED"
	ActualLRPStateRunning   ActualLRPState = "RUNNING"
	ActualLRPStateCrashed   ActualLRPState = "CRASHED"
)

var ActualLRPStates = []ActualLRPState{
	ActualLRPStateUnclaimed,
	ActualLRPStateClaimed,
	ActualLRPStateRunning,
	ActualLRPStateCrashed,
}

type ActualLRPKey struct {
	ProcessGuid string `json:"process_guid"`
	Index       int    `json:"index"`
	Domain      string `json:"domain"`
}

func NewActualLRPKey(processGuid string, index int, domain string) ActualLRPKey {
	return ActualLRPKey{
		ProcessGuid: processGuid,
		Index:       index,
		Domain:      domain,
	}
}

func (key ActualLRPKey) Validate() error {
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

type ActualLRPInstanceKey struct {
	InstanceGuid string `json:"instance_guid"`
	CellID       string `json:"cell_id"`
}

var emptyActualLRPInstanceKey = ActualLRPInstanceKey{}

func (key *ActualLRPInstanceKey) Empty() bool {
	return *key == emptyActualLRPInstanceKey
}

func NewActualLRPInstanceKey(instanceGuid string, cellID string) ActualLRPInstanceKey {
	return ActualLRPInstanceKey{
		InstanceGuid: instanceGuid,
		CellID:       cellID,
	}
}

func (key ActualLRPInstanceKey) Validate() error {
	var validationError ValidationError

	if key.CellID == "" {
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

type ActualLRPNetInfo struct {
	Address string        `json:"address"`
	Ports   []PortMapping `json:"ports"`
}

func NewActualLRPNetInfo(address string, ports []PortMapping) ActualLRPNetInfo {
	return ActualLRPNetInfo{
		Address: address,
		Ports:   ports,
	}
}

func EmptyActualLRPNetInfo() ActualLRPNetInfo {
	return NewActualLRPNetInfo("", []PortMapping{})
}

func (info *ActualLRPNetInfo) Empty() bool {
	return info.Address == "" && len(info.Ports) == 0
}

func (key ActualLRPNetInfo) Validate() error {
	var validationError ValidationError

	if key.Address == "" {
		return validationError.Append(ErrInvalidField{"address"})
	}

	return nil
}

const DefaultImmediateRestarts = 3
const DefaultMaxBackoffDuration = 16 * time.Minute
const DefaultMaxRestarts = 200

const CrashBackoffMinDuration = 30 * time.Second

func exponentialBackoff(exponent, max int) time.Duration {
	if exponent > max {
		exponent = max
	}
	return CrashBackoffMinDuration * time.Duration(powerOfTwo(exponent))
}

func powerOfTwo(pow int) int64 {
	if pow < 0 {
		panic("pow cannot be negative")
	}
	return 1 << uint(pow)
}

func calculateMaxBackoffCount(maxDuration time.Duration) int {
	total := math.Ceil(float64(maxDuration) / float64(CrashBackoffMinDuration))
	return int(math.Logb(total))
}

type RestartCalculator struct {
	ImmediateRestarts  int           `json:"immediate_restarts"`
	MaxBackoffCount    int           `json:"max_backoff_count"`
	MaxBackoffDuration time.Duration `json:"max_backoff_duration"`
	MaxRestartAttempts int           `json:"max_restart_attempts"`
}

func NewDefaultRestartCalculator() RestartCalculator {
	return NewRestartCalculator(DefaultImmediateRestarts, DefaultMaxBackoffDuration, DefaultMaxRestarts)
}

func NewRestartCalculator(immediateRestarts int, maxBackoffDuration time.Duration, maxRestarts int) RestartCalculator {
	return RestartCalculator{
		ImmediateRestarts:  immediateRestarts,
		MaxBackoffDuration: maxBackoffDuration,
		MaxBackoffCount:    calculateMaxBackoffCount(maxBackoffDuration),
		MaxRestartAttempts: maxRestarts,
	}
}

func (r RestartCalculator) Validate() error {
	var validationError ValidationError
	if r.MaxBackoffDuration < CrashBackoffMinDuration {
		err := fmt.Errorf("MaxBackoffDuration '%s' must be larger than CrashBackoffMinDuration '%s'", r.MaxBackoffDuration, CrashBackoffMinDuration)
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (r RestartCalculator) ShouldRestart(now, crashedAt int64, crashCount int) bool {
	switch {
	case crashCount < r.ImmediateRestarts:
		return true

	case crashCount < r.MaxRestartAttempts:
		backoffDuration := exponentialBackoff(crashCount-r.ImmediateRestarts, r.MaxBackoffCount)
		if backoffDuration > r.MaxBackoffDuration {
			backoffDuration = r.MaxBackoffDuration
		}
		nextRestartTime := crashedAt + backoffDuration.Nanoseconds()
		if nextRestartTime <= now {
			return true
		}
	}

	return false
}

type ActualLRP struct {
	ActualLRPKey
	ActualLRPInstanceKey
	ActualLRPNetInfo
	CrashCount      int             `json:"crash_count"`
	State           ActualLRPState  `json:"state"`
	PlacementError  string          `json:"placement_error,omitempty"`
	Since           int64           `json:"since"`
	ModificationTag ModificationTag `json:"modification_tag"`
}

type ModificationTag struct {
	Epoch string `json:"epoch"`
	Index uint   `json:"index"`
}

func (t *ModificationTag) Increment() {
	t.Index++
}

type ActualLRPChange struct {
	Before ActualLRP
	After  ActualLRP
}

const StaleUnclaimedActualLRPDuration = 30 * time.Second

func (actual ActualLRP) ShouldStartUnclaimed(now time.Time) bool {
	if actual.State != ActualLRPStateUnclaimed {
		return false
	}

	if now.Sub(time.Unix(0, actual.Since)) > StaleUnclaimedActualLRPDuration {
		return true
	}

	return false
}

func (actual ActualLRP) CellIsMissing(cellSet CellSet) bool {
	if actual.State == ActualLRPStateUnclaimed ||
		actual.State == ActualLRPStateCrashed {
		return false
	}

	return !cellSet.HasCellID(actual.CellID)
}

func (actual ActualLRP) ShouldRestartImmediately(calc RestartCalculator) bool {
	if actual.State != ActualLRPStateCrashed {
		return false
	}

	return calc.ShouldRestart(0, 0, actual.CrashCount)
}

func (actual ActualLRP) ShouldRestartCrash(now time.Time, calc RestartCalculator) bool {
	if actual.State != ActualLRPStateCrashed {
		return false
	}

	return calc.ShouldRestart(now.UnixNano(), actual.Since, actual.CrashCount)
}

func (before ActualLRP) AllowsTransitionTo(lrpKey ActualLRPKey, instanceKey ActualLRPInstanceKey, newState ActualLRPState) bool {
	if before.ActualLRPKey != lrpKey {
		return false
	}

	if before.State == ActualLRPStateClaimed && newState == ActualLRPStateRunning {
		return true
	}

	if (before.State == ActualLRPStateClaimed || before.State == ActualLRPStateRunning) &&
		(newState == ActualLRPStateClaimed || newState == ActualLRPStateRunning) &&
		(before.ActualLRPInstanceKey != instanceKey) {
		return false
	}

	return true
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
