package integration_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/cloudfoundry-incubator/runtime-schema/models/factories"
	. "github.com/onsi/ginkgo"
	ginkgo_config "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/lattice-cli/config"
	"runtime"
	"time"
)

var (
	numCpu                                     int
	domain, username, password, latticeCliPath string
)

type IntegrationTestRunner interface {
	Run(timeout time.Duration, verbose bool)
}

type integrationTestRunner struct {
	t                GinkgoTestingT
	testOutputWriter io.Writer
	config           *config.Config
}

type ginkgoTestingT struct{}

func (g *ginkgoTestingT) Fail() {
	os.Exit(1)
}

func NewIntegrationTestRunner(outputWriter io.Writer, config *config.Config) IntegrationTestRunner {
	t := &ginkgoTestingT{}
	return &integrationTestRunner{testOutputWriter: outputWriter, config: config, t: t}
}

func (runner *integrationTestRunner) Run(timeout time.Duration, verbose bool) {
	ginkgo_config.DefaultReporterConfig.Verbose = verbose
	defineTheGinkgoTests(runner, timeout)
	RegisterFailHandler(Fail)
	RunSpecs(runner.t, "Lattice Integration Tests")
}

var (
	tmpDir string
)

func defineTheGinkgoTests(runner *integrationTestRunner, timeout time.Duration) {

	var _ = BeforeSuite(func() {
		tmpDir = os.TempDir()

		err := runner.config.Load()
		domain = runner.config.Target()
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "Error loading config")
			return
		}

		latticeCliPath = os.Args[0]
	})

	var _ = AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	var _ = Describe("Lattice", func() {
		Context("when desiring a docker-based LRP", func() {

			var (
				appName string
				route   string
			)

			BeforeEach(func() {
				appName = fmt.Sprintf("whetstone-%s", factories.GenerateGuid())
				route = fmt.Sprintf("%s.%s", appName, domain)
			})

			AfterEach(func() {
				removeApp(timeout, latticeCliPath, appName)

				Eventually(errorCheckForRoute(route), timeout, 1).Should(HaveOccurred())
			})

			It("eventually runs a docker app", func() {
				startDockerApp(timeout, latticeCliPath, appName, "cloudfoundry/lattice-app", "--working-dir=/", "--env", "APP_NAME", "--", "/lattice-app", "--message", "Hello Whetstone", "--quiet")

				Eventually(errorCheckForRoute(route), timeout, 1).ShouldNot(HaveOccurred())

				logsStream := streamLogs(timeout, latticeCliPath, appName)
				defer func() { logsStream.Terminate().Wait() }()

				Eventually(logsStream.Out, timeout).Should(gbytes.Say("WHETSTONE-TEST-APP. Says Hello Whetstone."))

				scaleApp(timeout, latticeCliPath, appName)

				instanceCountChan := make(chan int, numCpu)
				go countInstances(route, instanceCountChan)
				Eventually(instanceCountChan, timeout).Should(Receive(Equal(3)))

			})

			It("eventually runs a docker app with metadata from Docker Hub", func() {
				startDockerApp(timeout, latticeCliPath, appName, "cloudfoundry/lattice-app")

				Eventually(errorCheckForRoute(route), timeout, .5).ShouldNot(HaveOccurred())
			})
		})
	})
}

func startDockerApp(timeout time.Duration, latticeCliPath, appName string, args ...string) {

	fmt.Fprintf(GinkgoWriter, "Whetstone is attempting to start %s\n", appName)
	startArgs := append([]string{"start", appName}, args...)
	command := command(timeout, latticeCliPath, startArgs...)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
	fmt.Fprintf(GinkgoWriter, "Yay! Whetstone started %s\n", appName)
}

func streamLogs(timeout time.Duration, latticeCliPath, appName string) *gexec.Session {
	fmt.Fprintf(GinkgoWriter, "Whetstone is attempting to stream logs from %s\n", appName)
	command := command(timeout, latticeCliPath, "logs", appName)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session
}

func scaleApp(timeout time.Duration, latticeCliPath, appName string) {
	fmt.Fprintf(GinkgoWriter, "Whetstone is attempting to scale %s\n", appName)
	command := command(timeout, latticeCliPath, "scale", appName, "3")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)
}

func removeApp(timeout time.Duration, latticeCliPath, appName string) {
	fmt.Fprintf(GinkgoWriter, "Whetstone is attempting to remove %s\n", appName)
	command := command(timeout, latticeCliPath, "remove", appName)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)
}

func targetLattice(timeout time.Duration, latticeCliPath, domain, username, password string) {
	fmt.Fprintf(GinkgoWriter, "Whetstone is attempting to target %s with username:'%s' ; password:'%s'\n", domain, username, password)
	stdinReader, stdinWriter := io.Pipe()

	command := command(timeout, latticeCliPath, "target", domain)
	command.Stdin = stdinReader

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	if username != "" || password != "" {
		Eventually(session.Out).Should(gbytes.Say("Username: "))
		stdinWriter.Write([]byte(username + "\n"))

		Eventually(session.Out).Should(gbytes.Say("Password: "))
		stdinWriter.Write([]byte(password + "\n"))
	}

	stdinWriter.Close()
	expectExit(timeout, session)
}

func command(timeout time.Duration, name string, arg ...string) *exec.Cmd {
	command := exec.Command(name, arg...)
	appName := "APP_NAME=WHETSTONE-TEST-APP"
	cliHome := fmt.Sprintf("LATTICE_CLI_HOME=%s", tmpDir)
	cliTimeout := fmt.Sprintf("LATTICE_CLI_TIMEOUT=%d", timeout.Seconds())

	command.Env = []string{cliHome, appName, cliTimeout}
	return command
}

func errorCheckForRoute(route string) func() error {
	fmt.Fprintf(GinkgoWriter, "Whetstone is polling for the route %s\n", route)
	return func() error {
		fmt.Fprint(GinkgoWriter, ".")
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
	fmt.Printf("\nsession=%#v\nTIMEOUT=%d\n", session, timeout)
	Eventually(session, timeout).Should(gexec.Exit(0))
	Expect(string(session.Out.Contents())).To(HaveSuffix("\n"))

}

func init() {
	numCpu = runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
	tmpDir = os.TempDir()
}
