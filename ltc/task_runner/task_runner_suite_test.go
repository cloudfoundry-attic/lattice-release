package task_runner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTaskRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TaskRunner Suite")
}
