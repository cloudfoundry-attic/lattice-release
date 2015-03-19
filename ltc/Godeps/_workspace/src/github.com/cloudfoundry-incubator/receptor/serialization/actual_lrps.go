package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func ActualLRPToResponse(actualLRP models.ActualLRP, evacuating bool) receptor.ActualLRPResponse {
	return receptor.ActualLRPResponse{
		ProcessGuid:     actualLRP.ProcessGuid,
		InstanceGuid:    actualLRP.InstanceGuid,
		CellID:          actualLRP.CellID,
		Domain:          actualLRP.Domain,
		Index:           actualLRP.Index,
		Address:         actualLRP.Address,
		Ports:           PortMappingFromModel(actualLRP.Ports),
		State:           actualLRPStateToResponseState(actualLRP.State),
		PlacementError:  actualLRP.PlacementError,
		Since:           actualLRP.Since,
		CrashCount:      actualLRP.CrashCount,
		CrashReason:     actualLRP.CrashReason,
		Evacuating:      evacuating,
		ModificationTag: actualLRPModificationTagToResponseModificationTag(actualLRP.ModificationTag),
	}
}

func actualLRPStateToResponseState(state models.ActualLRPState) receptor.ActualLRPState {
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

func actualLRPModificationTagToResponseModificationTag(modificationTag models.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: modificationTag.Index,
	}
}
