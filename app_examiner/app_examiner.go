package app_examiner

import (
	"github.com/cloudfoundry-incubator/receptor"
	"sort"
)

type AppInfo struct {
	ProcessGuid            string
	DesiredInstances       int
	ActualRunningInstances int
	DiskMB                 int
	MemoryMB               int
	Routes                 []string
}

type CellInfo struct {
	CellID           string
	RunningInstances int
	ClaimedInstances int
	Missing          bool
}

type AppExaminer interface {
	ListApps() ([]AppInfo, error)
	ListCells() ([]CellInfo, error)
}

type appExaminer struct {
	receptorClient receptor.Client
}

func New(receptorClient receptor.Client) *appExaminer {
	return &appExaminer{receptorClient}
}

func (e *appExaminer) ListCells() ([]CellInfo, error) {
	allCells := make(map[string]*CellInfo)
	cellList, err := e.receptorClient.Cells()
	if err != nil {
		return nil, err
	}

	for _, cell := range cellList {
		allCells[cell.CellID] = &CellInfo{CellID: cell.CellID}
	}

	actualLRPs, err := e.receptorClient.ActualLRPs()
	if err != nil {
		return nil, err
	}

	for _, actualLRP := range actualLRPs {
		if actualLRP.State == receptor.ActualLRPStateUnclaimed {
			continue
		}

		_, ok := allCells[actualLRP.CellID]
		if !ok {
			allCells[actualLRP.CellID] = &CellInfo{CellID: actualLRP.CellID, Missing: true}
		}

		if actualLRP.State == receptor.ActualLRPStateRunning {
			allCells[actualLRP.CellID].RunningInstances++
		} else if actualLRP.State == receptor.ActualLRPStateClaimed {
			allCells[actualLRP.CellID].ClaimedInstances++
		}
	}

	return sortCells(allCells), nil
}

func (e *appExaminer) ListApps() ([]AppInfo, error) {
	allApps := make(map[string]*AppInfo)

	desiredLRPs, err := e.receptorClient.DesiredLRPs()
	if err != nil {
		return nil, err
	}

	for _, desiredLRP := range desiredLRPs {
		allApps[desiredLRP.ProcessGuid] = &AppInfo{ProcessGuid: desiredLRP.ProcessGuid, DesiredInstances: desiredLRP.Instances, DiskMB: desiredLRP.DiskMB, MemoryMB: desiredLRP.MemoryMB, Routes: desiredLRP.Routes}
	}

	actualLRPs, err := e.receptorClient.ActualLRPs()
	if err != nil {
		return nil, err
	}

	for _, actualLRP := range actualLRPs {
		appInfo, ok := allApps[actualLRP.ProcessGuid]
		if !ok {
			appInfo = &AppInfo{ProcessGuid: actualLRP.ProcessGuid, DesiredInstances: 0}
			allApps[actualLRP.ProcessGuid] = appInfo
		}

		if actualLRP.State == receptor.ActualLRPStateRunning {
			appInfo.ActualRunningInstances++
		}
	}

	return sortApps(allApps), nil
}

func sortApps(allApps map[string]*AppInfo) []AppInfo {
	sortedKeys := sortAppKeys(allApps)

	sortedApps := make([]AppInfo, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		sortedApps = append(sortedApps, *allApps[key])
	}

	return sortedApps
}

func sortAppKeys(allApps map[string]*AppInfo) []string {
	keys := make([]string, 0, len(allApps))
	for key := range allApps {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func sortCells(allCells map[string]*CellInfo) []CellInfo {
	sortedKeys := sortCellKeys(allCells)

	sortedCells := make([]CellInfo, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		sortedCells = append(sortedCells, *allCells[key])

	}

	return sortedCells
}

func sortCellKeys(allApps map[string]*CellInfo) []string {
	keys := make([]string, 0, len(allApps))
	for key := range allApps {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
