package app_runner

import (
	"fmt"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

const (
	spyDownloadUrl string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
)

type DiegoAppRunner struct {
	receptorClient receptor.Client
	domain         string
}

func NewDiegoAppRunner(receptorClient receptor.Client, domain string) *DiegoAppRunner {
	return &DiegoAppRunner{receptorClient, domain}
}

func (appRunner *DiegoAppRunner) StartDockerApp(name, dockerImagePath, startCommand string, appArgs []string, privileged bool, memoryMB, diskMB, port int) error {
	if desiredLRPsCount, err := appRunner.desiredLRPsCount(name); err != nil {
		return err
	} else if desiredLRPsCount != 0 {
		return newExistingAppError(name)
	}
	return appRunner.desireLrp(name, startCommand, dockerImagePath, appArgs, privileged, memoryMB, diskMB, port)
}

func (appRunner *DiegoAppRunner) ScaleDockerApp(name string, instances int) error {
	if desiredLRPsCount, err := appRunner.desiredLRPsCount(name); err != nil {
		return err
	} else if desiredLRPsCount == 0 {
		return newAppNotStartedError(name)
	}

	return appRunner.updateLrp(name, instances)
}

func (appRunner *DiegoAppRunner) StopDockerApp(name string) error {
	if desiredLRPsCount, err := appRunner.desiredLRPsCount(name); err != nil {
		return err
	} else if desiredLRPsCount == 0 {
		return newAppNotStartedError(name)
	}

	return appRunner.receptorClient.DeleteDesiredLRP(name)
}

func (appRunner *DiegoAppRunner) IsDockerAppUp(processGuid string) (bool, error) {
	actualLrps, err := appRunner.receptorClient.ActualLRPsByProcessGuid(processGuid)
	status := len(actualLrps) > 0 && actualLrps[0].State == receptor.ActualLRPStateRunning

	return status, err
}

func (appRunner *DiegoAppRunner) desiredLRPsCount(name string) (int, error) {
	desiredLRPs, err := appRunner.receptorClient.DesiredLRPs()
	if err != nil {
		return 0, err
	}

	for _, desiredLRP := range desiredLRPs {
		if desiredLRP.ProcessGuid == name {
			return desiredLRP.Instances, nil
		}
	}

	return 0, nil
}

func (appRunner *DiegoAppRunner) desireLrp(name, startCommand, dockerImagePath string, appArgs []string, privileged bool, memoryMB, diskMB, port int) error {
	err := appRunner.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: name,
		Domain:      "diego-edge",
		RootFSPath:  dockerImagePath,
		Instances:   1,
		Stack:       "lucid64",
		Routes:      []string{fmt.Sprintf("%s.%s", name, appRunner.domain)},
		MemoryMB:    memoryMB,
		DiskMB:      diskMB,
		Ports:       []uint32{uint32(port)},
		LogGuid:     name,
		LogSource:   "APP",
		Setup: &models.DownloadAction{
			From: spyDownloadUrl,
			To:   "/tmp",
		},
		Action: &models.RunAction{
			Path:       startCommand,
			Args:       appArgs,
			Privileged: privileged,
		},
		Monitor: &models.RunAction{
			Path: "/tmp/spy",
			Args: []string{"-addr", fmt.Sprintf(":%d", port)},
		},
	})

	return err
}

func (appRunner *DiegoAppRunner) updateLrp(name string, instances int) error {
	err := appRunner.receptorClient.UpdateDesiredLRP(
		name,
		receptor.DesiredLRPUpdateRequest{
			Instances: &instances,
		},
	)

	return err
}
