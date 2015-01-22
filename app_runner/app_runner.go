package app_runner

//go:generate counterfeiter -o fake_app_runner/fake_app_runner.go . AppRunner

import (
	"fmt"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_repository_name_formatter"
)

type AppRunner interface {
	StartDockerApp(params StartDockerAppParams) error
	ScaleApp(name string, instances int) error
	RemoveApp(name string) error
	AppExists(name string) (bool, error)
	NumOfRunningAppInstances(name string) (int, error)
}

type StartDockerAppParams struct {
	Name                 string
	StartCommand         string
	DockerImagePath      string
	AppArgs              []string
	EnvironmentVariables map[string]string
	Privileged           bool
	Instances            int
	MemoryMB             int
	DiskMB               int
	Port                 int
	WorkingDir           string
}

const (
	spyDownloadUrl string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
	lrpDomain      string = "lattice"
)

type appRunner struct {
	receptorClient receptor.Client
	systemDomain   string
}

func New(receptorClient receptor.Client, systemDomain string) AppRunner {
	return &appRunner{receptorClient, systemDomain}
}

func (appRunner *appRunner) StartDockerApp(params StartDockerAppParams) error {
	if exists, err := appRunner.desiredLRPExists(params.Name); err != nil {
		return err
	} else if exists {
		return newExistingAppError(params.Name)
	}

	if err := appRunner.receptorClient.UpsertDomain(lrpDomain, 0); err != nil {
		return err
	}

	return appRunner.desireLrp(params)
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

func (appRunner *appRunner) desireLrp(params StartDockerAppParams) error {
	dockerImageUrl, err := docker_repository_name_formatter.FormatForReceptor(params.DockerImagePath)
	if err != nil {
		return err
	}
	err = appRunner.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid:          params.Name,
		Domain:               lrpDomain,
		RootFSPath:           dockerImageUrl,
		Instances:            params.Instances,
		Stack:                "lucid64",
		Routes:               []string{fmt.Sprintf("%s.%s", params.Name, appRunner.systemDomain)},
		MemoryMB:             params.MemoryMB,
		DiskMB:               params.DiskMB,
		Ports:                []uint32{uint32(params.Port)},
		LogGuid:              params.Name,
		LogSource:            "APP",
		EnvironmentVariables: buildEnvironmentVariables(params.EnvironmentVariables, params.Port),
		Setup: &models.DownloadAction{
			From: spyDownloadUrl,
			To:   "/tmp",
		},
		Action: &models.RunAction{
			Path:       params.StartCommand,
			Args:       params.AppArgs,
			Privileged: params.Privileged,
			Dir:        params.WorkingDir,
		},
		Monitor: &models.RunAction{
			Path:      "/tmp/spy",
			Args:      []string{"-addr", fmt.Sprintf(":%d", params.Port)},
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
