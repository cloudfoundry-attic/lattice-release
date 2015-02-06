package app_examiner

//go:generate counterfeiter -o fake_app_examiner/fake_app_examiner.go . AppExaminer

import (
	"sort"

	"errors"
	"github.com/cloudfoundry-incubator/receptor"
)

const AppNotFoundErrorMessage = "App not found."

type EnvironmentVariable struct {
	Name  string
	Value string
}

type AppInfo struct {
	ProcessGuid            string //TODO: Should be name??
	DesiredInstances       int
	ActualRunningInstances int
	Stack                  string
	EnvironmentVariables   []EnvironmentVariable
	StartTimeout           uint
	DiskMB                 int
	MemoryMB               int
	CPUWeight              uint
	Ports                  []uint32
	Routes                 []string
	LogGuid                string
	LogSource              string
	Annotation             string
	ActualInstances        []InstanceInfo
}

type PortMapping struct {
	HostPort      uint32
	ContainerPort uint32
}

type InstanceInfo struct {
	InstanceGuid   string
	CellID         string
	Index          int
	Ip             string
	Ports          []PortMapping
	State          string
	Since          int64
	PlacementError string
}

type instanceInfoSortableByIndex []InstanceInfo

func (x instanceInfoSortableByIndex) Len() int {
	return len(x)
}

func (x instanceInfoSortableByIndex) Less(i, j int) bool {
	return x[i].Index < x[j].Index
}

func (x instanceInfoSortableByIndex) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
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
	AppStatus(appName string) (AppInfo, error)
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
	desiredLRPs, err := e.receptorClient.DesiredLRPs()
	if err != nil {
		return nil, err
	}

	actualLRPs, err := e.receptorClient.ActualLRPs()
	if err != nil {
		return nil, err
	}

	appMap := mergeDesiredActualLRPs(desiredLRPs, actualLRPs)
	return sortApps(appMap), nil
}

func (e *appExaminer) AppStatus(appName string) (AppInfo, error) {
	desiredLRP, err := e.receptorClient.GetDesiredLRP(appName)
	if err != nil {
		receptorError := err.(receptor.Error)
		if receptorError.Type == receptor.DesiredLRPNotFound {
			desiredLRP = receptor.DesiredLRPResponse{}
		} else {
			return AppInfo{}, err
		}
	}

	actualLRPs, err := e.receptorClient.ActualLRPsByProcessGuid(appName)
	if err != nil {
		return AppInfo{}, err
	}

	appMap := mergeDesiredActualLRPs([]receptor.DesiredLRPResponse{desiredLRP}, actualLRPs)

	appInfoPtr, ok := appMap[appName]
	if !ok {
		return AppInfo{}, errors.New(AppNotFoundErrorMessage)
	}

	return *appInfoPtr, nil
}

func mergeDesiredActualLRPs(desiredLRPs []receptor.DesiredLRPResponse, actualLRPs []receptor.ActualLRPResponse) map[string]*AppInfo {
	appMap := make(map[string]*AppInfo)

	for _, desiredLRP := range desiredLRPs {
		appMap[desiredLRP.ProcessGuid] = &AppInfo{
			ProcessGuid:            desiredLRP.ProcessGuid,
			DesiredInstances:       desiredLRP.Instances,
			ActualRunningInstances: 0,
			Stack:                desiredLRP.Stack,
			EnvironmentVariables: buildEnvVars(desiredLRP),
			StartTimeout:         desiredLRP.StartTimeout,
			DiskMB:               desiredLRP.DiskMB,
			MemoryMB:             desiredLRP.MemoryMB,
			CPUWeight:            desiredLRP.CPUWeight,
			Ports:                desiredLRP.Ports,
			Routes:               desiredLRP.Routes,
			LogGuid:              desiredLRP.LogGuid,
			LogSource:            desiredLRP.LogSource,
			Annotation:           desiredLRP.Annotation,
		}
	}

	for _, actualLRP := range actualLRPs {

		appInfo, ok := appMap[actualLRP.ProcessGuid]
		if !ok {
			appInfo = &AppInfo{ProcessGuid: actualLRP.ProcessGuid, DesiredInstances: 0}
			appMap[actualLRP.ProcessGuid] = appInfo
		}

		if actualLRP.State == receptor.ActualLRPStateRunning {
			appInfo.ActualRunningInstances++
		}

		instancePorts := make([]PortMapping, 0)
		for _, respPorts := range actualLRP.Ports {
			instancePorts = append(instancePorts, PortMapping{HostPort: respPorts.HostPort, ContainerPort: respPorts.ContainerPort})
		}
		instanceInfo := InstanceInfo{
			InstanceGuid:   actualLRP.InstanceGuid,
			CellID:         actualLRP.CellID,
			Index:          actualLRP.Index,
			Ip:             actualLRP.Address,
			Ports:          instancePorts,
			State:          string(actualLRP.State),
			Since:          actualLRP.Since,
			PlacementError: actualLRP.PlacementError,
		}

		appMap[actualLRP.ProcessGuid].ActualInstances = append(appMap[actualLRP.ProcessGuid].ActualInstances, instanceInfo)
	}

	for _, appInfo := range appMap {
		sort.Sort(instanceInfoSortableByIndex(appInfo.ActualInstances))
	}

	return appMap
}

func buildEnvVars(desiredLRPResponse receptor.DesiredLRPResponse) []EnvironmentVariable {
	envVars := make([]EnvironmentVariable, 0)
	for _, envVar := range desiredLRPResponse.EnvironmentVariables {
		envVars = append(envVars, EnvironmentVariable{Name: envVar.Name, Value: envVar.Value})
	}
	return envVars
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
