package serialization

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
)

func DesiredLRPProtoToResponse(lrp *models.DesiredLRP) receptor.DesiredLRPResponse {
	return receptor.DesiredLRPResponse{
		ProcessGuid:          lrp.GetProcessGuid(),
		Domain:               lrp.GetDomain(),
		RootFS:               lrp.GetRootFs(),
		Instances:            int(lrp.GetInstances()),
		EnvironmentVariables: EnvironmentVariablesFromProto(lrp.GetEnvironmentVariables()),
		Setup:                models.UnwrapAction(lrp.GetSetup()),
		Action:               models.UnwrapAction(lrp.GetAction()),
		Monitor:              models.UnwrapAction(lrp.GetMonitor()),
		StartTimeout:         uint(lrp.GetStartTimeout()),
		DiskMB:               int(lrp.GetDiskMb()),
		MemoryMB:             int(lrp.GetMemoryMb()),
		CPUWeight:            uint(lrp.GetCpuWeight()),
		Privileged:           lrp.GetPrivileged(),
		Ports:                PortsFromProto(lrp.Ports),
		Routes:               RoutingInfoFromProto(lrp.Routes),
		LogGuid:              lrp.GetLogGuid(),
		LogSource:            lrp.GetLogSource(),
		MetricsGuid:          lrp.GetMetricsGuid(),
		Annotation:           lrp.GetAnnotation(),
		EgressRules:          EgressRulesFromProto(lrp.EgressRules),
		ModificationTag:      desiredLRPModificationTagProtoToResponseModificationTag(lrp.ModificationTag),
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

func DesiredLRPToResponse(lrp oldmodels.DesiredLRP) receptor.DesiredLRPResponse {
	return receptor.DesiredLRPResponse{
		ProcessGuid:          lrp.ProcessGuid,
		Domain:               lrp.Domain,
		RootFS:               lrp.RootFS,
		Instances:            lrp.Instances,
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

func EgressRulesFromProto(securityGroupRules []*models.SecurityGroupRule) []oldmodels.SecurityGroupRule {
	if securityGroupRules == nil {
		return nil
	}
	result := []oldmodels.SecurityGroupRule{}
	for _, v := range securityGroupRules {
		s := oldmodels.SecurityGroupRule{
			Protocol:     oldmodels.ProtocolName(v.GetProtocol()),
			Destinations: v.GetDestinations(),
			Ports:        PortsFromProto(v.Ports),
			PortRange: &oldmodels.PortRange{
				Start: uint16(v.GetPortRange().GetStart()),
				End:   uint16(v.GetPortRange().GetEnd()),
			},
			IcmpInfo: &oldmodels.ICMPInfo{
				Type: v.GetIcmpInfo().GetType(),
				Code: v.GetIcmpInfo().GetCode(),
			},
			Log: v.GetLog(),
		}
		result = append(result, s)
	}
	return result
}

func PortsFromProto(ports []uint32) []uint16 {
	result := []uint16{}
	for _, v := range ports {
		result = append(result, uint16(v))
	}
	return result
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

func desiredLRPModificationTagProtoToResponseModificationTag(modificationTag *models.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.GetEpoch(),
		Index: uint(modificationTag.GetIndex()),
	}
}

func desiredLRPModificationTagToResponseModificationTag(modificationTag oldmodels.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: modificationTag.Index,
	}
}
