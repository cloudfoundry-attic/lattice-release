package docker_metadata_fetcher

//go:generate counterfeiter -o fake_docker_session/fake_docker_session.go . DockerSession
//go:generate counterfeiter -o fake_docker_session/fake_docker_session_factory.go . DockerSessionFactory

import (
	"fmt"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/utils"
)

type DockerSession interface {
	GetRepositoryData(remote string) (*registry.RepositoryData, error)
	GetRemoteTags(registries []string, repository string, token []string) (map[string]string, error)
	GetRemoteImageJSON(imgID, registry string, token []string) ([]byte, int, error)
}

type DockerSessionFactory interface {
	MakeSession(repoName string) (DockerSession, error)
}

type dockerSessionFactory struct{}

func NewDockerSessionFactory() *dockerSessionFactory {
	return &dockerSessionFactory{}
}

func (factory *dockerSessionFactory) MakeSession(repoName string) (DockerSession, error) {
	hostname, repoName, err := registry.ResolveRepositoryName(repoName)
	if err != nil {
		return nil, fmt.Errorf("Error resolving Docker repository name:\n" + err.Error())
	}

	endpoint, err := registry.NewEndpoint(hostname, []string{"127.0.0.1/32"})
	if err != nil {
		return nil, fmt.Errorf("Error Connecting to Docker registry:\n" + err.Error())
	}
	authConfig := &registry.AuthConfig{}
	session, error := registry.NewSession(authConfig, utils.NewHTTPRequestFactory(), endpoint, true)
	return session, error
}
