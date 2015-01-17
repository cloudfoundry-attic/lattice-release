package whetstone_test

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
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	cli    string
	tmpDir string
)

var _ = BeforeSuite(func() {
	tmpDir = os.TempDir()

	var err error
	cli, err = gexec.Build("github.com/pivotal-cf-experimental/lattice-cli/ltc")
	Expect(err).ToNot(HaveOccurred())
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

			targetLattice(domain, username, password)
		})

		AfterEach(func() {
			removeApp(appName)

			Eventually(errorCheckForRoute(route), timeout, 1).Should(HaveOccurred())
		})

		It("eventually runs a docker app", func() {
			startDockerApp(appName, "docker:///cloudfoundry/lattice-app", "--env", "APP_NAME", "--", "/lattice-app", "--message", "Hello Whetstone", "--quiet")

			Eventually(errorCheckForRoute(route), timeout, 1).ShouldNot(HaveOccurred())

			logsStream := streamLogs(appName)
			Eventually(logsStream.Out, timeout).Should(gbytes.Say("WHETSTONE TEST APP. Says Hello Whetstone."))

			scaleApp(appName)

			instanceCountChan := make(chan int, numCpu)
			go countInstances(route, instanceCountChan)
			Eventually(instanceCountChan, timeout).Should(Receive(Equal(3)))

			logsStream.Terminate().Wait()
		})

		It("eventually runs a docker app with metadata from Docker Hub", func() {
			startDockerApp(appName, "docker:///cloudfoundry/lattice-app")

			Eventually(errorCheckForRoute(route), timeout, 1).ShouldNot(HaveOccurred())
		})
	})
})

func startDockerApp(appName string, args ...string) {
	startArgs := append([]string{"start", appName}, args...)
	command := command(cli, startArgs...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	expectExit(session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
}

func streamLogs(appName string) *gexec.Session {
	command := command(cli, "logs", appName)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session
}

func scaleApp(appName string) {
	command := command(cli, "scale", appName, "3")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	expectExit(session)
}

func removeApp(appName string) {
	command := command(cli, "remove", appName)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	expectExit(session)
}

func targetLattice(domain, username, password string) {
	stdinReader, stdinWriter := io.Pipe()

	command := command(cli, "target", domain)
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
	expectExit(session)
}

func command(name string, arg ...string) *exec.Cmd {
	command := exec.Command(name, arg...)

	appName := "APP_NAME=WHETSTONE TEST APP"
	cliHome := fmt.Sprintf("LATTICE_CLI_HOME=%s", tmpDir)
	cliTimeout := fmt.Sprintf("LATTICE_CLI_TIMEOUT=%d", timeout)

	command.Env = []string{cliHome, appName, cliTimeout}
	return command
}

func errorCheckForRoute(route string) func() error {
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

func expectExit(session *gexec.Session) {
	Eventually(session, timeout).Should(gexec.Exit(0))
	Expect(string(session.Out.Contents())).To(HaveSuffix("\n"))
}
