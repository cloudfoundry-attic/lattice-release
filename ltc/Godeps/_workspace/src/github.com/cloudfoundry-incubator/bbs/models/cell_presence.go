package models

type CellSet map[string]CellPresence

func (set CellSet) Add(cell CellPresence) {
	set[cell.CellID] = cell
}

func (set CellSet) Each(predicate func(cell CellPresence)) {
	for _, cell := range set {
		predicate(cell)
	}
}

func (set CellSet) HasCellID(cellID string) bool {
	_, ok := set[cellID]
	return ok
}

type CellCapacity struct {
	MemoryMB   int `json:"memory_mb"`
	DiskMB     int `json:"disk_mb"`
	Containers int `json:"containers"`
}

func NewCellCapacity(memoryMB, diskMB, containers int) CellCapacity {
	return CellCapacity{
		MemoryMB:   memoryMB,
		DiskMB:     diskMB,
		Containers: containers,
	}
}

func (cap CellCapacity) Validate() error {
	var validationError ValidationError

	if cap.MemoryMB <= 0 {
		validationError = validationError.Append(ErrInvalidField{"memory_mb"})
	}

	if cap.DiskMB < 0 {
		validationError = validationError.Append(ErrInvalidField{"disk_mb"})
	}

	if cap.Containers <= 0 {
		validationError = validationError.Append(ErrInvalidField{"containers"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

type CellPresence struct {
	CellID          string              `json:"cell_id"`
	RepAddress      string              `json:"rep_address"`
	Zone            string              `json:"zone"`
	Capacity        CellCapacity        `json:"capacity"`
	RootFSProviders map[string][]string `json:"rootfs_providers"`
}

func NewCellPresence(cellID, repAddress, zone string, capacity CellCapacity, rootFSProviders, preloadedRootFSes []string) CellPresence {
	rootFSProviderMap := make(map[string][]string)

	for _, provider := range rootFSProviders {
		rootFSProviderMap[provider] = []string{}
	}

	rootFSProviderMap["preloaded"] = preloadedRootFSes

	return CellPresence{
		CellID:          cellID,
		RepAddress:      repAddress,
		Zone:            zone,
		Capacity:        capacity,
		RootFSProviders: rootFSProviderMap,
	}
}

func (c CellPresence) Validate() error {
	var validationError ValidationError

	if c.CellID == "" {
		validationError = validationError.Append(ErrInvalidField{"cell_id"})
	}

	if c.RepAddress == "" {
		validationError = validationError.Append(ErrInvalidField{"rep_address"})
	}

	if err := c.Capacity.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
