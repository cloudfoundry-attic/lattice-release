package docker_metadata_fetcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/docker_runner/docker_repository_name_formatter"
	"github.com/docker/docker/image"
	"github.com/docker/docker/nat"
)

type ImageMetadata struct {
	WorkingDir   string
	ExposedPorts []uint16
	StartCommand []string
}

//go:generate counterfeiter -o fake_docker_metadata_fetcher/fake_docker_metadata_fetcher.go . DockerMetadataFetcher
type DockerMetadataFetcher interface {
	FetchMetadata(dockerPath string) (*ImageMetadata, error)
}

type dockerMetadataFetcher struct {
	dockerSessionFactory DockerSessionFactory
}

func New(sessionFactory DockerSessionFactory) DockerMetadataFetcher {
	return &dockerMetadataFetcher{
		dockerSessionFactory: sessionFactory,
	}
}

func (fetcher *dockerMetadataFetcher) FetchMetadata(dockerPath string) (*ImageMetadata, error) {
	indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerPath)
	if err != nil {
		return nil, err
	}

	var reposName string
	if len(indexName) > 0 {
		reposName = fmt.Sprintf("%s/%s", indexName, remoteName)
	} else {
		reposName = remoteName
	}

	var session DockerSession
	if session, err = fetcher.dockerSessionFactory.MakeSession(reposName, false); err != nil {
		if !strings.Contains(err.Error(), "this private registry supports only HTTP or HTTPS with an unknown CA certificate") {
			return nil, err
		}

		session, err = fetcher.dockerSessionFactory.MakeSession(reposName, true)
		if err != nil {
			return nil, err
		}
	}

	repoData, err := session.GetRepositoryData(remoteName)
	if err != nil {
		return nil, err
	}

	tagsList, err := session.GetRemoteTags(repoData.Endpoints, remoteName, repoData.Tokens)
	if err != nil {
		return nil, err
	}

	imgID, ok := tagsList[tag]
	if !ok {
		return nil, fmt.Errorf("Unknown tag: %s:%s", remoteName, tag)
	}

	var img *image.Image
	endpoint := repoData.Endpoints[0]
	imgJSON, _, err := session.GetRemoteImageJSON(imgID, endpoint, repoData.Tokens)
	if err != nil {
		return nil, err
	}

	img, err = image.NewImgJSON(imgJSON)
	if err != nil {
		return nil, fmt.Errorf("Error parsing remote image json for specified docker image:\n%s", err)
	}
	if img.Config == nil {
		return nil, fmt.Errorf("Parsing start command failed")
	}

	startCommand := append(img.Config.Entrypoint, img.Config.Cmd...)
	exposedPorts := sortPorts(img.ContainerConfig.ExposedPorts)

	return &ImageMetadata{
		WorkingDir:   img.Config.WorkingDir,
		StartCommand: startCommand,
		ExposedPorts: exposedPorts,
	}, nil
}

func sortPorts(dockerExposedPorts map[nat.Port]struct{}) []uint16 {
	intPorts := make([]int, 0)
	for natPort, _ := range dockerExposedPorts {
		if natPort.Proto() == "tcp" {
			intPorts = append(intPorts, natPort.Int())
		}
	}
	sort.IntSlice(intPorts).Sort()

	exposedPorts := make([]uint16, 0)
	for _, port := range intPorts {
		exposedPorts = append(exposedPorts, uint16(port))
	}
	return exposedPorts
}
