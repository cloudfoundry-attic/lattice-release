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
