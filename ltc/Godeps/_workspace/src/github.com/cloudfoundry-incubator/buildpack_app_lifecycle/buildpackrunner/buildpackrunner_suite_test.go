package buildpackrunner_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestBuildpackrunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpackrunner Suite")
}

var tmpDir string
var httpServer *httptest.Server
var gitUrl url.URL

var _ = SynchronizedBeforeSuite(func() []byte {
	gitPath, err := exec.LookPath("git")
	Expect(err).NotTo(HaveOccurred())

	tmpDir, err = ioutil.TempDir("", "tmpDir")
	Expect(err).NotTo(HaveOccurred())
	buildpackDir := filepath.Join(tmpDir, "fake-buildpack")
	err = os.MkdirAll(buildpackDir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	execute(buildpackDir, "rm", "-rf", ".git")
	execute(buildpackDir, gitPath, "init")
	execute(buildpackDir, gitPath, "config", "user.email", "you@example.com")
	execute(buildpackDir, gitPath, "config", "user.name", "your name")

	err = ioutil.WriteFile(filepath.Join(buildpackDir, "content"),
		[]byte("some content"), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	execute(buildpackDir, gitPath, "add", ".")
	execute(buildpackDir, gitPath, "add", "-A")
	execute(buildpackDir, gitPath, "commit", "-am", "fake commit")
	execute(buildpackDir, gitPath, "branch", "a_branch")
	execute(buildpackDir, gitPath, "tag", "-m", "annotated tag", "a_tag")
	execute(buildpackDir, gitPath, "tag", "a_lightweight_tag")
	execute(buildpackDir, gitPath, "update-server-info")

	httpServer = httptest.NewServer(http.FileServer(http.Dir(tmpDir)))

	gitUrl = url.URL{
		Scheme: "http",
		Host:   httpServer.Listener.Addr().String(),
		Path:   "/fake-buildpack/.git",
	}
	return []byte(gitUrl.String())
}, func(data []byte) {
	u, err := url.Parse(string(data))
	Expect(err).NotTo(HaveOccurred())
	gitUrl = *u
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	httpServer.Close()
	os.RemoveAll(tmpDir)
})

func execute(dir string, execCmd string, args ...string) {
	cmd := exec.Command(execCmd, args...)
	cmd.Dir = dir
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())
}
