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

type AppExaminer interface {
	ListApps() ([]AppInfo, error)
}

type appExaminer struct {
	receptorClient receptor.Client
}

func New(receptorClient receptor.Client) *appExaminer {
	return &appExaminer{receptorClient}
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
