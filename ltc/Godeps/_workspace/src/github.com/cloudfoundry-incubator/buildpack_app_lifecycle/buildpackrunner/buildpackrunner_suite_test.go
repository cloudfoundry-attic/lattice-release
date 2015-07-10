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

var buildpackDir string
var fileGitUrl url.URL
var gitUrl url.URL
var httpServer *httptest.Server
var tmpDir string

var _ = SynchronizedBeforeSuite(func() []byte {
	gitPath, err := exec.LookPath("git")
	Expect(err).NotTo(HaveOccurred())

	tmpDir, err = ioutil.TempDir("", "tmpDir")
	Expect(err).NotTo(HaveOccurred())
	buildpackDir := filepath.Join(tmpDir, "fake-buildpack")
	err = os.MkdirAll(buildpackDir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	submoduleDir := filepath.Join(tmpDir, "submodule")
	err = os.MkdirAll(submoduleDir, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	execute(buildpackDir, "rm", "-rf", ".git")
	execute(buildpackDir, gitPath, "init")
	execute(buildpackDir, gitPath, "config", "user.email", "you@example.com")
	execute(buildpackDir, gitPath, "config", "user.name", "your name")
	writeFile(filepath.Join(buildpackDir, "content"), "some content")

	execute(submoduleDir, "rm", "-rf", ".git")
	execute(submoduleDir, gitPath, "init")
	execute(submoduleDir, gitPath, "config", "user.email", "you@example.com")
	execute(submoduleDir, gitPath, "config", "user.name", "your name")
	writeFile(filepath.Join(submoduleDir, "README"), "1st commit")
	execute(submoduleDir, gitPath, "add", ".")
	execute(submoduleDir, gitPath, "commit", "-am", "first commit")
	writeFile(filepath.Join(submoduleDir, "README"), "2nd commit")
	execute(submoduleDir, gitPath, "commit", "-am", "second commit")

	execute(buildpackDir, gitPath, "submodule", "add", "file://"+submoduleDir, "sub")
	execute(buildpackDir+"/sub", gitPath, "checkout", "HEAD^")
	execute(buildpackDir, gitPath, "add", "-A")
	execute(buildpackDir, gitPath, "commit", "-m", "fake commit")
	execute(buildpackDir, gitPath, "commit", "--allow-empty", "-m", "empty commit")
	execute(buildpackDir, gitPath, "tag", "-m", "annotated tag", "a_tag")
	execute(buildpackDir, gitPath, "tag", "a_lightweight_tag")
	execute(buildpackDir, gitPath, "checkout", "-b", "a_branch")
	execute(buildpackDir+"/sub", gitPath, "checkout", "master")
	execute(buildpackDir, gitPath, "add", "-A")
	execute(buildpackDir, gitPath, "commit", "-am", "update submodule")
	execute(buildpackDir, gitPath, "checkout", "master")
	execute(buildpackDir, gitPath, "update-server-info")

	return []byte(string(tmpDir))

}, func(data []byte) {
	tmpDir = string(data)
	httpServer = httptest.NewServer(http.FileServer(http.Dir(tmpDir)))

	gitUrl = url.URL{
		Scheme: "http",
		Host:   httpServer.Listener.Addr().String(),
		Path:   "/fake-buildpack/.git",
	}

	fileGitUrl = url.URL{
		Scheme: "file",
		Path:   tmpDir + "/fake-buildpack",
	}
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

func writeFile(filepath, content string) {
	err := ioutil.WriteFile(filepath,
		[]byte(content), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())
}
