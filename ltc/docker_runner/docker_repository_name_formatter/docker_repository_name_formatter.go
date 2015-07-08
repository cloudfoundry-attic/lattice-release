package docker_repository_name_formatter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/docker/docker/registry"
)

const (
	DockerScheme      = "docker"
	DockerIndexServer = "docker.io"
)

func FormatForReceptor(dockerPath string) (string, error) {
	return convertDockerURI(dockerPath)
}

func ParseRepoNameAndTagFromImageReference(dockerPath string) (string, string, string, error) {
	return parseDockerRepoUrl(dockerPath)
}

func convertDockerURI(dockerURI string) (string, error) {
	if strings.Contains(dockerURI, "://") {
		return "", fmt.Errorf("docker URI [%s] should not contain scheme", dockerURI)
	}

	indexName, remoteName, tag, err := parseDockerRepoUrl(dockerURI)
	if err != nil {
		return "", err
	}

	return (&url.URL{
		Scheme:   DockerScheme,
		Path:     indexName + "/" + remoteName,
		Fragment: tag,
	}).String(), nil
}

// via https://github.com/docker/docker/blob/a271eaeba224652e3a12af0287afbae6f82a9333/registry/config.go#L295
func parseDockerRepoUrl(dockerURI string) (indexName, remoteName, tag string, err error) {
	nameParts := strings.SplitN(dockerURI, "/", 2)

	if officialRegistry(nameParts) {
		// URI without host
		indexName = ""
		remoteName = dockerURI

		// URI has format docker.io/<path>
		if nameParts[0] == DockerIndexServer {
			indexName = DockerIndexServer
			remoteName = nameParts[1]
		}

		// Remote name contain no '/' - prefix it with "library/"
		// via https://github.com/docker/docker/blob/a271eaeba224652e3a12af0287afbae6f82a9333/registry/config.go#L343
		if strings.IndexRune(remoteName, '/') == -1 {
			remoteName = "library/" + remoteName
		}
	} else {
		indexName = nameParts[0]
		remoteName = nameParts[1]
	}

	remoteName, tag = parseDockerRepositoryTag(remoteName)

	_, err = registry.ParseRepositoryInfo(remoteName)
	if err != nil {
		return "", "", "", err
	}

	return indexName, remoteName, tag, nil
}

func officialRegistry(nameParts []string) bool {
	return len(nameParts) == 1 ||
		nameParts[0] == DockerIndexServer ||
		(!strings.Contains(nameParts[0], ".") &&
			!strings.Contains(nameParts[0], ":") &&
			nameParts[0] != "localhost")
}

// via https://github.com/docker/docker/blob/4398108/pkg/parsers/parsers.go#L72
func parseDockerRepositoryTag(remoteName string) (string, string) {
	n := strings.LastIndex(remoteName, ":")
	if n < 0 {
		return remoteName, "latest" // before:  remoteName, ""
	}

	if tag := remoteName[n+1:]; !strings.Contains(tag, "/") {
		return remoteName[:n], tag
	}

	return remoteName, ""
}
