package droplet_runner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDropletRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DropletRunner Suite")
}
