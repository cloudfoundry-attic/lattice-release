package docker_metadata_fetcher

//go:generate counterfeiter -o fake_docker_metadata_fetcher/fake_docker_metadata_fetcher.go . DockerMetadataFetcher

import (
	"fmt"

	"github.com/docker/docker/image"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/utils"
)

type ImageMetadata struct {
	WorkingDir   string
	StartCommand []string
}

type DockerMetadataFetcher interface {
	FetchMetadata(repoName string, tag string) (*ImageMetadata, error)
}

type dockerMetadataFetcher struct {
}

func New() DockerMetadataFetcher {
	return &dockerMetadataFetcher{}
}

func (fetcher *dockerMetadataFetcher) FetchMetadata(repoName string, tag string) (*ImageMetadata, error) {
	hostname, repoName, err := registry.ResolveRepositoryName(repoName)
	if err != nil {
		return nil, err
	}

	endpoint, err := registry.NewEndpoint(hostname, []string{"127.0.0.1/32"})
	if err != nil {
		return nil, err
	}

	authConfig := &registry.AuthConfig{}
	session, err := registry.NewSession(authConfig, utils.NewHTTPRequestFactory(), endpoint, true)
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
	for _, endpoint := range repoData.Endpoints {
		imgJSON, _, err := session.GetRemoteImageJSON(imgID, endpoint, repoData.Tokens)
		if err == nil {
			img, err = image.NewImgJSON(imgJSON)
			if err != nil {
				return nil, err
			}
		}
	}

	if img.Config == nil {
		return nil, fmt.Errorf("Parsing start command failed")
	}

	startCommand := append(img.Config.Entrypoint, img.Config.Cmd...)
	return &ImageMetadata{WorkingDir: img.Config.WorkingDir, StartCommand: startCommand}, nil
}
