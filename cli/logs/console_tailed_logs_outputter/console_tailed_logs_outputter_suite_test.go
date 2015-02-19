package console_tailed_logs_outputter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConsoleTailedLogOutputter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConsoleTailedLogsOutputter Suite")
}
