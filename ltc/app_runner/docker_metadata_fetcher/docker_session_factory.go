package docker_metadata_fetcher

import (
	"fmt"

	"github.com/docker/docker/registry"
	"github.com/docker/docker/utils"
)

//go:generate counterfeiter -o fake_docker_session/fake_docker_session.go . DockerSession
type DockerSession interface {
	GetRepositoryData(remote string) (*registry.RepositoryData, error)
	GetRemoteTags(registries []string, repository string, token []string) (map[string]string, error)
	GetRemoteImageJSON(imgID, registry string, token []string) ([]byte, int, error)
}

//go:generate counterfeiter -o fake_docker_session/fake_docker_session_factory.go . DockerSessionFactory
type DockerSessionFactory interface {
	MakeSession(reposName string) (DockerSession, error)
}

type dockerSessionFactory struct{}

func NewDockerSessionFactory() *dockerSessionFactory {
	return &dockerSessionFactory{}
}

func (factory *dockerSessionFactory) MakeSession(reposName string) (DockerSession, error) {
	repositoryInfo, err := registry.ParseRepositoryInfo(reposName)
	if err != nil {
		return nil, fmt.Errorf("Error resolving Docker repository name:\n" + err.Error())
	}

	endpoint, err := registry.NewEndpoint(repositoryInfo.Index)
	if err != nil {
		return nil, fmt.Errorf("Error Connecting to Docker registry:\n" + err.Error())
	}
	authConfig := &registry.AuthConfig{}
	session, error := registry.NewSession(authConfig, utils.NewHTTPRequestFactory(), endpoint, true)
	return session, error
}
