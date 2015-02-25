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
		Evacuating:      evacuating,
		ModificationTag: actualLRPModificationTagToResponseModificationTag(actualLRP.ModificationTag),
	}
}

func ActualLRPFromResponse(resp receptor.ActualLRPResponse) models.ActualLRP {
	return models.ActualLRP{
		ActualLRPKey:          models.NewActualLRPKey(resp.ProcessGuid, resp.Index, resp.Domain),
		ActualLRPContainerKey: models.NewActualLRPContainerKey(resp.InstanceGuid, resp.CellID),
		ActualLRPNetInfo:      models.NewActualLRPNetInfo(resp.Address, PortMappingToModel(resp.Ports)),
		State:                 actualLRPStateFromResponseState(resp.State),
		PlacementError:        resp.PlacementError,
		Since:                 resp.Since,
		ModificationTag:       actualLRPModificationTagFromResponseModificationTag(resp.ModificationTag),
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

func actualLRPStateFromResponseState(state receptor.ActualLRPState) models.ActualLRPState {
	switch state {
	case receptor.ActualLRPStateUnclaimed:
		return models.ActualLRPStateUnclaimed
	case receptor.ActualLRPStateClaimed:
		return models.ActualLRPStateClaimed
	case receptor.ActualLRPStateRunning:
		return models.ActualLRPStateRunning
	default:
		return ""
	}
}

func actualLRPModificationTagFromResponseModificationTag(modificationTag receptor.ModificationTag) models.ModificationTag {
	return models.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: modificationTag.Index,
	}
}

func actualLRPModificationTagToResponseModificationTag(modificationTag models.ModificationTag) receptor.ModificationTag {
	return receptor.ModificationTag{
		Epoch: modificationTag.Epoch,
		Index: modificationTag.Index,
	}
}
