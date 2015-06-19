package app_runner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAppRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppRunner Suite")
}
