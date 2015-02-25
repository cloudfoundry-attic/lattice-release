package matchers_test

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type woohoo struct {
	Flag bool
}

type woomap map[woohoo]string

var _ = Describe("ContainExactlyMatcher", func() {
	It("matches if the array contains exactly the elements in the expected array, but is order independent.", func() {

		Expect([]string{"hi there", "ho there", "hallo"}).To(matchers.ContainExactly([]string{"hi there", "ho there", "hallo"}))
		Expect([]string{"hi there", "ho there", "hallo"}).To(matchers.ContainExactly([]string{"ho there", "hallo", "hi there"}))
		Expect([]woohoo{woohoo{Flag: true}}).To(matchers.ContainExactly([]woohoo{woohoo{Flag: true}}))
		Expect([]woohoo{woohoo{Flag: true}, woohoo{Flag: false}}).To(matchers.ContainExactly([]woohoo{woohoo{Flag: true}, woohoo{Flag: false}}))

		Expect([]string{"hi there", "ho there", "hallo"}).ToNot(matchers.ContainExactly([]string{"hi there", "bye bye"}))
		Expect([]string{"hi there", "ho there", "hallo"}).ToNot(matchers.ContainExactly([]string{"ho there", "hi there"}))
		Expect([]string{"hi there", "ho there", "hallo"}).ToNot(matchers.ContainExactly([]string{"buhbye"}))
		Expect([]string{"hi there", "ho there", "hallo"}).ToNot(matchers.ContainExactly([]string{}))

		Expect([]woohoo{woohoo{Flag: false}}).ToNot(matchers.ContainExactly([]woohoo{woohoo{Flag: true}}))
		Expect([]woohoo{woohoo{Flag: false}, woohoo{Flag: false}}).ToNot(matchers.ContainExactly([]woohoo{woohoo{Flag: true}, woohoo{Flag: false}}))
	})

	It("handles map types", func() {
		Expect(woomap{woohoo{true}: "fun", woohoo{false}: "not fun"}).To(matchers.ContainExactly(woomap{woohoo{false}: "not fun", woohoo{true}: "fun"}))
	})

	It("handles duplicate elements", func() {
		Expect([]int{-7, -7, 9, 4}).To(matchers.ContainExactly([]int{4, 9, -7, -7}))
		Expect([]int{-7, -7, 9, 4}).ToNot(matchers.ContainExactly([]int{4, 9, -7, 44}))
		Expect([]int{4, -7, 9, 44}).ToNot(matchers.ContainExactly([]int{4, 9, -7, -7}))
	})

	It("fails for non-array or slices", func() {
		failures := InterceptGomegaFailures(func() {
			Expect([]string{"hi there", "ho there", "hallo"}).ToNot(matchers.ContainExactly(46))
			Expect(23).ToNot(matchers.ContainExactly([]string{"hi there", "ho there", "hallo"}))
			Expect("woo").ToNot(matchers.ContainExactly([]woohoo{woohoo{Flag: true}, woohoo{Flag: false}}))
		})
		Expect(failures[0]).To(Equal("Matcher can only take an array or slice"))
		Expect(failures[1]).To(Equal("Matcher can only take an array or slice"))
		Expect(failures[2]).To(Equal("Matcher can only take an array or slice"))
	})

})
