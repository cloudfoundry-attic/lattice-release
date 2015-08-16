package cluster_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	ginkgo_config "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/nu7hatch/gouuid"
)

var numCPU int

func init() {
	numCPU = runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
}

type ClusterTestRunner interface {
	Run(timeout time.Duration, verbose bool)
}

type clusterTestRunner struct {
	testingT          GinkgoTestingT
	config            *config.Config
	latticeCliHome    string
	ltcExecutablePath string
}

type ginkgoTestingT struct{}

func (g *ginkgoTestingT) Fail() {
	os.Exit(1)
}

func forceAbs(path string) string {
	if filepath.IsAbs(path) || !strings.Contains(path, "/") {
		return path
	}

	abs, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	return abs
}

func NewClusterTestRunner(config *config.Config, latticeCliHome string) ClusterTestRunner {
	return &clusterTestRunner{
		config:            config,
		testingT:          &ginkgoTestingT{},
		latticeCliHome:    latticeCliHome,
		ltcExecutablePath: forceAbs(os.Args[0]),
	}
}

func (runner *clusterTestRunner) Run(timeout time.Duration, verbose bool) {
	ginkgo_config.DefaultReporterConfig.Verbose = verbose
	ginkgo_config.DefaultReporterConfig.SlowSpecThreshold = float64(45)
	defineTheGinkgoTests(runner, timeout)
	RegisterFailHandler(Fail)
	RunSpecs(runner.testingT, "Lattice Integration Tests")
}

func defineTheGinkgoTests(runner *clusterTestRunner, timeout time.Duration) {
	BeforeSuite(func() {
		if err := runner.config.Load(); err != nil {
			fmt.Fprintln(getStyledWriter("test"), "Error loading config")
			return
		}
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	Describe("Lattice", func() {
		Context("docker", func() {
			Context("when desiring a docker-based LRP", func() {
				var appName, appRoute string

				BeforeEach(func() {
					appGUID, err := uuid.NewV4()
					Expect(err).NotTo(HaveOccurred())

					appName = fmt.Sprintf("lattice-test-app-%s", appGUID.String())
					appRoute = fmt.Sprintf("%s.%s", appName, runner.config.Target())
				})

				AfterEach(func() {
					runner.removeApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))

					Eventually(errorCheckForRoute(appRoute), timeout, 1).Should(HaveOccurred())
				})

				It("should run a docker app", func() {
					debugLogsStream := runner.streamDebugLogs(timeout)
					defer func() { debugLogsStream.Terminate().Wait() }()

					runner.createDockerApp(timeout, appName, "cloudfoundry/lattice-app", fmt.Sprintf("--timeout=%s", timeout.String()), "--working-dir=/", "--env", "APP_NAME", "--", "/lattice-app", "--message", "Hello Lattice User", "--quiet")

					Eventually(errorCheckForRoute(appRoute), timeout, 1).ShouldNot(HaveOccurred())

					Eventually(debugLogsStream.Out, timeout).Should(gbytes.Say("rep.*cell-\\d+"))
					Eventually(debugLogsStream.Out, timeout).Should(gbytes.Say("garden-linux.*cell-\\d+"))
					debugLogsStream.Terminate().Wait()

					logsStream := runner.streamLogs(timeout, appName)
					defer func() { logsStream.Terminate().Wait() }()

					Eventually(logsStream.Out, timeout).Should(gbytes.Say("LATTICE-TEST-APP. Says Hello Lattice User."))

					runner.scaleApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))

					instanceCountChan := make(chan int, numCPU)
					go countInstances(appRoute, instanceCountChan)
					Eventually(instanceCountChan, timeout).Should(Receive(Equal(3)))
				})

				It("should run a docker app using metadata from Docker Hub", func() {
					runner.createDockerApp(timeout, appName, "cloudfoundry/lattice-app")

					Eventually(errorCheckForRoute(appRoute), timeout, .5).ShouldNot(HaveOccurred())
				})

				Context("when the docker app requires escalated privileges to run", func() {
					It("should start the nginx app successfully", func() {
						By("passing the `--run-as-root` flag to `ltc create`")
						runner.createDockerApp(timeout, appName, "cloudfoundry/lattice-app", "--run-as-root", fmt.Sprintf("--timeout=%s", timeout.String()))

						Eventually(errorCheckForRoute(appRoute), timeout, .5).ShouldNot(HaveOccurred())

						resp, err := makeGetRequestToURL(appRoute + "/env")
						Expect(err).NotTo(HaveOccurred())
						defer resp.Body.Close()
						respBytes, err := ioutil.ReadAll(resp.Body)
						Expect(err).NotTo(HaveOccurred())

						Expect(respBytes).To(MatchRegexp("<dt>USER</dt><dd>root</dd>"), "lattice-app should report running as root")
					})
				})
			})

			Context("when desiring a docker-based LRP with tcp routes", func() {
				var (
					externalPort uint16
					appName      string
				)

				BeforeEach(func() {
					externalPort = 50000
					appGUID, err := uuid.NewV4()
					Expect(err).NotTo(HaveOccurred())

					appName = fmt.Sprintf("lattice-test-app-%s", appGUID.String())
				})

				AfterEach(func() {
					runner.removeApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))
				})

				It("should run a docker app exposing tcp routes", func() {
					runner.createDockerApp(timeout, appName, "cloudfoundry/tcp-sample-receiver", fmt.Sprintf("--tcp-routes=5222:%d", externalPort), fmt.Sprintf("--timeout=%s", timeout.String()))
					Eventually(errorCheckForConnection(runner.config.Target(), externalPort), timeout, 1).ShouldNot(HaveOccurred())
				})
			})
		})

		Context("droplets", func() {
			var dropletName, appName, dropletFolderURL, appRoute string

			BeforeEach(func() {
				// Generate a droplet name up front so that it can persist across droplet tests
				dropletGUID, err := uuid.NewV4()
				Expect(err).NotTo(HaveOccurred())
				dropletName = "droplet-" + dropletGUID.String()

				appName = "running-" + dropletName

				blobTarget := runner.config.BlobStore()

				if blobTarget.Username != "" {
					dropletFolderURL = fmt.Sprintf("%s:%s@%s:%s/blobs/%s",
						blobTarget.Username,
						blobTarget.Password,
						blobTarget.Host,
						blobTarget.Port,
						dropletName)
				} else {
					dropletFolderURL = fmt.Sprintf("%s:%s/blobs/%s",
						blobTarget.Host,
						blobTarget.Port,
						dropletName)
				}

				appRoute = fmt.Sprintf("%s.%s", appName, runner.config.Target())
			})

			AfterEach(func() {
				runner.removeApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))
				Eventually(errorCheckForRoute(appRoute), timeout, .5).Should(HaveOccurred())

				runner.removeDroplet(timeout, dropletName)
				Eventually(errorCheckURLExists(dropletFolderURL+"/droplet.tgz"), timeout, 1).Should(HaveOccurred())
			})

			It("builds, lists and launches a droplet", func() {
				By("checking out lattice-app from github")
				gitDir := runner.cloneRepo(timeout, "https://github.com/pivotal-cf-experimental/lattice-app.git")
				defer os.RemoveAll(gitDir)

				By("launching a build task")
				runner.buildDroplet(timeout, dropletName, "https://github.com/cloudfoundry/go-buildpack.git", gitDir)

				By("uploading a compiled droplet to the blob store")
				Eventually(errorCheckURLExists(dropletFolderURL+"/droplet.tgz"), timeout, 1).ShouldNot(HaveOccurred())

				By("listing droplets")
				runner.listDroplets(timeout, dropletName)

				By("launching the droplet")
				runner.launchDroplet(timeout, appName, dropletName)

				Eventually(errorCheckForRoute(appRoute), timeout, .5).ShouldNot(HaveOccurred())
			})
		})
	})
}

func (runner *clusterTestRunner) cloneRepo(timeout time.Duration, repoURL string) string {
	tmpDir, err := ioutil.TempDir("", "repo")
	Expect(err).NotTo(HaveOccurred())

	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to clone %s to %s", repoURL, tmpDir)))

	command := exec.Command("/usr/bin/env", "git", "clone", repoURL, tmpDir)

	session, err := gexec.Start(command, getStyledWriter("git-clone"), getStyledWriter("git-clone"))
	Expect(err).NotTo(HaveOccurred())
	expectExitInBuffer(timeout, session, session.Err)

	Eventually(session.Err).Should(gbytes.Say(fmt.Sprintf("Cloning into '%s'...", tmpDir)))

	fmt.Fprintf(getStyledWriter("test"), "Cloned %s into %s\n", repoURL, tmpDir)

	return tmpDir
}

func (runner *clusterTestRunner) buildDroplet(timeout time.Duration, dropletName, buildpack, srcDir string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Submitting build of %s with buildpack %s", dropletName, buildpack)))

	command := runner.command("build-droplet", dropletName, buildpack, "--timeout", timeout.String())
	command.Dir = srcDir

	session, err := gexec.Start(command, getStyledWriter("build-droplet"), getStyledWriter("build-droplet"))
	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Submitted build of " + dropletName))
}

func (runner *clusterTestRunner) launchDroplet(timeout time.Duration, appName, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Launching droplet %s as %s", dropletName, appName)))

	command := runner.command("launch-droplet", appName, dropletName)

	session, err := gexec.Start(command, getStyledWriter("launch-droplet"), getStyledWriter("launch-droplet"))
	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
}

func (runner *clusterTestRunner) listDroplets(timeout time.Duration, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline("Attempting to find droplet in the list"))

	command := runner.command("list-droplets")

	session, err := gexec.Start(command, getStyledWriter("list-droplets"), getStyledWriter("list-droplets"))
	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(dropletName))

	fmt.Fprintln(getStyledWriter("test"), "Found", dropletName, "in the list!")
}

func (runner *clusterTestRunner) removeDroplet(timeout time.Duration, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove droplet %s", dropletName)))

	command := runner.command("remove-droplet", dropletName)

	session, err := gexec.Start(command, getStyledWriter("remove-droplet"), getStyledWriter("remove-droplet"))
	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Droplet removed"))

	fmt.Fprintln(getStyledWriter("test"), "Removed", dropletName)
}

func (runner *clusterTestRunner) createDockerApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to create %s", appName)))
	createArgs := append([]string{"create", appName}, args...)
	command := runner.command(createArgs...)

	session, err := gexec.Start(command, getStyledWriter("create"), getStyledWriter("create"))

	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
	fmt.Fprintln(getStyledWriter("test"), "Yay! Created", appName)
}

func (runner *clusterTestRunner) streamLogs(timeout time.Duration, appName string, args ...string) *gexec.Session {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to stream logs from %s", appName)))
	command := runner.command("logs", appName)

	session, err := gexec.Start(command, getStyledWriter("logs"), getStyledWriter("logs"))

	Expect(err).NotTo(HaveOccurred())
	return session
}

func (runner *clusterTestRunner) streamDebugLogs(timeout time.Duration, args ...string) *gexec.Session {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline("Attempting to stream cluster debug logs"))
	command := runner.command("debug-logs")

	session, err := gexec.Start(command, getStyledWriter("debug"), getStyledWriter("debug"))

	Expect(err).NotTo(HaveOccurred())
	return session
}

func (runner *clusterTestRunner) scaleApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to scale %s", appName)))
	command := runner.command("scale", appName, "3")

	session, err := gexec.Start(command, getStyledWriter("scale"), getStyledWriter("scale"))

	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)
}

func (runner *clusterTestRunner) removeApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove app %s", appName)))
	command := runner.command("remove", appName)

	session, err := gexec.Start(command, getStyledWriter("remove"), getStyledWriter("remove"))

	Expect(err).NotTo(HaveOccurred())
	expectExit(timeout, session)
}

//TODO: add subcommand string param
func (runner *clusterTestRunner) command(arg ...string) *exec.Cmd {
	command := exec.Command(runner.ltcExecutablePath, arg...)
	appName := "APP_NAME=LATTICE-TEST-APP"
	cliHome := fmt.Sprintf("LATTICE_CLI_HOME=%s", runner.latticeCliHome)
	command.Env = []string{cliHome, appName}
	return command
}

func getStyledWriter(prefix string) io.Writer {
	return gexec.NewPrefixedWriter(fmt.Sprintf("[%s] ", colors.Yellow(prefix)), GinkgoWriter)
}

func errorCheckForConnection(ip string, port uint16) func() error {
	fmt.Fprintln(getStyledWriter("test"), "Connection to ", ip, ":", port)
	return func() error {
		response, err := makeTcpConnRequest(ip, port, "test")
		if err != nil {
			return err
		}
		fmt.Fprintln(getStyledWriter("test"), "Received response '", response, "'")

		if !strings.Contains(response, "docker-server1:test") {
			return errors.New("Did not get correct response from connection")
		}

		return nil
	}
}

func errorCheckForRoute(appRoute string) func() error {
	fmt.Fprintln(getStyledWriter("test"), "Polling for the appRoute", appRoute)
	return func() error {
		response, err := makeGetRequestToURL(appRoute)
		if err != nil {
			return err
		}

		io.Copy(ioutil.Discard, response.Body)
		defer response.Body.Close()

		if response.StatusCode != 200 {
			return fmt.Errorf("Status code %d should be 200", response.StatusCode)
		}

		return nil
	}
}

func errorCheckURLExists(url string) func() error {
	fmt.Fprintln(getStyledWriter("test"), "Polling for", url)
	return func() error {
		response, err := makeGetRequestToURL(url)
		if err != nil {
			return err
		}

		io.Copy(ioutil.Discard, response.Body)
		defer response.Body.Close()

		if response.StatusCode == 404 {
			return fmt.Errorf("URL %s doesn't exist", url)
		}

		return nil
	}
}

func countInstances(appRoute string, instanceCountChan chan<- int) {
	defer GinkgoRecover()
	instanceIndexRoute := fmt.Sprintf("%s/index", appRoute)
	instancesSeen := make(map[int]bool)

	instanceIndexChan := make(chan int, numCPU)

	for i := 0; i < numCPU; i++ {
		go pollForInstanceIndices(instanceIndexRoute, instanceIndexChan)
	}

	for {
		instanceIndex := <-instanceIndexChan
		instancesSeen[instanceIndex] = true
		instanceCountChan <- len(instancesSeen)
	}
}

func pollForInstanceIndices(appRoute string, instanceIndexChan chan<- int) {
	defer GinkgoRecover()
	for {
		response, err := makeGetRequestToURL(appRoute)
		Expect(err).To(BeNil())

		responseBody, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		Expect(err).To(BeNil())

		instanceIndex, err := strconv.Atoi(string(responseBody))
		if err != nil {
			continue
		}
		instanceIndexChan <- instanceIndex
	}
}

func makeTcpConnRequest(ip string, port uint16, req string) (string, error) {
	conn, err := net.Dial("tcp", ip+fmt.Sprintf(":%d", port))
	if err != nil {
		return "", err
	}

	fmt.Fprintf(conn, req+"\n")
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return line, nil
}

func makeGetRequestToURL(url string) (*http.Response, error) {
	routeWithScheme := fmt.Sprintf("http://%s", url)
	resp, err := http.DefaultClient.Get(routeWithScheme)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func expectExit(timeout time.Duration, session *gexec.Session) {
	expectExitInBuffer(timeout, session, session.Out)
}

func expectExitInBuffer(timeout time.Duration, session *gexec.Session, outputBuffer *gbytes.Buffer) {
	Eventually(session, timeout).Should(gexec.Exit(0))
	Expect(string(outputBuffer.Contents())).To(HaveSuffix("\n"))
}
