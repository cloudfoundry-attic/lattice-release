package logs_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogsHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogsHelpers Suite")
}
