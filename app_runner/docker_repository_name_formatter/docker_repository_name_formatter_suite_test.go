package docker_repository_name_formatter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDockerRepositoryNameFormatter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DockerRepositoryNameFormatter Suite")
}
