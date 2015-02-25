package serialization

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func DesiredLRPFromRequest(req receptor.DesiredLRPCreateRequest) models.DesiredLRP {
	return models.DesiredLRP{
		ProcessGuid:          req.ProcessGuid,
		Domain:               req.Domain,
		RootFSPath:           req.RootFSPath,
		Instances:            req.Instances,
		Stack:                req.Stack,
		EnvironmentVariables: EnvironmentVariablesToModel(req.EnvironmentVariables),
		Setup:                req.Setup,
		Action:               req.Action,
		Monitor:              req.Monitor,
		StartTimeout:         req.StartTimeout,
		DiskMB:               req.DiskMB,
		MemoryMB:             req.MemoryMB,
		CPUWeight:            req.CPUWeight,
		Privileged:           req.Privileged,
		Ports:                req.Ports,
		Routes:               RoutingInfoToRawMessages(req.Routes),
		LogGuid:              req.LogGuid,
		LogSource:            req.LogSource,
		MetricsGuid:          req.MetricsGuid,
		Annotation:           req.Annotation,
		EgressRules:          req.EgressRules,
		ModificationTag:      models.ModificationTag{},
	}
}

func DesiredLRPToResponse(lrp models.DesiredLRP) receptor.DesiredLRPResponse {
	return receptor.DesiredLRPResponse{
		ProcessGuid:          lrp.ProcessGuid,
		Domain:               lrp.Domain,
		RootFSPath:           lrp.RootFSPath,
		Instances:            lrp.Instances,
		Stack:                lrp.Stack,
		EnvironmentVariables: EnvironmentVariablesFromModel(lrp.EnvironmentVariables),
		Setup:                lrp.Setup,
		Action:               lrp.Action,
		Monitor:              lrp.Monitor,
		StartTimeout:         lrp.StartTimeout,
		DiskMB:               lrp.DiskMB,
		MemoryMB:             lrp.MemoryMB,
		CPUWeight:            lrp.CPUWeight,
		Privileged:           lrp.Privileged,
		Ports:                lrp.Ports,
		Routes:               RoutingInfoFromRawMessages(lrp.Routes),
		LogGuid:              lrp.LogGuid,
		LogSource:            lrp.LogSource,
		MetricsGuid:          lrp.MetricsGuid,
		Annotation:           lrp.Annotation,
		EgressRules:          lrp.EgressRules,
		ModificationTag:      desiredLRPModificationTagToResponseModificationTag(lrp.ModificationTag),
	}
}

func DesiredLRPFromResponse(resp receptor.DesiredLRPResponse) models.DesiredLRP {
	return models.DesiredLRP{
		ProcessGuid:          resp.ProcessGuid,
		Domain:               resp.Domain,
		RootFSPath:           resp.RootFSPath,
		Instances:            resp.Instances,
		Stack:                resp.Stack,
		EnvironmentVariables: EnvironmentVariablesToModel(resp.EnvironmentVariables),
		Setup:                resp.Setup,
		Action:               resp.Action,
		Monitor:              resp.Monitor,
		StartTimeout:         resp.StartTimeout,
		DiskMB:               resp.DiskMB,
		MemoryMB:             resp.MemoryMB,
		CPUWeight:            resp.CPUWeight,
		Privileged:           resp.Privileged,
		Ports:                resp.Ports,
		Routes:               RoutingInfoToRawMessages(resp.Routes),
		LogGuid:              resp.LogGuid,
		LogSource:            resp.LogSource,
		MetricsGuid:          resp.MetricsGuid,
		Annotation:           resp.Annotation,
	}
}

func DesiredLRPUpdateFromRequest(req receptor.DesiredLRPUpdateRequest) models.DesiredLRPUpdate {
	return models.DesiredLRPUpdate{
		Instances:  req.Instances,
		Routes:     RoutingInfoToRawMessages(req.Routes),
		Annotation: req.Annotation,
	}
}

func RoutingInfoToRawMessages(r receptor.RoutingInfo) map[string]*json.RawMessage {
	var messages map[string]*json.RawMessage

	if r != nil {
		messages = map[string]*json.RawMessage{}
		for key, value := range r {
			messages[key] = value
		}
	}

	return messages
}

func RoutingInfoFromRawMessages(raw map[string]*json.RawMessage) receptor.RoutingInfo {
	if raw == nil {
		return nil
	}

	info := receptor.RoutingInfo{}
	for key, value := range raw {
		info[key] = value
	}
	return info
}

func desiredLRPModificationTagToResponseModificationTag(modificationTag models.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: modificationTag.Index,
	}
}
