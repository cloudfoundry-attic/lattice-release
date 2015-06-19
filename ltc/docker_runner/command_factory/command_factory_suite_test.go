package command_factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDockerRunnerCommandFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DockerRunner CommandFactory Suite")
}
