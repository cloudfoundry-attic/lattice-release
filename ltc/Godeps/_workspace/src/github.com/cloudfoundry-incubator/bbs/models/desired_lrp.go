package models

import (
	"net/url"
	"regexp"

	"github.com/cloudfoundry-incubator/bbs/format"
)

const PreloadedRootFSScheme = "preloaded"

var processGuidPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type DesiredLRPChange struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

type DesiredLRPFilter struct {
	Domain string
}

func PreloadedRootFS(stack string) string {
	return (&url.URL{
		Scheme: PreloadedRootFSScheme,
		Opaque: stack,
	}).String()
}

func (*DesiredLRP) Version() format.Version {
	return format.V0
}

func (*DesiredLRP) MigrateFromVersion(v format.Version) error {
	return nil
}

func (desired *DesiredLRP) ApplyUpdate(update *DesiredLRPUpdate) *DesiredLRP {
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

func (desired DesiredLRP) Validate() error {
	var validationError ValidationError

	if desired.GetDomain() == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if !processGuidPattern.MatchString(desired.GetProcessGuid()) {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if desired.GetRootFs() == "" {
		validationError = validationError.Append(ErrInvalidField{"rootfs"})
	}

	rootFSURL, err := url.Parse(desired.GetRootFs())
	if err != nil || rootFSURL.Scheme == "" {
		validationError = validationError.Append(ErrInvalidField{"rootfs"})
	}

	if desired.Setup != nil {
		if err := desired.Setup.Validate(); err != nil {
			validationError = validationError.Append(ErrInvalidField{"setup"})
			validationError = validationError.Append(err)
		}
	}

	if desired.Action == nil {
		validationError = validationError.Append(ErrInvalidActionType)
	} else if err := desired.Action.Validate(); err != nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
		validationError = validationError.Append(err)
	}

	if desired.Monitor != nil {
		if err := desired.Monitor.Validate(); err != nil {
			validationError = validationError.Append(ErrInvalidField{"monitor"})
			validationError = validationError.Append(err)
		}
	}

	if desired.GetInstances() < 0 {
		validationError = validationError.Append(ErrInvalidField{"instances"})
	}

	if desired.GetCpuWeight() > 100 {
		validationError = validationError.Append(ErrInvalidField{"cpu_weight"})
	}

	if len(desired.GetAnnotation()) > maximumAnnotationLength {
		validationError = validationError.Append(ErrInvalidField{"annotation"})
	}

	totalRoutesLength := 0
	if desired.Routes != nil {
		for _, value := range *desired.Routes {
			totalRoutesLength += len(*value)
			if totalRoutesLength > maximumRouteLength {
				validationError = validationError.Append(ErrInvalidField{"routes"})
				break
			}
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

func (desired *DesiredLRPUpdate) Validate() error {
	var validationError ValidationError

	if desired.GetInstances() < 0 {
		validationError = validationError.Append(ErrInvalidField{"instances"})
	}

	if len(desired.GetAnnotation()) > maximumAnnotationLength {
		validationError = validationError.Append(ErrInvalidField{"annotation"})
	}

	totalRoutesLength := 0
	if desired.Routes != nil {
		for _, value := range *desired.Routes {
			totalRoutesLength += len(*value)
			if totalRoutesLength > maximumRouteLength {
				validationError = validationError.Append(ErrInvalidField{"routes"})
				break
			}
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
