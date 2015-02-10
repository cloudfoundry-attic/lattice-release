package docker_repository_name_formatter

import (
	"github.com/docker/docker/registry"
	"strings"
)

func FormatForReceptor(dockerImageReference string) (string, error) {
	dockerRepositoryName, tag := ParseRepoNameAndTagFromImageReference(dockerImageReference)

	_, err := registry.ParseRepositoryInfo(dockerRepositoryName)
	if err != nil {
		return "", err
	}

	if strings.Contains(dockerRepositoryName, "/") {
		return "docker:///" + dockerRepositoryName + "#" + tag, nil
	} else {
		return "docker:///library/" + dockerRepositoryName + "#" + tag, nil
	}
}

func ParseRepoNameAndTagFromImageReference(dockerImageReference string) (string, string) {
	imageWithTag := strings.Split(dockerImageReference, ":")
	dockerRepositoryName := imageWithTag[0]

	tag := "latest"
	if len(imageWithTag) > 1 {
		tag = imageWithTag[1]
	}
	return dockerRepositoryName, tag
}
