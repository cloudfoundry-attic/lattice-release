package docker_app_runner

//go:generate counterfeiter -o fake_app_runner/fake_app_runner.go . AppRunner

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_repository_name_formatter"
	"github.com/pivotal-cf-experimental/lattice-cli/route_helpers"
)

type AppRunner interface {
	StartDockerApp(params StartDockerAppParams) error
	ScaleApp(name string, instances int) error
	RemoveApp(name string) error
	AppExists(name string) (bool, error)
	NumOfRunningAppInstances(name string) (int, error)
}

type PortConfig struct {
	Monitored uint16
	Exposed   []uint16
}

func (portConfig PortConfig) IsEmpty() bool {
	return len(portConfig.Exposed) == 0
}

type StartDockerAppParams struct {
	Name                 string
	StartCommand         string
	DockerImagePath      string
	AppArgs              []string
	EnvironmentVariables map[string]string
	Privileged           bool
	Monitor              bool
	Instances            int
	MemoryMB             int
	DiskMB               int
	Ports                PortConfig
	WorkingDir           string
}

const (
	healthcheckDownloadUrl string = "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz"
	lrpDomain              string = "lattice"
)

type appRunner struct {
	receptorClient receptor.Client
	systemDomain   string
}

func New(receptorClient receptor.Client, systemDomain string) AppRunner {
	return &appRunner{receptorClient, systemDomain}
}

func (appRunner *appRunner) StartDockerApp(params StartDockerAppParams) error {
	exposedContainsMonitored := false
	for _, port := range params.Ports.Exposed {
		if port == params.Ports.Monitored {
			exposedContainsMonitored = true
			break
		}
	}

	if params.Monitor && !exposedContainsMonitored {
		return errors.New("Monitored Port must be in the Exposed Ports.")
	}

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

	envVars := buildEnvironmentVariables(params.EnvironmentVariables)
	envVars = append(envVars, receptor.EnvironmentVariable{Name: "PORT", Value: fmt.Sprintf("%d", params.Ports.Monitored)})

	req := receptor.DesiredLRPCreateRequest{
		ProcessGuid:          params.Name,
		Domain:               lrpDomain,
		RootFSPath:           dockerImageUrl,
		Instances:            params.Instances,
		Stack:                "lucid64",
		Routes:               appRunner.buildRoutingInfo(params.Name, params.Ports),
		MemoryMB:             params.MemoryMB,
		DiskMB:               params.DiskMB,
		Privileged:           true,
		Ports:                params.Ports.Exposed,
		LogGuid:              params.Name,
		LogSource:            "APP",
		EnvironmentVariables: envVars,
		Setup: &models.DownloadAction{
			From: healthcheckDownloadUrl,
			To:   "/tmp",
		},
		Action: &models.RunAction{
			Path:       params.StartCommand,
			Args:       params.AppArgs,
			Privileged: params.Privileged,
			Dir:        params.WorkingDir,
		},
	}

	if params.Monitor {
		req.Monitor = &models.RunAction{
			Path:      "/tmp/healthcheck",
			Args:      []string{"-port", fmt.Sprintf("%d", params.Ports.Monitored)},
			LogSource: "HEALTH",
		}
	}
	err = appRunner.receptorClient.CreateDesiredLRP(req)

	return err
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

func (appRunner *appRunner) buildRoutingInfo(appName string, portConfig PortConfig) receptor.RoutingInfo {
	appRoutes := route_helpers.AppRoutes{}

	appRoutes = append(appRoutes, route_helpers.AppRoute{
		Hostnames: []string{fmt.Sprintf("%s.%s", appName, appRunner.systemDomain)},
		Port:      portConfig.Monitored,
	})

	for _, port := range portConfig.Exposed {
		appRoutes = append(appRoutes, route_helpers.AppRoute{
			Hostnames: []string{fmt.Sprintf("%s-%s.%s", appName, strconv.Itoa(int(port)), appRunner.systemDomain)},
			Port:      port,
		})
	}

	return appRoutes.RoutingInfo()
}

func buildEnvironmentVariables(environmentVariables map[string]string) []receptor.EnvironmentVariable {
	appEnvVars := make([]receptor.EnvironmentVariable, 0, len(environmentVariables)+1)
	for name, value := range environmentVariables {
		appEnvVars = append(appEnvVars, receptor.EnvironmentVariable{Name: name, Value: value})
	}
	return appEnvVars
}
