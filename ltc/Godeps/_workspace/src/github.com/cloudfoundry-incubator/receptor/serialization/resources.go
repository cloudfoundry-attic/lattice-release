package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func EnvironmentVariablesToModel(envVars []receptor.EnvironmentVariable) []models.EnvironmentVariable {
	if envVars == nil {
		return nil
	}
	out := make([]models.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func EnvironmentVariablesFromModel(envVars []models.EnvironmentVariable) []receptor.EnvironmentVariable {
	if envVars == nil {
		return nil
	}
	out := make([]receptor.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func PortMappingFromModel(ports []models.PortMapping) []receptor.PortMapping {
	if ports == nil {
		return nil
	}
	out := make([]receptor.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = val.ContainerPort
		out[i].HostPort = val.HostPort
	}
	return out
}

func PortMappingToModel(ports []receptor.PortMapping) []models.PortMapping {
	if ports == nil {
		return nil
	}
	out := make([]models.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = val.ContainerPort
		out[i].HostPort = val.HostPort
	}
	return out
}
