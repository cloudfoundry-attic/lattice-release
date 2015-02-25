package app_examiner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAppExaminer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppExaminer Suite")
}
