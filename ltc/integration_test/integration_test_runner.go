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
	ginkgo_config.DefaultReporterConfig.SlowSpecThreshold = float64(20)
	if cliHelp {
		defineTheMainTests(runner)
	} else {
		defineTheGinkgoTests(runner, timeout)
	}
	RegisterFailHandler(Fail)
	RunSpecs(runner.testingT, "Lattice Integration Tests")
}

func defineTheGinkgoTests(runner *integrationTestRunner, timeout time.Duration) {

	var _ = BeforeSuite(func() {
		err := runner.config.Load()
		if err != nil {
			fmt.Fprintf(getStyledWriter("test"), "Error loading config")
			return
		}
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
}

func (runner *integrationTestRunner) createDockerApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintf(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to create %s", appName))+"\n")
	createArgs := append([]string{"create", appName}, args...)
	command := runner.command(createArgs...)

	session, err := gexec.Start(command, getStyledWriter("create"), getStyledWriter("create"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
	fmt.Fprintf(getStyledWriter("test"), "Yay! Created %s\n", appName)
}

func (runner *integrationTestRunner) streamLogs(timeout time.Duration, appName string, args ...string) *gexec.Session {
	fmt.Fprintf(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to stream logs from %s", appName))+"\n")
	command := runner.command("logs", appName)

	session, err := gexec.Start(command, getStyledWriter("logs"), getStyledWriter("logs"))

	Expect(err).ToNot(HaveOccurred())
	return session
}

func (runner *integrationTestRunner) streamDebugLogs(timeout time.Duration, args ...string) *gexec.Session {
	fmt.Fprintf(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to stream cluster debug logs"))+"\n")
	command := runner.command("debug-logs")

	session, err := gexec.Start(command, getStyledWriter("debug"), getStyledWriter("debug"))

	Expect(err).ToNot(HaveOccurred())
	return session
}

func (runner *integrationTestRunner) scaleApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintf(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to scale %s", appName))+"\n")
	command := runner.command("scale", appName, "3")

	session, err := gexec.Start(command, getStyledWriter("scale"), getStyledWriter("scale"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)
}

func (runner *integrationTestRunner) removeApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintf(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove %s", appName))+"\n")
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
	fmt.Fprintf(getStyledWriter("test"), "Polling for the route %s\n", route)
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
	Eventually(session, timeout).Should(gexec.Exit(0))
	Expect(string(session.Out.Contents())).To(HaveSuffix("\n"))
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
