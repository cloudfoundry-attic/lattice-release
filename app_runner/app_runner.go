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
}

func NewDiegoAppRunner(receptorClient receptor.Client) *DiegoAppRunner {
	return &DiegoAppRunner{receptorClient}
}

func (appRunner *DiegoAppRunner) StartDockerApp(name string, startCommand string, dockerImagePath string) error {
	if existingLrpCount, err := appRunner.existingLrpsCount(name); err != nil {
		return err
	} else if existingLrpCount != 0 {
		return newExistingAppError(name)
	}

	return appRunner.desireLrp(name, startCommand, dockerImagePath)
}

func (appRunner *DiegoAppRunner) ScaleDockerApp(name string, instances int) error {
	if existingLrpCount, err := appRunner.existingLrpsCount(name); err != nil {
		return err
	} else if existingLrpCount == 0 {
		return newAppNotStartedError(name)
	}

	return appRunner.updateLrp(name, instances)
}

func (appRunner *DiegoAppRunner) existingLrpsCount(name string) (int, error) {
	actualLrps, err := appRunner.receptorClient.ActualLRPsByProcessGuid(name)
	return len(actualLrps), err
}

func (appRunner *DiegoAppRunner) desireLrp(name, startCommand, dockerImagePath string) error {
	err := appRunner.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: name,
		Domain:      "diego-edge",
		RootFSPath:  dockerImagePath,
		Instances:   1,
		Stack:       "lucid64",
		Routes:      []string{fmt.Sprintf("%s.192.168.11.11.xip.io", name)},
		MemoryMB:    128,
		DiskMB:      1024,
		Ports:       []uint32{8080},
		LogGuid:     name,
		LogSource:   "APP",
		Setup: &models.DownloadAction{
			From: spyDownloadUrl,
			To:   "/tmp",
		},
		Action: &models.RunAction{
			Path: startCommand,
		},
		Monitor: &models.RunAction{
			Path: "/tmp/spy",
			Args: []string{"-addr", ":8080"},
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
