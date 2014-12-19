package app_runner

import (
	"fmt"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type AppRunner interface {
	StartDockerApp(name, startCommand, dockerImagePath string, appArgs []string, environmentVariables map[string]string, privileged bool, instances, memoryMB, diskMB, port int) error
	ScaleApp(name string, instances int) error
	RemoveApp(name string) error
	AppExists(name string) (bool, error)
	NumOfRunningAppInstances(name string) (int, error)
}

const (
	spyDownloadUrl string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
)

type appRunner struct {
	receptorClient receptor.Client
	domain         string
}

func New(receptorClient receptor.Client, domain string) AppRunner {
	return &appRunner{receptorClient, domain}
}

func (appRunner *appRunner) StartDockerApp(name, dockerImagePath, startCommand string, appArgs []string, environmentVariables map[string]string, privileged bool, instances, memoryMB, diskMB, port int) error {
	if exists, err := appRunner.desiredLRPExists(name); err != nil {
		return err
	} else if exists {
		return newExistingAppError(name)
	}
	return appRunner.desireLrp(name, startCommand, dockerImagePath, appArgs, environmentVariables, privileged, instances, memoryMB, diskMB, port)
}

func (appRunner *appRunner) ScaleApp(name string, instances int) error {
	if exists, err := appRunner.desiredLRPExists(name); err != nil {
		return err
	} else if !exists {
		return newAppNotStartedError(name)
	}

	return appRunner.updateLrp(name, instances)
}

func (appRunner *appRunner) RemoveApp(name string) error {
	if lrpExists, err := appRunner.desiredLRPExists(name); err != nil {
		return err
	} else if !lrpExists {
		return newAppNotStartedError(name)
	}

	return appRunner.receptorClient.DeleteDesiredLRP(name)
}

func (appRunner *appRunner) AppExists(name string) (bool, error) {
	actualLRPs, err := appRunner.receptorClient.ActualLRPs()
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

func (appRunner *appRunner) NumOfRunningAppInstances(name string) (count int, err error) {
	runningInstances := 0
	instances, err := appRunner.receptorClient.ActualLRPsByProcessGuid(name)
	if err != nil {
		return 0, err
	}

	for _, instance := range instances {
		if instance.State == receptor.ActualLRPStateRunning {
			runningInstances += 1
		}
	}

	return runningInstances, nil
}

func (appRunner *appRunner) desiredLRPExists(name string) (exists bool, err error) {
	desiredLRPs, err := appRunner.receptorClient.DesiredLRPs()
	if err != nil {
		return false, err
	}

	for _, desiredLRP := range desiredLRPs {
		if desiredLRP.ProcessGuid == name {
			return true, nil
		}
	}

	return false, nil
}

func (appRunner *appRunner) desireLrp(name, startCommand, dockerImagePath string, appArgs []string, environmentVariables map[string]string, privileged bool, instances, memoryMB, diskMB, port int) error {
	err := appRunner.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid:          name,
		Domain:               "diego-edge",
		RootFSPath:           dockerImagePath,
		Instances:            instances,
		Stack:                "lucid64",
		Routes:               []string{fmt.Sprintf("%s.%s", name, appRunner.domain)},
		MemoryMB:             memoryMB,
		DiskMB:               diskMB,
		Ports:                []uint32{uint32(port)},
		LogGuid:              name,
		LogSource:            "APP",
		EnvironmentVariables: buildEnvironmentVariables(environmentVariables, port),
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
			Path:      "/tmp/spy",
			Args:      []string{"-addr", fmt.Sprintf(":%d", port)},
			LogSource: "HEALTH",
		},
	})

	return err
}

func buildEnvironmentVariables(environmentVariables map[string]string, port int) []receptor.EnvironmentVariable {
	appEnvVars := make([]receptor.EnvironmentVariable, 0, len(environmentVariables)+1)
	for name, value := range environmentVariables {
		appEnvVars = append(appEnvVars, receptor.EnvironmentVariable{Name: name, Value: value})
	}
	return append(appEnvVars, receptor.EnvironmentVariable{Name: "PORT", Value: fmt.Sprintf("%d", port)})
}

func (appRunner *appRunner) updateLrp(name string, instances int) error {
	err := appRunner.receptorClient.UpdateDesiredLRP(
		name,
		receptor.DesiredLRPUpdateRequest{
			Instances: &instances,
		},
	)

	return err
}
