package app_runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/receptor"
)

type MonitorMethod int

//go:generate counterfeiter -o fake_app_runner/fake_app_runner.go . AppRunner
type AppRunner interface {
	CreateApp(params CreateAppParams) error
	SubmitLrp(lrpJSON []byte) (string, error)
	ScaleApp(name string, instances int) error
	UpdateAppRoutes(name string, routes RouteOverrides) error
	UpdateApp(updateAppParams UpdateAppParams) error
	RemoveApp(name string) error
}

//go:generate counterfeiter -o fake_keygen/fake_keygen.go . KeyGenerator
type KeyGenerator interface {
	GenerateRSAPrivateKey(bits int) (pemEncodedPrivateKey string, err error)
	GenerateRSAKeyPair(bits int) (pemEncodedPrivateKey string, authorizedKey string, err error)
}

type MonitorConfig struct {
	Method            MonitorMethod
	URI               string
	Port              uint16
	Timeout           time.Duration
	CustomCommand     string
	CustomCommandArgs []string
}

type RouteOverrides []RouteOverride

type RouteOverride struct {
	HostnamePrefix string
	Port           uint16
}

type TcpRoutes []TcpRoute

type TcpRoute struct {
	ExternalPort uint16
	Port         uint16
}

type AppEnvironmentParams struct {
	EnvironmentVariables map[string]string
	Privileged           bool
	User                 string
	Monitor              MonitorConfig
	Instances            int
	CPUWeight            uint
	MemoryMB             int
	DiskMB               int
	ExposedPorts         []uint16
	WorkingDir           string
	RouteOverrides       RouteOverrides
	TcpRoutes            TcpRoutes
	NoRoutes             bool
}

type CreateAppParams struct {
	AppEnvironmentParams

	Name         string
	StartCommand string
	RootFS       string
	AppArgs      []string
	Timeout      time.Duration
	Annotation   string
	Setup        *models.Action
}

type UpdateAppParams struct {
	Name           string
	RouteOverrides RouteOverrides
	TcpRoutes      TcpRoutes
	NoRoutes       bool
}

const (
	NoMonitor MonitorMethod = iota
	PortMonitor
	URLMonitor
	CustomMonitor

	AttemptedToCreateLatticeDebugErrorMessage = reserved_app_ids.LatticeDebugLogStreamAppId + " is a reserved app name. It is used internally to stream debug logs for lattice components."
)

const (
	healthcheckDownloadUrl string = "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz"
	lrpDomain              string = "lattice"
)

type appRunner struct {
	receptorClient receptor.Client
	systemDomain   string
	keygen         KeyGenerator
}

func New(receptorClient receptor.Client, systemDomain string, keygen KeyGenerator) AppRunner {
	return &appRunner{receptorClient, systemDomain, keygen}
}

func (appRunner *appRunner) CreateApp(params CreateAppParams) error {
	if params.Name == reserved_app_ids.LatticeDebugLogStreamAppId {
		return errors.New(AttemptedToCreateLatticeDebugErrorMessage)
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

func (appRunner *appRunner) SubmitLrp(lrpJSON []byte) (string, error) {
	desiredLRP := receptor.DesiredLRPCreateRequest{}

	if err := json.Unmarshal(lrpJSON, &desiredLRP); err != nil {
		return "", err
	}

	if desiredLRP.ProcessGuid == reserved_app_ids.LatticeDebugLogStreamAppId {
		return desiredLRP.ProcessGuid, errors.New(AttemptedToCreateLatticeDebugErrorMessage)
	}

	if exists, err := appRunner.desiredLRPExists(desiredLRP.ProcessGuid); err != nil {
		return desiredLRP.ProcessGuid, err
	} else if exists {
		return desiredLRP.ProcessGuid, newExistingAppError(desiredLRP.ProcessGuid)
	}

	if err := appRunner.receptorClient.UpsertDomain(lrpDomain, 0); err != nil {
		return desiredLRP.ProcessGuid, err
	}

	return desiredLRP.ProcessGuid, appRunner.receptorClient.CreateDesiredLRP(desiredLRP)
}

func (appRunner *appRunner) ScaleApp(name string, instances int) error {
	if exists, err := appRunner.desiredLRPExists(name); err != nil {
		return err
	} else if !exists {
		return newAppNotStartedError(name)
	}

	return appRunner.updateLrpInstances(name, instances)
}

func (appRunner *appRunner) UpdateAppRoutes(name string, routes RouteOverrides) error {
	if exists, err := appRunner.desiredLRPExists(name); err != nil {
		return err
	} else if !exists {
		return newAppNotStartedError(name)
	}

	return appRunner.updateLrpRoutes(name, routes)
}

func (appRunner *appRunner) UpdateApp(params UpdateAppParams) error {
	if exists, err := appRunner.desiredLRPExists(params.Name); err != nil {
		return err
	} else if !exists {
		return newAppNotStartedError(params.Name)
	}

	routes := appRunner.buildRoutes(params.NoRoutes, params.RouteOverrides, params.TcpRoutes)
	err := appRunner.receptorClient.UpdateDesiredLRP(
		params.Name,
		receptor.DesiredLRPUpdateRequest{
			Routes: routes.RoutingInfo(),
		},
	)
	return err
}

func (appRunner *appRunner) RemoveApp(name string) error {
	if lrpExists, err := appRunner.desiredLRPExists(name); err != nil {
		return err
	} else if !lrpExists {
		return newAppNotStartedError(name)
	}

	return appRunner.receptorClient.DeleteDesiredLRP(name)
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

func (appRunner *appRunner) buildRoutes(noRoutes bool, routeOverrides RouteOverrides, tcpRoutes TcpRoutes) route_helpers.Routes {
	var routes route_helpers.Routes

	var appRoutes route_helpers.AppRoutes
	if noRoutes {
		appRoutes = route_helpers.AppRoutes{}
	} else if len(routeOverrides) > 0 {
		routeMap := make(map[uint16][]string)
		for _, override := range routeOverrides {
			routeMap[override.Port] = append(routeMap[override.Port], fmt.Sprintf("%s.%s", override.HostnamePrefix, appRunner.systemDomain))
		}
		for port, hostnames := range routeMap {
			appRoutes = append(appRoutes, route_helpers.AppRoute{
				Hostnames: hostnames,
				Port:      port,
			})
		}
	}

	var appTcpRoutes route_helpers.TcpRoutes
	if !noRoutes && len(tcpRoutes) > 0 {
		for _, tcpRoute := range tcpRoutes {
			appTcpRoutes = append(appTcpRoutes, route_helpers.TcpRoute{
				ExternalPort: tcpRoute.ExternalPort,
				Port:         tcpRoute.Port,
			})
		}
		routes.TcpRoutes = appTcpRoutes
	}

	routes.AppRoutes = appRoutes
	return routes
}

func (appRunner *appRunner) buildRoutesWithDefaults(params CreateAppParams, primaryPort uint16) route_helpers.Routes {
	routes := appRunner.buildRoutes(params.NoRoutes, params.RouteOverrides, params.TcpRoutes)

	if len(routes.AppRoutes) == 0 && len(routes.TcpRoutes) == 0 && !params.NoRoutes {
		routes.AppRoutes = route_helpers.BuildDefaultRoutingInfo(params.Name, params.ExposedPorts, primaryPort, appRunner.systemDomain)
	}

	return routes
}

func (appRunner *appRunner) desireLrp(params CreateAppParams) error {
	primaryPort := route_helpers.GetPrimaryPort(params.Monitor.Port, params.ExposedPorts)

	private, public, err := appRunner.keygen.GenerateRSAKeyPair(2048)
	if err != nil {
		return err
	}

	routes := appRunner.buildRoutesWithDefaults(params, primaryPort)
	routes.DiegoSSHRoute = &route_helpers.DiegoSSHRoute{
		Port:       2222,
		PrivateKey: private,
	}

	vcapAppURIs := []string{}
	for _, route := range routes.AppRoutes {
		vcapAppURIs = append(vcapAppURIs, route.Hostnames...)
	}

	vcapApplication := struct {
		ApplicationName string   `json:"application_name"`
		ApplicationURIs []string `json:"application_uris"`
		Name            string   `json:"name"`
		URIs            []string `json:"uris"`
		Limits          struct {
			Disk   int `json:"disk,omitempty"`
			Memory int `json:"mem,omitempty"`
		} `json:"limits,omitempty"`
	}{}

	vcapApplication.ApplicationName = params.Name
	vcapApplication.Name = params.Name
	vcapApplication.ApplicationURIs = vcapAppURIs
	vcapApplication.URIs = vcapAppURIs
	vcapApplication.Limits.Disk = params.DiskMB
	vcapApplication.Limits.Memory = params.MemoryMB

	vcapAppBytes, err := json.Marshal(vcapApplication)
	if err != nil {
		return err
	}

	envVars := buildEnvironmentVariables(params.EnvironmentVariables)
	envVars = append(envVars, receptor.EnvironmentVariable{Name: "VCAP_APPLICATION", Value: string(vcapAppBytes)})
	envVars = append(envVars, receptor.EnvironmentVariable{Name: "PORT", Value: fmt.Sprintf("%d", primaryPort)})

	if _, exists := params.EnvironmentVariables["VCAP_SERVICES"]; !exists {
		envVars = append(envVars, receptor.EnvironmentVariable{Name: "VCAP_SERVICES", Value: "{}"})
	}

	setupAction := &models.SerialAction{
		Actions: []*models.Action{
			params.Setup,
			models.WrapAction(&models.DownloadAction{
				From: "http://file_server.service.dc1.consul:8080/v1/static/diego-sshd.tgz",
				To:   "/tmp",
				User: "vcap",
			}),
		},
	}

	hostKey, err := appRunner.keygen.GenerateRSAPrivateKey(2048)
	if err != nil {
		return err
	}

	req := receptor.DesiredLRPCreateRequest{
		ProcessGuid:          params.Name,
		Domain:               lrpDomain,
		RootFS:               params.RootFS,
		Instances:            params.Instances,
		Routes:               routes.RoutingInfo(),
		CPUWeight:            params.CPUWeight,
		MemoryMB:             params.MemoryMB,
		DiskMB:               params.DiskMB,
		Privileged:           params.Privileged,
		Ports:                append(params.ExposedPorts, 2222),
		LogGuid:              params.Name,
		LogSource:            "APP",
		MetricsGuid:          params.Name,
		EnvironmentVariables: envVars,
		Annotation:           params.Annotation,
		Setup:                &models.Action{SerialAction: setupAction},
		Action: &models.Action{
			ParallelAction: &models.ParallelAction{
				Actions: []*models.Action{
					models.WrapAction(&models.RunAction{
						Path: "/tmp/diego-sshd",
						Args: []string{
							"-address=0.0.0.0:2222",
							fmt.Sprintf("-authorizedKey=%s", public),
							fmt.Sprintf("-hostKey=%s", hostKey),
						},
						Dir:  "/tmp",
						User: params.User,
					}),
					models.WrapAction(&models.RunAction{
						Path: params.StartCommand,
						Args: params.AppArgs,
						Dir:  params.WorkingDir,
						User: params.User,
					}),
				},
			},
		},
	}

	var healthCheckArgs []string
	if params.Monitor.Timeout != 0 {
		healthCheckArgs = append(healthCheckArgs, "-timeout", fmt.Sprint(params.Monitor.Timeout))
	}
	switch params.Monitor.Method {
	case PortMonitor:
		req.Monitor = &models.Action{
			RunAction: &models.RunAction{
				Path:      "/tmp/healthcheck",
				Args:      append(healthCheckArgs, "-port", fmt.Sprint(params.Monitor.Port)),
				LogSource: "HEALTH",
				User:      params.User,
			},
		}
	case URLMonitor:
		req.Monitor = &models.Action{
			RunAction: &models.RunAction{
				Path:      "/tmp/healthcheck",
				Args:      append(healthCheckArgs, "-port", fmt.Sprint(params.Monitor.Port), "-uri", params.Monitor.URI),
				LogSource: "HEALTH",
				User:      params.User,
			},
		}
	case CustomMonitor:
		req.Monitor = &models.Action{
			RunAction: &models.RunAction{
				Path:      "/bin/sh",
				Args:      []string{"-c", params.Monitor.CustomCommand},
				LogSource: "HEALTH",
				User:      params.User,
			},
		}
	}

	return appRunner.receptorClient.CreateDesiredLRP(req)
}

func (appRunner *appRunner) updateLrpInstances(name string, instances int) error {
	err := appRunner.receptorClient.UpdateDesiredLRP(
		name,
		receptor.DesiredLRPUpdateRequest{
			Instances: &instances,
		},
	)

	return err
}

func (appRunner *appRunner) updateLrpRoutes(name string, routes RouteOverrides) error {
	appRoutes := route_helpers.AppRoutes{}

	routeMap := make(map[uint16][]string)
	for _, override := range routes {
		routeMap[override.Port] = append(routeMap[override.Port], fmt.Sprintf("%s.%s", override.HostnamePrefix, appRunner.systemDomain))
	}
	for port, hostnames := range routeMap {
		appRoutes = append(appRoutes, route_helpers.AppRoute{
			Hostnames: hostnames,
			Port:      port,
		})
	}

	err := appRunner.receptorClient.UpdateDesiredLRP(
		name,
		receptor.DesiredLRPUpdateRequest{
			Routes: appRoutes.RoutingInfo(),
		},
	)

	return err
}

func buildEnvironmentVariables(environmentVariables map[string]string) []receptor.EnvironmentVariable {
	appEnvVars := make([]receptor.EnvironmentVariable, 0, len(environmentVariables)+1)
	for name, value := range environmentVariables {
		appEnvVars = append(appEnvVars, receptor.EnvironmentVariable{Name: name, Value: value})
	}
	return appEnvVars
}
