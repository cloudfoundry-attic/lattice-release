package matchers_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
)

var _ = Describe("BeExactlyNilMatcher", func() {
    It("should succeed when passed nil", func() {
        Expect(nil).To(matchers.BeExactlyNil())
    })

    It("should fail when passed a typed nil", func() {
        var a []int
        Expect(a).ToNot(matchers.BeExactlyNil())
    })

    It("should fail when passing nil pointer", func() {
        var f *struct{}
        Expect(f).ToNot(matchers.BeExactlyNil())
    })

    It("should fail when passing nil channel", func() {
        var c chan struct{}
        Expect(c).ToNot(matchers.BeExactlyNil())
    })

    It("should fail when not passed nil", func() {
        Expect(0).ToNot(matchers.BeExactlyNil())
        Expect(false).ToNot(matchers.BeExactlyNil())
        Expect("").ToNot(matchers.BeExactlyNil())
    })

    Context("when the matcher fails", func() {
        It("reports the failure message", func() {
            failures := InterceptGomegaFailures(func() {
                Expect(uint16(44)).To(matchers.BeExactlyNil())
            })
            Expect(failures[0]).To(Equal("Expected 44 to be exactly nil (not just a nil pointer or some other nil-like thing)"))
        })

        It("reports the negated failure message", func() {
            failures := InterceptGomegaFailures(func() {
                Expect(nil).NotTo(matchers.BeExactlyNil())
            })
            Expect(failures[0]).To(Equal("Expected <nil> not to be exactly nil, but it really was nil (not just a nil pointer or some other nil-like thing)!"))
        })
    })
})
