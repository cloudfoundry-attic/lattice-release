package cluster_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
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
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/nu7hatch/gouuid"
)

var numCpu int

func init() {
	numCpu = runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)
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

func NewClusterTestRunner(config *config.Config, latticeCliHome string) ClusterTestRunner {
	return &clusterTestRunner{
		config:            config,
		testingT:          &ginkgoTestingT{},
		latticeCliHome:    latticeCliHome,
		ltcExecutablePath: os.Args[0],
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
		err := runner.config.Load()
		if err != nil {
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
					appGuid, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())

					appName = fmt.Sprintf("lattice-test-app-%s", appGuid.String())
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

					instanceCountChan := make(chan int, numCpu)
					go countInstances(appRoute, instanceCountChan)
					Eventually(instanceCountChan, timeout).Should(Receive(Equal(3)))
				})

				It("should run a docker app using metadata from Docker Hub", func() {
					runner.createDockerApp(timeout, appName, "cloudfoundry/lattice-app")

					Eventually(errorCheckForRoute(appRoute), timeout, .5).ShouldNot(HaveOccurred())
				})

				Context("when the docker app requires escalated privileges to run", func() {
					BeforeEach(func() {
						appGuid, err := uuid.NewV4()
						Expect(err).ToNot(HaveOccurred())

						appName = fmt.Sprintf("nginx-app-%s", appGuid.String())
						appRoute = fmt.Sprintf("%s.%s", appName, runner.config.Target())
					})

					It("should start the nginx app successfully", func() {
						By("passing the `--run-as-root` flag to `ltc create`")
						runner.createDockerApp(timeout, appName, "nginx", "--run-as-root")

						Eventually(errorCheckForRoute(appRoute), timeout, .5).ShouldNot(HaveOccurred())
					})
				})
			})
		})

		Context("tcp routing", func() {
			var (
				appName      string
				externalPort uint16
				desiredLrp   receptor.DesiredLRPCreateRequest
			)

			BeforeEach(func() {
				appGuid, err := uuid.NewV4()
				Expect(err).ToNot(HaveOccurred())

				appName = fmt.Sprintf("lattice-test-app-%s", appGuid.String())
				externalPort = 64000
				containerPort := uint16(5222)
				routingInfo := json.RawMessage([]byte(fmt.Sprintf("{\"external_port\":%d, \"container_port\":%d}", externalPort, containerPort)))

				desiredLrp = receptor.DesiredLRPCreateRequest{
					ProcessGuid: appGuid.String(),
					LogGuid:     "log-guid",
					Domain:      "ge",
					Instances:   1,
					Setup: &models.SerialAction{
						Actions: []models.Action{
							&models.RunAction{
								Path: "sh",
								User: "vcap",
								Args: []string{
									"-c",
									"curl https://s3.amazonaws.com/router-release-blobs/tcp-sample-receiver.linux -o /tmp/tcp-sample-receiver && chmod +x /tmp/tcp-sample-receiver",
								},
							},
						},
					},
					Action: &models.ParallelAction{
						Actions: []models.Action{
							&models.RunAction{
								Path: "sh",
								User: "vcap",
								Args: []string{
									"-c",
									fmt.Sprintf("/tmp/tcp-sample-receiver -address 0.0.0.0:%d -serverId %s", containerPort, 1),
								},
							},
						},
					},
					Monitor: &models.RunAction{
						Path: "sh",
						User: "vcap",
						Args: []string{
							"-c",
							fmt.Sprintf("nc -z 0.0.0.0 %d", containerPort),
						}},
					StartTimeout: 60,
					RootFS:       "docker:///cloudfoundry/trusty64",
					MemoryMB:     128,
					DiskMB:       128,
					Ports:        []uint16{containerPort},
					Routes: receptor.RoutingInfo{
						"tcp-router": &routingInfo,
					},
					EgressRules: []models.SecurityGroupRule{
						{
							Protocol:     models.TCPProtocol,
							Destinations: []string{"0.0.0.0-255.255.255.255"},
							Ports:        []uint16{80, 443},
						},
						{
							Protocol:     models.UDPProtocol,
							Destinations: []string{"0.0.0.0/0"},
							PortRange: &models.PortRange{
								Start: 53,
								End:   53,
							},
						},
					},
				}
			})

			It("routes tcp traffic to a container", func() {
				helperErr := test_helpers.TempJsonFile(desiredLrp, func(file string) {
					By("Submitting an LRP")
					runner.submitLrp(timeout, file)
				})
				Expect(helperErr).NotTo(HaveOccurred())

				By("connecting to the running LRP over TCP")
				Eventually(errorCheckForConnection(runner.config.Target(), externalPort), timeout, 1).ShouldNot(HaveOccurred())
			})
		})

		if runner.config.BlobTarget().AccessKey != "" && runner.config.BlobTarget().SecretKey != "" {
			Context("droplets", func() {
				var dropletName, appName, dropletFolderURL, appRoute string

				BeforeEach(func() {
					// Generate a droplet name up front so that it can persist across droplet tests
					dropletGuid, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					dropletName = "droplet-" + dropletGuid.String()

					appName = "running-" + dropletName

					blobTarget := runner.config.BlobTarget()
					dropletFolderURL = fmt.Sprintf("%s:%d/%s/%s",
						blobTarget.TargetHost,
						blobTarget.TargetPort,
						blobTarget.BucketName,
						dropletName)

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
		}
	})
}

func (runner *clusterTestRunner) cloneRepo(timeout time.Duration, repoURL string) string {
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

func (runner *clusterTestRunner) submitLrp(timeout time.Duration, jsonPath string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to submit lrp at %s", jsonPath)))

	command := runner.command("submit-lrp", jsonPath)

	session, err := gexec.Start(command, getStyledWriter("submit-lrp"), getStyledWriter("submit-lrp"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Successfully submitted"))

	fmt.Fprintln(getStyledWriter("test"), "Submitted lrp")
}

func (runner *clusterTestRunner) uploadBits(timeout time.Duration, dropletName, bits string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to upload %s to %s", bits, dropletName)))

	command := runner.command("upload-bits", dropletName, bits)

	session, err := gexec.Start(command, getStyledWriter("upload-bits"), getStyledWriter("upload-bits"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Successfully uploaded " + dropletName))

	fmt.Fprintln(getStyledWriter("test"), "Uploaded", bits, "to", dropletName)
}

func (runner *clusterTestRunner) buildDroplet(timeout time.Duration, dropletName, buildpack, srcDir string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Submitting build of %s with buildpack %s", dropletName, buildpack)))

	command := runner.command("build-droplet", dropletName, buildpack, "--timeout", timeout.String())
	command.Dir = srcDir

	session, err := gexec.Start(command, getStyledWriter("build-droplet"), getStyledWriter("build-droplet"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Submitted build of " + dropletName))
}

func (runner *clusterTestRunner) launchDroplet(timeout time.Duration, appName, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Launching droplet %s as %s", dropletName, appName)))

	command := runner.command("launch-droplet", appName, dropletName)

	session, err := gexec.Start(command, getStyledWriter("launch-droplet"), getStyledWriter("launch-droplet"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
}

func (runner *clusterTestRunner) listDroplets(timeout time.Duration, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline("Attempting to find droplet in the list"))

	command := runner.command("list-droplets")

	session, err := gexec.Start(command, getStyledWriter("list-droplets"), getStyledWriter("list-droplets"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(dropletName))

	fmt.Fprintln(getStyledWriter("test"), "Found", dropletName, "in the list!")
}

func (runner *clusterTestRunner) removeDroplet(timeout time.Duration, dropletName string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove droplet %s", dropletName)))

	command := runner.command("remove-droplet", dropletName)

	session, err := gexec.Start(command, getStyledWriter("remove-droplet"), getStyledWriter("remove-droplet"))
	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say("Droplet removed"))

	fmt.Fprintln(getStyledWriter("test"), "Removed", dropletName)
}

func (runner *clusterTestRunner) createDockerApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to create %s", appName)))
	createArgs := append([]string{"create", appName}, args...)
	command := runner.command(createArgs...)

	session, err := gexec.Start(command, getStyledWriter("create"), getStyledWriter("create"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)

	Expect(session.Out).To(gbytes.Say(appName + " is now running."))
	fmt.Fprintln(getStyledWriter("test"), "Yay! Created", appName)
}

func (runner *clusterTestRunner) streamLogs(timeout time.Duration, appName string, args ...string) *gexec.Session {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to stream logs from %s", appName)))
	command := runner.command("logs", appName)

	session, err := gexec.Start(command, getStyledWriter("logs"), getStyledWriter("logs"))

	Expect(err).ToNot(HaveOccurred())
	return session
}

func (runner *clusterTestRunner) streamDebugLogs(timeout time.Duration, args ...string) *gexec.Session {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline("Attempting to stream cluster debug logs"))
	command := runner.command("debug-logs")

	session, err := gexec.Start(command, getStyledWriter("debug"), getStyledWriter("debug"))

	Expect(err).ToNot(HaveOccurred())
	return session
}

func (runner *clusterTestRunner) scaleApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to scale %s", appName)))
	command := runner.command("scale", appName, "3")

	session, err := gexec.Start(command, getStyledWriter("scale"), getStyledWriter("scale"))

	Expect(err).ToNot(HaveOccurred())
	expectExit(timeout, session)
}

func (runner *clusterTestRunner) removeApp(timeout time.Duration, appName string, args ...string) {
	fmt.Fprintln(getStyledWriter("test"), colors.PurpleUnderline(fmt.Sprintf("Attempting to remove app %s", appName)))
	command := runner.command("remove", appName)

	session, err := gexec.Start(command, getStyledWriter("remove"), getStyledWriter("remove"))

	Expect(err).ToNot(HaveOccurred())
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

		if response != "server-1:test" {
			errors.New("Did not get correct response from connection")
		}

		return nil
	}
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

func countInstances(appRoute string, instanceCountChan chan<- int) {
	defer GinkgoRecover()
	instanceIndexRoute := fmt.Sprintf("%s/index", appRoute)
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

func pollForInstanceIndices(appRoute string, instanceIndexChan chan<- int) {
	defer GinkgoRecover()
	for {
		response, err := makeGetRequestToRoute(appRoute)
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
