package integration_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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

var (
	numCpu int
)

func init() {
	numCpu = runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
}

type IntegrationTestRunner interface {
	Run(timeout time.Duration, verbose, cliHelp bool)
}

type integrationTestRunner struct {
	testingT          GinkgoTestingT
	config            *config.Config
	latticeCliHome    string
	ltcExecutablePath string
}

type ginkgoTestingT struct{}

func (g *ginkgoTestingT) Fail() {
	os.Exit(1)
}

func NewIntegrationTestRunner(config *config.Config, latticeCliHome string) IntegrationTestRunner {
	return &integrationTestRunner{
		config:            config,
		testingT:          &ginkgoTestingT{},
		latticeCliHome:    latticeCliHome,
		ltcExecutablePath: os.Args[0],
	}
}

func (runner *integrationTestRunner) Run(timeout time.Duration, verbose, cliHelp bool) {
	ginkgo_config.DefaultReporterConfig.Verbose = verbose
	ginkgo_config.DefaultReporterConfig.SlowSpecThreshold = float64(45)
	if cliHelp {
		defineTheMainTests(runner)
	} else {
		defineTheGinkgoTests(runner, timeout)
	}
	RegisterFailHandler(Fail)
	RunSpecs(runner.testingT, "Lattice Integration Tests")
}

func defineTheGinkgoTests(runner *integrationTestRunner, timeout time.Duration) {
	var (
		dropletName string
	)

	var _ = BeforeSuite(func() {
		err := runner.config.Load()
		if err != nil {
			fmt.Fprintln(getStyledWriter("test"), "Error loading config")
			return
		}

		// Generate a droplet name up front so that it can persist across droplet tests
		dropletGuid, err := uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		dropletName = "droplet-" + dropletGuid.String()
	})

	var _ = AfterSuite(func() {
		gexec.CleanupBuildArtifacts()

		runner.removeDroplet(timeout, dropletName)

		blobTarget := runner.config.BlobTarget()
		dropletURL := fmt.Sprintf("%s:%d/%s/%s/",
			blobTarget.TargetHost,
			blobTarget.TargetPort,
			blobTarget.BucketName,
			dropletName)
		Eventually(errorCheckURLExists(dropletURL), timeout, 1).Should(HaveOccurred())
	})

	var _ = Describe("Lattice", func() {
		Context("docker", func() {
			Context("when desiring a docker-based LRP", func() {

				var (
					appName string
					route   string
				)

				BeforeEach(func() {
					appGuid, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())

					appName = fmt.Sprintf("lattice-test-app-%s", appGuid.String())
					route = fmt.Sprintf("%s.%s", appName, runner.config.Target())
				})

				AfterEach(func() {
					runner.removeApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))

					Eventually(errorCheckForRoute(route), timeout, 1).Should(HaveOccurred())
				})

				It("eventually runs a docker app", func() {
					debugLogsStream := runner.streamDebugLogs(timeout)
					defer func() { debugLogsStream.Terminate().Wait() }()

					runner.createDockerApp(timeout, appName, "cloudfoundry/lattice-app", fmt.Sprintf("--timeout=%s", timeout.String()), "--working-dir=/", "--env", "APP_NAME", "--", "/lattice-app", "--message", "Hello Lattice User", "--quiet")

					Eventually(errorCheckForRoute(route), timeout, 1).ShouldNot(HaveOccurred())

					Eventually(debugLogsStream.Out, timeout).Should(gbytes.Say("rep.*cell-\\d+"))
					Eventually(debugLogsStream.Out, timeout).Should(gbytes.Say("garden-linux.*cell-\\d+"))
					debugLogsStream.Terminate().Wait()

					logsStream := runner.streamLogs(timeout, appName)
					defer func() { logsStream.Terminate().Wait() }()

					Eventually(logsStream.Out, timeout).Should(gbytes.Say("LATTICE-TEST-APP. Says Hello Lattice User."))

					runner.scaleApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))

					instanceCountChan := make(chan int, numCpu)
					go countInstances(route, instanceCountChan)
					Eventually(instanceCountChan, timeout).Should(Receive(Equal(3)))
				})

				It("eventually runs a docker app with metadata from Docker Hub", func() {
					runner.createDockerApp(timeout, appName, "cloudfoundry/lattice-app")

					Eventually(errorCheckForRoute(route), timeout, .5).ShouldNot(HaveOccurred())
				})
			})
		})

		Context("droplets", func() {
			It("uploads a file named bits.tgz to the blob store", func() {
				tmpFile, err := ioutil.TempFile("", "bits.txt")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString("01001100010000010101010001010100010010010100001101000101")
				Expect(err).ToNot(HaveOccurred())

				tmpFile.Close()

				runner.uploadBits(timeout, dropletName, tmpFile.Name())

				blobTarget := runner.config.BlobTarget()
				bitsURL := fmt.Sprintf("%s:%d/%s/%s/bits.tgz",
					blobTarget.TargetHost,
					blobTarget.TargetPort,
					blobTarget.BucketName,
					dropletName)
				Eventually(errorCheckURLExists(bitsURL), timeout, 1).ShouldNot(HaveOccurred())
			})

			It("builds a droplet", func() {
				By("checking out lattice-app from github")
				gitDir := runner.cloneRepo(timeout, "https://github.com/pivotal-cf-experimental/lattice-app.git")
				defer os.RemoveAll(gitDir)

				By("launching a build task")
				runner.buildDroplet(timeout, dropletName, "https://github.com/cloudfoundry/go-buildpack.git", gitDir)

				By("uploading to the blob store")
				blobTarget := runner.config.BlobTarget()
				bitsURL := fmt.Sprintf("%s:%d/%s/%s",
					blobTarget.TargetHost,
					blobTarget.TargetPort,
					blobTarget.BucketName,
					dropletName)
				Eventually(errorCheckURLExists(bitsURL+"/bits.tgz"), timeout, 1).ShouldNot(HaveOccurred())

				By("uploading a compiled droplet to the blob store")
				Eventually(errorCheckURLExists(bitsURL+"/droplet.tgz"), timeout, 1).ShouldNot(HaveOccurred())
			})

			It("lists droplets", func() {
				runner.listDroplets(timeout, dropletName)
			})

			It("launches a droplet", func() {
				appName := "running-" + dropletName

				runner.launchDroplet(timeout, appName, dropletName)

				route := fmt.Sprintf("%s.%s", appName, runner.config.Target())
				Eventually(errorCheckForRoute(route), timeout, .5).ShouldNot(HaveOccurred())

				runner.removeApp(timeout, appName, fmt.Sprintf("--timeout=%s", timeout.String()))

				Eventually(errorCheckForRoute(route), timeout, 1).Should(HaveOccurred())
			})
		})

	})
}

func (runner *integrationTestRunner) cloneRepo(timeout time.Duration, repoURL string) string {
	tmpDir, err := ioutil.TempDir("", "repo")
	Expect(err).ToNot(HaveOccurred())

	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to clone %s to %s", repoURL, tmpDir)))

	command := exec.Command("/usr/bin/env", "git", "clone", repoURL, tmpDir)

	session, err := gexec.Start(command, getStyledWriter("git-clone"), getStyledWriter("git-clone"))
	Expect(err).ToNot(HaveOccurred())
	expectExitInBuffer(timeout, session, session.Err)

	Eventually(session.Err).Should(gbytes.Say(fmt.Sprintf("Cloning into '%s'...", tmpDir)))

	fmt.Fprintf(getStyledWriter("test"), "Cloned %s into %s\n", repoURL, tmpDir)

	return tmpDir
}

func (runner *integrationTestRunner) uploadBits(timeout time.Duration, dropletName, bits string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to upload %s to %s", bits, dropletName)))

	command := runner.command("upload-bits", dropletName, bits)

	session, err := gexec.Start(command, getStyledWriter("upload-bits"), getStyledWriter("upload-bits"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Successfully uploaded " + dropletName))

	fmt.Fprintln(getStyledWriter("test"), "Uploaded", bits, "to", dropletName)
}

func (runner *integrationTestRunner) buildDroplet(timeout time.Duration, dropletName, buildpack, srcDir string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Submitting build of %s with buildpack %s", dropletName, buildpack)))

	command := runner.command("build-droplet", dropletName, buildpack)
	command.Dir = srcDir

	session, err := gexec.Start(command, getStyledWriter("build-droplet"), getStyledWriter("build-droplet"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Submitted build of " + dropletName))
}

func (runner *integrationTestRunner) launchDroplet(timeout time.Duration, appName, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Launching droplet %s as %s", dropletName, appName)))

	command := runner.command("launch-droplet", appName, dropletName)

	session, err := gexec.Start(command, getStyledWriter("launch-droplet"), getStyledWriter("launch-droplet"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
}

func (runner *integrationTestRunner) listDroplets(timeout time.Duration, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline("Attempting to find droplet in the list"))

	command := runner.command("list-droplets")

	session, err := gexec.Start(command, getStyledWriter("list-droplets"), getStyledWriter("list-droplets"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(dropletName))

	fmt.Fprintln(getStyledWriter("test"), "Found", dropletName, "in the list!")
}

func (runner *integrationTestRunner) removeDroplet(timeout time.Duration, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove droplet %s", dropletName)))

	command := runner.command("remove-droplet", dropletName)

	session, err := gexec.Start(command, getStyledWriter("remove-droplet"), getStyledWriter("remove-droplet"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Droplet removed"))

	fmt.Fprintln(getStyledWriter("test"), "Removed", dropletName)
}

func (runner *integrationTestRunner) createDockerApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to create %s", appName)))
	createArgs := append([]string{"create", appName}, args...)
	command := runner.command(createArgs...)

	session, err := gexec.Start(command, getStyledWriter("create"), getStyledWriter("create"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
	fmt.Fprintln(getStyledWriter("test"), "Yay! Created", appName)
}

func (runner *integrationTestRunner) streamLogs(timeout time.Duration, appName string, args ...string) *gexec.Session {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to stream logs from %s", appName)))
	command := runner.command("logs", appName)

	session, err := gexec.Start(command, getStyledWriter("logs"), getStyledWriter("logs"))

	Expect(err).ToNot(HaveOccurred())
	return session
}

func (runner *integrationTestRunner) streamDebugLogs(timeout time.Duration, args ...string) *gexec.Session {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline("Attempting to stream cluster debug logs"))
	command := runner.command("debug-logs")

	session, err := gexec.Start(command, getStyledWriter("debug"), getStyledWriter("debug"))

	Expect(err).ToNot(HaveOccurred())
	return session
}

func (runner *integrationTestRunner) scaleApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to scale %s", appName)))
	command := runner.command("scale", appName, "3")

	session, err := gexec.Start(command, getStyledWriter("scale"), getStyledWriter("scale"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)
}

func (runner *integrationTestRunner) removeApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove app %s", appName)))
	command := runner.command("remove", appName)

	session, err := gexec.Start(command, getStyledWriter("remove"), getStyledWriter("remove"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)
}

//TODO: add subcommand string param
func (runner *integrationTestRunner) command(arg ...string) *exec.Cmd {
	command := exec.Command(runner.ltcExecutablePath, arg...)
	appName := "APP_NAME=LATTICE-TEST-APP"
	cliHome := fmt.Sprintf("LATTICE_CLI_HOME=%s", runner.latticeCliHome)
	command.Env = []string{cliHome, appName}
	return command
}

func getStyledWriter(prefix string) io.Writer {
	return gexec.NewPrefixedWriter(fmt.Sprintf("[%s] ", colors.Yellow(prefix)), GinkgoWriter)
}

func errorCheckForRoute(route string) func() error {
	fmt.Fprintln(getStyledWriter("test"), "Polling for the route", route)
	return func() error {
		response, err := makeGetRequestToRoute(route)
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
		response, err := makeGetRequestToRoute(url)
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

func countInstances(route string, instanceCountChan chan<- int) {
	defer GinkgoRecover()
	instanceIndexRoute := fmt.Sprintf("%s/index", route)
	instancesSeen := make(map[int]bool)

	instanceIndexChan := make(chan int, numCpu)

	for i := 0; i < numCpu; i++ {
		go pollForInstanceIndices(instanceIndexRoute, instanceIndexChan)
	}

	for {
		instanceIndex := <-instanceIndexChan
		instancesSeen[instanceIndex] = true
		instanceCountChan <- len(instancesSeen)
	}
}

func pollForInstanceIndices(route string, instanceIndexChan chan<- int) {
	defer GinkgoRecover()
	for {
		response, err := makeGetRequestToRoute(route)
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

func makeGetRequestToRoute(route string) (*http.Response, error) {
	routeWithScheme := fmt.Sprintf("http://%s", route)
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

func defineTheMainTests(runner *integrationTestRunner) {
	Describe("exit codes", func() {
		It("exits non-zero when an unknown command is invoked", func() {
			command := runner.command("unknownCommand")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 3*time.Second).Should(gbytes.Say("not a registered command"))
			Eventually(session).Should(gexec.Exit(1))
		})

		It("exits non-zero when known command is invoked with invalid option", func() {
			command := runner.command("status", "--badFlag")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 3*time.Second).Should(gexec.Exit(1))
		})
	})
}
