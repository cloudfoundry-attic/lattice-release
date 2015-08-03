package models_test

import (
	. "github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	Describe("Equal", func() {
		It("is true when the types are the same", func() {
			err1 := &Error{Type: "a", Message: "some-message"}
			err2 := &Error{Type: "a", Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeTrue())
		})

		It("is false when the types are different", func() {
			err1 := &Error{Type: "a", Message: "some-message"}
			err2 := &Error{Type: "b", Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeFalse())
		})

		It("is false when one is nil", func() {
			var err1 *Error = nil
			err2 := &Error{Type: "b", Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeFalse())
		})

		It("is true when both errors are nil", func() {
			var err1 *Error = nil
			var err2 *Error = nil
			Expect(err1.Equal(err2)).To(BeTrue())
		})
	})
})
