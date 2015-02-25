package task_handler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTaskHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TaskHandler Suite")
}
