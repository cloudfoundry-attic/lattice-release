package models_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ValidatorErrorCase struct {
	Message string
	models.Validator
}

func testValidatorErrorCase(testCase ValidatorErrorCase) {
	message := testCase.Message
	invalid := testCase.Validator

	Context("when invalid", func() {
		It("returns an error indicating '"+message+"'", func() {
			err := invalid.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(message))
		})
	})
}

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}
