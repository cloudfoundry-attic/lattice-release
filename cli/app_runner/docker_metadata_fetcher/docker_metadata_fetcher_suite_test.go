package docker_metadata_fetcher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDockerMetadataFetcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DockerMetadataFetcher Suite")
}
