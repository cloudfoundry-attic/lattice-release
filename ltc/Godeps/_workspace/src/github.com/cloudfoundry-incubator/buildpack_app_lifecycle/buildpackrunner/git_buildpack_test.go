package buildpackrunner_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle/buildpackrunner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GitBuildpack", func() {

	Describe("Clone", func() {
		var cloneTarget string
		BeforeEach(func() {
			var err error
			cloneTarget, err = ioutil.TempDir(tmpDir, "clone")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(cloneTarget)
		})

		It("clones a URL", func() {
			err := buildpackrunner.GitClone(gitUrl, cloneTarget)
			Expect(err).NotTo(HaveOccurred())
			Expect(currentBranch(cloneTarget)).To(Equal("master"))
		})

		It("clones a URL with a branch", func() {
			branchUrl := gitUrl
			branchUrl.Fragment = "a_branch"
			err := buildpackrunner.GitClone(branchUrl, cloneTarget)
			Expect(err).NotTo(HaveOccurred())
			Expect(currentBranch(cloneTarget)).To(Equal("a_branch"))
		})

		It("clones a URL with a lightweight tag", func() {
			branchUrl := gitUrl
			branchUrl.Fragment = "a_lightweight_tag"
			err := buildpackrunner.GitClone(branchUrl, cloneTarget)
			Expect(err).NotTo(HaveOccurred())
			Expect(currentBranch(cloneTarget)).To(Equal("a_lightweight_tag"))
		})

		Context("with bogus git URLs", func() {
			It("returns an error", func() {
				By("passing an invalid path", func() {
					badUrl := gitUrl
					badUrl.Path = "/a/bad/path"
					err := buildpackrunner.GitClone(badUrl, cloneTarget)
					Expect(err).To(HaveOccurred())
				})

				By("passing a bad tag/branch", func() {
					badUrl := gitUrl
					badUrl.Fragment = "notfound"
					err := buildpackrunner.GitClone(badUrl, cloneTarget)
					Expect(err).To(HaveOccurred())
				})
			})
		})

	})
})

func currentBranch(gitDir string) string {
	cmd := exec.Command("git", "symbolic-ref", "--short", "-q", "HEAD")
	cmd.Dir = gitDir
	bytes, err := cmd.Output()
	if err != nil {
		// try the tag
		cmd := exec.Command("git", "name-rev", "--name-only", "--tags", "HEAD")
		cmd.Dir = gitDir
		bytes, err = cmd.Output()
	}
	Expect(err).NotTo(HaveOccurred())
	return strings.TrimSpace(string(bytes))
}
