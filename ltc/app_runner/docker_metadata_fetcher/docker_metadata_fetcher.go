package docker_metadata_fetcher

import (
	"fmt"
	"sort"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_repository_name_formatter"
	"github.com/docker/docker/image"
	"github.com/docker/docker/nat"
)

type ImageMetadata struct {
	WorkingDir   string
	Ports        docker_app_runner.PortConfig
	StartCommand []string
}

//go:generate counterfeiter -o fake_docker_metadata_fetcher/fake_docker_metadata_fetcher.go . DockerMetadataFetcher
type DockerMetadataFetcher interface {
	FetchMetadata(dockerImageReference string) (*ImageMetadata, error)
}

type dockerMetadataFetcher struct {
	dockerSessionFactory DockerSessionFactory
}

func New(sessionFactory DockerSessionFactory) DockerMetadataFetcher {
	return &dockerMetadataFetcher{
		dockerSessionFactory: sessionFactory,
	}
}

func (fetcher *dockerMetadataFetcher) FetchMetadata(dockerImageReference string) (*ImageMetadata, error) {

	indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)
	if err != nil {
		return nil, err
	}

	var reposName string
	if len(indexName) > 0 {
		reposName = fmt.Sprintf("%s/%s", indexName, remoteName)
	} else {
		reposName = remoteName
	}

	session, err := fetcher.dockerSessionFactory.MakeSession(reposName)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Error parsing remote image json for specified docker image:\n%s", err.Error())
	}

	if img.Config == nil {
		return nil, fmt.Errorf("Parsing start command failed")
	}

	startCommand := append(img.Config.Entrypoint, img.Config.Cmd...)

	uintExposedPorts := sortPorts(img.ContainerConfig.ExposedPorts)
	var monitoredPort uint16

	if len(uintExposedPorts) > 0 {
		monitoredPort = uintExposedPorts[0]
	}

	return &ImageMetadata{
		WorkingDir:   img.Config.WorkingDir,
		StartCommand: startCommand,
		Ports: docker_app_runner.PortConfig{
			Monitored: monitoredPort,
			Exposed:   uintExposedPorts,
		},
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

	uintExposedPorts := make([]uint16, 0)
	for _, port := range intPorts {
		uintExposedPorts = append(uintExposedPorts, uint16(port))
	}
	return uintExposedPorts
}
