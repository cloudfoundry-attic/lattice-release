package cf_ignore_test

import (
	"errors"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ignore", func() {
	var cfIgnore cf_ignore.CFIgnore

	BeforeEach(func() {
		cfIgnore = cf_ignore.New()
	})

	It("excludes files based on exact path matches", func() {
		Expect(cfIgnore.Parse(strings.NewReader("the-dir/the-path"))).To(Succeed())
		Expect(cfIgnore.ShouldIgnore("the-dir/the-path")).To(BeTrue())
	})

	It("excludes the contents of directories based on exact path matches", func() {
		Expect(cfIgnore.Parse(strings.NewReader("dir1/dir2"))).To(Succeed())
		Expect(cfIgnore.ShouldIgnore("dir1/dir2/the-file")).To(BeTrue())
		Expect(cfIgnore.ShouldIgnore("dir1/dir2/dir3/the-file")).To(BeTrue())
	})

	It("excludes files based on star patterns", func() {
		Expect(cfIgnore.Parse(strings.NewReader("dir1/*.so"))).To(Succeed())
		Expect(cfIgnore.ShouldIgnore("dir1/file1.so")).To(BeTrue())
		Expect(cfIgnore.ShouldIgnore("dir1/file2.cc")).To(BeFalse())
	})

	It("excludes files based on double-star patterns", func() {
		Expect(cfIgnore.Parse(strings.NewReader(`dir1/**/*.so`))).To(Succeed())
		Expect(cfIgnore.ShouldIgnore("dir1/dir2/dir3/file1.so")).To(BeTrue())
		Expect(cfIgnore.ShouldIgnore("different-dir/dir2/file.so")).To(BeFalse())
	})

	It("allows files to be explicitly included", func() {
		err := cfIgnore.Parse(strings.NewReader(`
			node_modules/*
			!node_modules/common
		`))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfIgnore.ShouldIgnore("node_modules/something-else")).To(BeTrue())
		Expect(cfIgnore.ShouldIgnore("node_modules/common")).To(BeFalse())
	})

	It("applies the patterns in order from top to bottom", func() {
		err := cfIgnore.Parse(strings.NewReader(`
			stuff/*
			!stuff/*.c
			stuff/exclude.c
		`))
		Expect(err).NotTo(HaveOccurred())

		Expect(cfIgnore.ShouldIgnore("stuff/something.txt")).To(BeTrue())
		Expect(cfIgnore.ShouldIgnore("stuff/exclude.c")).To(BeTrue())
		Expect(cfIgnore.ShouldIgnore("stuff/include.c")).To(BeFalse())
	})

	It("ignores certain commonly ingored files by default", func() {
		Expect(cfIgnore.Parse(strings.NewReader(""))).To(Succeed())
		Expect(cfIgnore.ShouldIgnore(".git/objects")).To(BeTrue())

		Expect(cfIgnore.Parse(strings.NewReader("!.git"))).To(Succeed())
		Expect(cfIgnore.ShouldIgnore(".git/objects")).To(BeFalse())
	})

	Describe("files named manifest.yml", func() {
		It("ignores manifest.yml at the top level", func() {
			Expect(cfIgnore.Parse(strings.NewReader(""))).To(Succeed())
			Expect(cfIgnore.ShouldIgnore("manifest.yml")).To(BeTrue())
		})

		It("does not ignore nested manifest.yml files", func() {
			Expect(cfIgnore.Parse(strings.NewReader(""))).To(Succeed())
			Expect(cfIgnore.ShouldIgnore("public/assets/manifest.yml")).To(BeFalse())
		})
	})

	Describe("Parse", func() {
		Context("when the ignored reader returns an error", func() {
			It("returns an error", func() {
				err := cfIgnore.Parse(errorReader{})
				Expect(err).To(MatchError("no bueno"))
			})
		})

		Context("when parse is called multiple times", func() {
			It("adds patterns to the existing ignore patterns", func() {
				Expect(cfIgnore.Parse(strings.NewReader("dir2"))).To(Succeed())
				Expect(cfIgnore.Parse(strings.NewReader("dir4"))).To(Succeed())
				Expect(cfIgnore.ShouldIgnore("dir1/dir2")).To(BeTrue())
				Expect(cfIgnore.ShouldIgnore("dir3/dir4")).To(BeTrue())
			})
		})
	})
})

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("no bueno")
}
