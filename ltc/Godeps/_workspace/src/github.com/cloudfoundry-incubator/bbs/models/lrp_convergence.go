package models

type ConvergenceInput struct {
	AllProcessGuids map[string]struct{}
	DesiredLRPs     map[string]*DesiredLRP
	ActualLRPs      map[string]map[int32]*ActualLRP
	Domains         DomainSet
	Cells           CellSet
}

type ConvergenceChanges struct {
	ActualLRPsForExtraIndices      []*ActualLRP
	ActualLRPKeysForMissingIndices []*ActualLRPKey
	ActualLRPsWithMissingCells     []*ActualLRP
	RestartableCrashedActualLRPs   []*ActualLRP
	StaleUnclaimedActualLRPs       []*ActualLRP
}
