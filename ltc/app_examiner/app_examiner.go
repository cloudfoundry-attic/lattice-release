package app_examiner

import (
	"errors"
	"sort"

	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/receptor"
)

const AppNotFoundErrorMessage = "App not found."

type EnvironmentVariable struct {
	Name  string
	Value string
}

type AppInfo struct {
	ProcessGuid            string
	DesiredInstances       int
	ActualRunningInstances int
	EnvironmentVariables   []EnvironmentVariable
	StartTimeout           uint
	DiskMB                 int
	MemoryMB               int
	CPUWeight              uint
	Ports                  []uint16
	Routes                 route_helpers.AppRoutes
	LogGuid                string
	LogSource              string
	Annotation             string
	ActualInstances        []InstanceInfo
}

type PortMapping struct {
	HostPort      uint16
	ContainerPort uint16
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
	CrashCount     int
	HasMetrics     bool
	Metrics        InstanceMetrics
}

type InstanceMetrics struct {
	CpuPercentage float64
	MemoryBytes   uint64
	DiskBytes     uint64
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
	Zone             string
	MemoryMB         int
	DiskMB           int
	Containers       int
}

//go:generate counterfeiter -o fake_app_examiner/fake_app_examiner.go . AppExaminer
type AppExaminer interface {
	ListApps() ([]AppInfo, error)
	ListCells() ([]CellInfo, error)
	AppStatus(appName string) (AppInfo, error)
	AppExists(name string) (bool, error)
	RunningAppInstancesInfo(name string) (int, bool, error)
}

type appExaminer struct {
	receptorClient receptor.Client
	noaaConsumer   NoaaConsumer
}

func New(receptorClient receptor.Client, noaaConsumer NoaaConsumer) AppExaminer {
	return &appExaminer{receptorClient, noaaConsumer}
}

func (e *appExaminer) ListCells() ([]CellInfo, error) {
	allCells := make(map[string]*CellInfo)
	cellList, err := e.receptorClient.Cells()
	if err != nil {
		return nil, err
	}

	for _, cell := range cellList {
		allCells[cell.CellID] = &CellInfo{
			CellID:     cell.CellID,
			Zone:       cell.Zone,
			MemoryMB:   cell.Capacity.MemoryMB,
			DiskMB:     cell.Capacity.DiskMB,
			Containers: cell.Capacity.Containers,
		}
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

	containerMetrics, err := e.noaaConsumer.GetContainerMetrics(appName, "")
	if err != nil {
		return *appInfoPtr, nil
	}

	indexMap := make(map[int]int, 0)
	for index, instance := range appInfoPtr.ActualInstances {
		indexMap[instance.Index] = index
	}

	for _, metric := range containerMetrics {
		metricIndex, ok := indexMap[int(metric.GetInstanceIndex())]
		if !ok {
			continue
		}
		instanceInfo := &appInfoPtr.ActualInstances[metricIndex]
		instanceInfo.HasMetrics = true
		instanceInfo.Metrics = InstanceMetrics{
			CpuPercentage: metric.GetCpuPercentage(),
			MemoryBytes:   metric.GetMemoryBytes(),
			DiskBytes:     metric.GetDiskBytes(),
		}
	}
	return *appInfoPtr, nil
}

func (e *appExaminer) AppExists(name string) (bool, error) {
	actualLRPs, err := e.receptorClient.ActualLRPs()
	if err != nil {
		return false, err
	}

	for _, actualLRP := range actualLRPs {
		if actualLRP.ProcessGuid == name {
			return true, nil
		}
	}

	return false, nil
}

func (e *appExaminer) RunningAppInstancesInfo(name string) (count int, placementError bool, err error) {
	runningInstances := 0
	placementErrorOccurred := false
	instances, err := e.receptorClient.ActualLRPsByProcessGuid(name)
	if err != nil {
		return 0, false, err
	}

	for _, instance := range instances {
		if instance.State == receptor.ActualLRPStateRunning {
			runningInstances += 1
		}

		if instance.PlacementError != "" {
			placementErrorOccurred = true
		}
	}

	return runningInstances, placementErrorOccurred, nil
}

func mergeDesiredActualLRPs(desiredLRPs []receptor.DesiredLRPResponse, actualLRPs []receptor.ActualLRPResponse) map[string]*AppInfo {
	appMap := make(map[string]*AppInfo)

	for _, desiredLRP := range desiredLRPs {
		appMap[desiredLRP.ProcessGuid] = &AppInfo{
			ProcessGuid:            desiredLRP.ProcessGuid,
			DesiredInstances:       desiredLRP.Instances,
			ActualRunningInstances: 0,
			EnvironmentVariables:   buildEnvVars(desiredLRP),
			StartTimeout:           desiredLRP.StartTimeout,
			DiskMB:                 desiredLRP.DiskMB,
			MemoryMB:               desiredLRP.MemoryMB,
			CPUWeight:              desiredLRP.CPUWeight,
			Ports:                  desiredLRP.Ports,
			Routes:                 route_helpers.AppRoutesFromRoutingInfo(desiredLRP.Routes),
			LogGuid:                desiredLRP.LogGuid,
			LogSource:              desiredLRP.LogSource,
			Annotation:             desiredLRP.Annotation,
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
			CrashCount:     actualLRP.CrashCount,
			HasMetrics:     false,
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

func sortCellKeys(allCells map[string]*CellInfo) []string {
	keys := make([]string, 0, len(allCells))
	for key := range allCells {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
