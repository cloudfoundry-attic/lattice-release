package docker_app_runner

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_repository_name_formatter"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

//go:generate counterfeiter -o fake_docker_app_runner/fake_docker_app_runner.go . DockerAppRunner
type DockerAppRunner interface {
	CreateDockerApp(params app_runner.CreateAppParams) error
}

type dockerAppRunner struct {
	appRunner app_runner.AppRunner
}

func New(appRunner app_runner.AppRunner) DockerAppRunner {
	return &dockerAppRunner{appRunner}
}

func (dockerAppRunner *dockerAppRunner) CreateDockerApp(params app_runner.CreateAppParams) error {
	params.GetRootFS = func() (string, error) {
		return docker_repository_name_formatter.FormatForReceptor(params.RootFS)
	}

	params.GetSetupAction = func() models.Action {
		return &models.DownloadAction{
			From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
			To:   "/tmp",
		}
	}

	return dockerAppRunner.appRunner.CreateApp(params)
}
