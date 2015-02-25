package models

import (
	"encoding/json"
	"regexp"
)

const maximumRouteLength = 4 * 1024

type DomainSet map[string]struct{}

func (set DomainSet) Add(domain string) {
	set[domain] = struct{}{}
}

func (set DomainSet) Each(predicate func(domain string)) {
	for domain := range set {
		predicate(domain)
	}
}

func (set DomainSet) Contains(domain string) bool {
	_, found := set[domain]
	return found
}

type DesiredLRPsByProcessGuid map[string]DesiredLRP

func (set DesiredLRPsByProcessGuid) Add(desired DesiredLRP) {
	set[desired.ProcessGuid] = desired
}

func (set DesiredLRPsByProcessGuid) Each(predicate func(desired DesiredLRP)) {
	for _, desired := range set {
		predicate(desired)
	}
}

type DesiredLRP struct {
	ProcessGuid          string                      `json:"process_guid"`
	Domain               string                      `json:"domain"`
	RootFSPath           string                      `json:"rootfs"`
	Instances            int                         `json:"instances"`
	Stack                string                      `json:"stack"`
	EnvironmentVariables []EnvironmentVariable       `json:"env,omitempty"`
	Setup                Action                      `json:"-"`
	Action               Action                      `json:"-"`
	StartTimeout         uint                        `json:"start_timeout"`
	Monitor              Action                      `json:"-"`
	DiskMB               int                         `json:"disk_mb"`
	MemoryMB             int                         `json:"memory_mb"`
	CPUWeight            uint                        `json:"cpu_weight"`
	Privileged           bool                        `json:"privileged"`
	Ports                []uint16                    `json:"ports"`
	Routes               map[string]*json.RawMessage `json:"routes,omitempty"`
	LogSource            string                      `json:"log_source"`
	LogGuid              string                      `json:"log_guid"`
	MetricsGuid          string                      `json:"metrics_guid"`
	Annotation           string                      `json:"annotation,omitempty"`
	EgressRules          []SecurityGroupRule         `json:"egress_rules,omitempty"`
	ModificationTag      ModificationTag             `json:"modification_tag"`
}

type InnerDesiredLRP DesiredLRP

type mDesiredLRP struct {
	SetupRaw   *json.RawMessage `json:"setup,omitempty"`
	ActionRaw  *json.RawMessage `json:"action,omitempty"`
	MonitorRaw *json.RawMessage `json:"monitor,omitempty"`
	*InnerDesiredLRP
}

type DesiredLRPUpdate struct {
	Instances  *int                        `json:"instances,omitempty"`
	Routes     map[string]*json.RawMessage `json:"routes,omitempty"`
	Annotation *string                     `json:"annotation,omitempty"`
}

type DesiredLRPChange struct {
	Before DesiredLRP
	After  DesiredLRP
}

func (desired DesiredLRP) ApplyUpdate(update DesiredLRPUpdate) DesiredLRP {
	if update.Instances != nil {
		desired.Instances = *update.Instances
	}
	if update.Routes != nil {
		desired.Routes = update.Routes
	}
	if update.Annotation != nil {
		desired.Annotation = *update.Annotation
	}
	return desired
}

var processGuidPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (desired DesiredLRP) Validate() error {
	var validationError ValidationError

	if desired.Domain == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if !processGuidPattern.MatchString(desired.ProcessGuid) {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if desired.Stack == "" {
		validationError = validationError.Append(ErrInvalidField{"stack"})
	}

	if desired.Setup != nil {
		err := desired.Setup.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if desired.Action == nil {
		validationError = validationError.Append(ErrInvalidActionType)
	} else {
		err := desired.Action.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if desired.Monitor != nil {
		err := desired.Monitor.Validate()
		if err != nil {
			validationError = validationError.Append(err)
		}
	}

	if desired.Instances < 0 {
		validationError = validationError.Append(ErrInvalidField{"instances"})
	}

	if desired.CPUWeight > 100 {
		validationError = validationError.Append(ErrInvalidField{"cpu_weight"})
	}

	if len(desired.Annotation) > maximumAnnotationLength {
		validationError = validationError.Append(ErrInvalidField{"annotation"})
	}

	totalRoutesLength := 0
	for _, value := range desired.Routes {
		totalRoutesLength += len(*value)
		if totalRoutesLength > maximumRouteLength {
			validationError = validationError.Append(ErrInvalidField{"routes"})
			break
		}
	}

	for _, rule := range desired.EgressRules {
		err := rule.Validate()
		if err != nil {
			validationError = validationError.Append(ErrInvalidField{"egress_rules"})
			validationError = validationError.Append(err)
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (desired *DesiredLRP) UnmarshalJSON(payload []byte) error {
	mLRP := &mDesiredLRP{InnerDesiredLRP: (*InnerDesiredLRP)(desired)}
	err := json.Unmarshal(payload, mLRP)
	if err != nil {
		return err
	}

	var a Action
	if mLRP.ActionRaw == nil {
		a = nil
	} else {
		a, err = UnmarshalAction(*mLRP.ActionRaw)
		if err != nil {
			return err
		}
	}
	desired.Action = a

	if mLRP.SetupRaw == nil {
		a = nil
	} else {
		a, err = UnmarshalAction(*mLRP.SetupRaw)
		if err != nil {
			return err
		}
		desired.Setup = a
	}

	if mLRP.MonitorRaw == nil {
		a = nil
	} else {
		a, err = UnmarshalAction(*mLRP.MonitorRaw)
		if err != nil {
			return err
		}
		desired.Monitor = a
	}

	return nil
}

func (desired DesiredLRP) MarshalJSON() ([]byte, error) {
	var setupRaw, actionRaw, monitorRaw *json.RawMessage

	if desired.Action != nil {
		raw, err := MarshalAction(desired.Action)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		actionRaw = &rm
	}

	if desired.Setup != nil {
		raw, err := MarshalAction(desired.Setup)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		setupRaw = &rm
	}
	if desired.Monitor != nil {
		raw, err := MarshalAction(desired.Monitor)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		monitorRaw = &rm
	}

	innerDesiredLRP := InnerDesiredLRP(desired)

	mLRP := &mDesiredLRP{
		SetupRaw:        setupRaw,
		ActionRaw:       actionRaw,
		MonitorRaw:      monitorRaw,
		InnerDesiredLRP: &innerDesiredLRP,
	}

	return json.Marshal(mLRP)
}
