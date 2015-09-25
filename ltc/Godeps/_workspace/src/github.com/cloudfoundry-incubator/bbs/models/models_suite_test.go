package models_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/bbs/format"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ValidatorErrorCase struct {
	Message string
	format.Versioner
}

func testValidatorErrorCase(testCase ValidatorErrorCase) {
	message := testCase.Message
	model := testCase.Versioner

	Context("when invalid", func() {
		It("returns an error indicating '"+message+"'", func() {
			err := model.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		})
	})
}

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}
