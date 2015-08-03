package main_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/config_helpers"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/cloudfoundry-incubator/receptor"
)

var _ = Describe("LatticeCli Main", func() {
	const latticeCliHomeVar = "LATTICE_CLI_HOME"

	var (
		outputBuffer               *gbytes.Buffer
		fakeServer                 *ghttp.Server
		listenerAddr, configRoot   string
		listenerHost, listenerPort string
	)

	desiredResponse := []receptor.DesiredLRPResponse{
		{
			ProcessGuid: "some-guid",
			Instances:   1,
		},
	}
	actualResponse := []receptor.ActualLRPResponse{
		{
			ProcessGuid: "some-guid",
			Index:       12,
			State:       receptor.ActualLRPStateRunning,
		},
	}
	tasksResponse := []receptor.TaskResponse{
		{TaskGuid: "task-guid"},
	}

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
		fakeServer = ghttp.NewServer()
		listenerAddr = fakeServer.HTTPTestServer.Listener.Addr().String()

		fakeServer.RouteToHandler("GET", "/v1/desired_lrps",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/desired_lrps"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, desiredResponse),
			),
		)
		fakeServer.RouteToHandler("GET", "/v1/actual_lrps",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/actual_lrps"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, actualResponse),
			),
		)
		fakeServer.RouteToHandler("GET", "/v1/tasks",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/tasks"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, tasksResponse),
			),
		)

		var err error
		configRoot, err = ioutil.TempDir("", "config")
		Expect(err).NotTo(HaveOccurred())

		ltcConfig := config.New(persister.NewFilePersister(config_helpers.ConfigFileLocation(configRoot)))

		listenerHost, listenerPort, err = net.SplitHostPort(listenerAddr)
		Expect(err).NotTo(HaveOccurred())
		ltcConfig.SetTarget(fmt.Sprintf("%s.xip.io:%s", listenerHost, listenerPort))
		Expect(ltcConfig.Save()).To(Succeed())

		os.Setenv(latticeCliHomeVar, configRoot)
	})
	AfterEach(func() {
		fakeServer.Close()
		Expect(os.RemoveAll(configRoot)).To(Succeed())
		Expect(os.Unsetenv(latticeCliHomeVar)).To(Succeed())
	})

	ltcCommand := func(args ...string) *exec.Cmd {
		command := exec.Command(ltcPath, args...)
		cliHome := fmt.Sprintf("LATTICE_CLI_HOME=%s", configRoot)
		command.Env = []string{cliHome}

		return command
	}

	Context("when ltc is invoked with no args", func() {
		It("should exit zero and display ltc help text", func() {
			command := ltcCommand()

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("ltc - Command line interface for Lattice."))
		})
	})

	Describe("Exit Codes", func() {
		It("should exit zero and run desired command", func() {
			command := ltcCommand("target")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Target:"))
			Expect(session.Out).To(gbytes.Say(fmt.Sprintf("%s.xip.io:%s", listenerHost, listenerPort)))
		})

		Context("when an unknown command is invoked", func() {
			It("should exit non-zero when an unknown command is invoked", func() {
				command := ltcCommand("unknown-command")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out).To(gbytes.Say("not a registered command"))
			})
		})

		Context("when a known command is invoked with an invalid flag", func() {
			It("should exit non-zero and print incorrect usage", func() {
				command := ltcCommand("status", "--badFlag")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out).To(test_helpers.SayIncorrectUsage())
			})
		})
	})
})
