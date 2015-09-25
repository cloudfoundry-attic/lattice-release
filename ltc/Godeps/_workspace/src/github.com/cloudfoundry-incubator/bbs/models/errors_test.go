package models_test

import (
	"errors"

	. "github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	Describe("ConvertError", func() {
		It("maintains nils", func() {
			var err error = nil
			bbsError := ConvertError(err)
			Expect(bbsError).To(BeNil())
			Expect(bbsError == nil).To(BeTrue())
		})

		It("can convert a *Error back to *Error", func() {
			var err error = NewError(Error_ResourceConflict, "some message")
			bbsError := ConvertError(err)
			Expect(bbsError.Type).To(Equal(Error_ResourceConflict))
			Expect(bbsError.Message).To(Equal("some message"))
		})

		It("can convert a regular error to a *Error with unknown type", func() {
			var err error = errors.New("fail")
			bbsError := ConvertError(err)
			Expect(bbsError.Type).To(Equal(Error_UnknownError))
			Expect(bbsError.Message).To(Equal("fail"))
		})
	})

	Describe("Equal", func() {
		It("is true when the types are the same", func() {
			err1 := &Error{Type: 0, Message: "some-message"}
			err2 := &Error{Type: 0, Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeTrue())
		})

		It("is false when the types are different", func() {
			err1 := &Error{Type: 0, Message: "some-message"}
			err2 := &Error{Type: 1, Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeFalse())
		})

		It("is false when one is nil", func() {
			var err1 *Error = nil
			err2 := &Error{Type: 0, Message: "some-other-message"}
			Expect(err1.Equal(err2)).To(BeFalse())
		})

		It("is true when both errors are nil", func() {
			var err1 *Error = nil
			var err2 *Error = nil
			Expect(err1.Equal(err2)).To(BeTrue())
		})
	})
})
