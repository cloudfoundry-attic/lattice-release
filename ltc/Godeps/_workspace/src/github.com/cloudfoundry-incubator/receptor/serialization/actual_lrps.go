package serialization

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
)

func ActualLRPProtoToResponse(actualLRP *models.ActualLRP, evacuating bool) receptor.ActualLRPResponse {
	return receptor.ActualLRPResponse{
		ProcessGuid:     actualLRP.ProcessGuid,
		InstanceGuid:    actualLRP.InstanceGuid,
		CellID:          actualLRP.CellId,
		Domain:          actualLRP.Domain,
		Index:           int(actualLRP.Index),
		Address:         actualLRP.Address,
		Ports:           PortMappingFromProto(actualLRP.Ports),
		State:           actualLRPProtoStateToResponseState(actualLRP.State),
		PlacementError:  actualLRP.PlacementError,
		Since:           actualLRP.Since,
		CrashCount:      int(actualLRP.CrashCount),
		CrashReason:     actualLRP.CrashReason,
		Evacuating:      evacuating,
		ModificationTag: actualLRPProtoModificationTagToResponseModificationTag(actualLRP.ModificationTag),
	}
}

func actualLRPProtoStateToResponseState(state string) receptor.ActualLRPState {
	switch state {
	case models.ActualLRPStateUnclaimed:
		return receptor.ActualLRPStateUnclaimed
	case models.ActualLRPStateClaimed:
		return receptor.ActualLRPStateClaimed
	case models.ActualLRPStateRunning:
		return receptor.ActualLRPStateRunning
	case models.ActualLRPStateCrashed:
		return receptor.ActualLRPStateCrashed
	default:
		return receptor.ActualLRPStateInvalid
	}
}

func actualLRPProtoModificationTagToResponseModificationTag(modificationTag models.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: uint(modificationTag.Index),
	}
}
