package serialization

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
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

func DesiredLRPFromRequest(req receptor.DesiredLRPCreateRequest) *models.DesiredLRP {
	return &models.DesiredLRP{
		ProcessGuid:          req.ProcessGuid,
		Domain:               req.Domain,
		RootFs:               req.RootFS,
		Instances:            int32(req.Instances),
		EnvironmentVariables: EnvironmentVariablesToModel(req.EnvironmentVariables),
		Setup:                req.Setup,
		Action:               req.Action,
		Monitor:              req.Monitor,
		StartTimeout:         uint32(req.StartTimeout),
		DiskMb:               int32(req.DiskMB),
		MemoryMb:             int32(req.MemoryMB),
		CpuWeight:            uint32(req.CPUWeight),
		Privileged:           req.Privileged,
		Ports:                PortsToProto(req.Ports),
		Routes:               RoutingInfoToRawMessages(req.Routes),
		LogGuid:              req.LogGuid,
		LogSource:            req.LogSource,
		MetricsGuid:          req.MetricsGuid,
		Annotation:           req.Annotation,
		EgressRules:          req.EgressRules,
		ModificationTag:      &models.ModificationTag{},
	}
}

func DesiredLRPUpdateFromRequest(req receptor.DesiredLRPUpdateRequest) *models.DesiredLRPUpdate {
	var requestedInstances *int32
	if req.Instances != nil {
		requestedInstancesValue := int32(*req.Instances)
		requestedInstances = &requestedInstancesValue
	}

	return &models.DesiredLRPUpdate{
		Instances:  requestedInstances,
		Routes:     RoutingInfoToRawMessages(req.Routes),
		Annotation: req.Annotation,
	}
}

func RoutingInfoToRawMessages(r receptor.RoutingInfo) *models.Routes {
	if r == nil {
		return nil
	}

	routes := models.Routes{}
	for key, value := range r {
		routes[key] = value
	}

	return &routes
}

func PortsFromProto(ports []uint32) []uint16 {
	result := []uint16{}
	for _, v := range ports {
		result = append(result, uint16(v))
	}
	return result
}

func PortsToProto(ports []uint16) []uint32 {
	result := []uint32{}
	for _, v := range ports {
		result = append(result, uint32(v))
	}
	return result
}
