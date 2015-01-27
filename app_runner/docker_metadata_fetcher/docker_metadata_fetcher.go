package docker_metadata_fetcher

//go:generate counterfeiter -o fake_docker_metadata_fetcher/fake_docker_metadata_fetcher.go . DockerMetadataFetcher

import (
	"fmt"
	"github.com/docker/docker/image"
)

type ImageMetadata struct {
	WorkingDir   string
	StartCommand []string
}

type DockerMetadataFetcher interface {
	FetchMetadata(repoName string, tag string) (*ImageMetadata, error)
}

type dockerMetadataFetcher struct {
	dockerSessionFactory DockerSessionFactory
}

func New(sessionFactory DockerSessionFactory) DockerMetadataFetcher {
	return &dockerMetadataFetcher{
		dockerSessionFactory: sessionFactory,
	}
}

func (fetcher *dockerMetadataFetcher) FetchMetadata(repoName string, tag string) (*ImageMetadata, error) {

	session, err := fetcher.dockerSessionFactory.MakeSession(repoName)
	if err != nil {
		return nil, err
	}

	repoData, err := session.GetRepositoryData(repoName)
	if err != nil {
		return nil, err
	}

	tagsList, err := session.GetRemoteTags(repoData.Endpoints, repoName, repoData.Tokens)
	if err != nil {
		return nil, err
	}

	imgID, ok := tagsList[tag]
	if !ok {
		return nil, fmt.Errorf("Unknown tag: %s:%s", repoName, tag)
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
	return &ImageMetadata{WorkingDir: img.Config.WorkingDir, StartCommand: startCommand}, nil
}
