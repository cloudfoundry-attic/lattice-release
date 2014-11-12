package app_runner

import (
	"fmt"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

const (
	spyDownloadUrl           string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
	repUrlRelativeToExecutor string = "http://127.0.0.1:20515"
)

type diegoAppRunner struct {
	receptorClient receptor.Client
}

func NewDiegoAppRunner(receptorClient receptor.Client) *diegoAppRunner {
	return &diegoAppRunner{receptorClient}
}

func (appRunner *diegoAppRunner) StartDockerApp(name string, startCommand string, dockerImagePath string) error {
	err := appRunner.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: name,
		Domain:      "diego-edge",
		RootFSPath:  dockerImagePath,
		Instances:   1,
		Stack:       "lucid64",
		Routes:      []string{fmt.Sprintf("%s.192.168.11.11.xip.io", name)},
		MemoryMB:    128,
		DiskMB:      1024,
		Ports: []receptor.PortMapping{
			{ContainerPort: 8080},
		},
		Log: receptor.LogConfig{
			Guid:       name,
			SourceName: "APP",
		},
		Actions: []models.ExecutorAction{
			{
				Action: models.DownloadAction{
					From:     spyDownloadUrl,
					To:       "/tmp",
					CacheKey: "",
				},
			},
			models.Parallel(
				models.ExecutorAction{
					Action: models.RunAction{
						Path: startCommand,
					},
				},
				models.ExecutorAction{
					models.MonitorAction{
						Action: models.ExecutorAction{
							models.RunAction{
								Path: "/tmp/spy",
								Args: []string{"-addr", ":8080"},
							},
						},
						HealthyThreshold:   1,
						UnhealthyThreshold: 1,
						HealthyHook: models.HealthRequest{
							Method: "PUT",
							URL: fmt.Sprintf(
								"%s/lrp_running/%s/PLACEHOLDER_INSTANCE_INDEX/PLACEHOLDER_INSTANCE_GUID",
								repUrlRelativeToExecutor,
								name,
							),
						},
					},
				},
			),
		},
	})

	return err
}
