package task_examiner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTaskExaminer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TaskExaminer Suite")
}
