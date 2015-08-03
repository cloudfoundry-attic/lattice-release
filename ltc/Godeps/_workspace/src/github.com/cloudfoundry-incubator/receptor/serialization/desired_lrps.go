package serialization

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
)

func DesiredLRPProtoToResponse(lrp *models.DesiredLRP) receptor.DesiredLRPResponse {
	return receptor.DesiredLRPResponse{
		ProcessGuid:          lrp.ProcessGuid,
		Domain:               lrp.Domain,
		RootFS:               lrp.RootFs,
		Instances:            int(lrp.Instances),
		EnvironmentVariables: EnvironmentVariablesFromProto(lrp.EnvironmentVariables),
		Setup:                lrp.Setup,
		Action:               lrp.Action,
		Monitor:              lrp.Monitor,
		StartTimeout:         uint(lrp.StartTimeout),
		DiskMB:               int(lrp.DiskMb),
		MemoryMB:             int(lrp.MemoryMb),
		CPUWeight:            uint(lrp.CpuWeight),
		Privileged:           lrp.Privileged,
		Ports:                PortsFromProto(lrp.Ports),
		Routes:               RoutingInfoFromProto(lrp.Routes),
		LogGuid:              lrp.LogGuid,
		LogSource:            lrp.LogSource,
		MetricsGuid:          lrp.MetricsGuid,
		Annotation:           lrp.Annotation,
		EgressRules:          lrp.EgressRules,
		ModificationTag:      desiredLRPModificationTagProtoToResponseModificationTag(lrp.ModificationTag),
	}
}

func RoutingInfoFromProto(routes *models.Routes) receptor.RoutingInfo {
	if routes == nil {
		return nil
	}

	info := receptor.RoutingInfo{}
	for key, value := range *routes {
		info[key] = value
	}
	return info
}

func desiredLRPModificationTagProtoToResponseModificationTag(modificationTag *models.ModificationTag) receptor.ModificationTag {
	if modificationTag == nil {
		return receptor.ModificationTag{}
	}
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: uint(modificationTag.Index),
	}
}

// old code -- delete when BBS server is done

func DesiredLRPFromRequest(req receptor.DesiredLRPCreateRequest) oldmodels.DesiredLRP {
	return oldmodels.DesiredLRP{
		ProcessGuid:          req.ProcessGuid,
		Domain:               req.Domain,
		RootFS:               req.RootFS,
		Instances:            req.Instances,
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
		ModificationTag:      oldmodels.ModificationTag{},
	}
}

func DesiredLRPUpdateFromRequest(req receptor.DesiredLRPUpdateRequest) oldmodels.DesiredLRPUpdate {
	return oldmodels.DesiredLRPUpdate{
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

func PortsFromProto(ports []uint32) []uint16 {
	result := []uint16{}
	for _, v := range ports {
		result = append(result, uint16(v))
	}
	return result
}
