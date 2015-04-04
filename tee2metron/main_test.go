package main_test

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const maxUpdDatagramSize = 65507

var _ = Describe("tee2metron", func() {
	var tee2MetronPath, chattyProcessPath string

	BeforeSuite(func() {
		var err error
		tee2MetronPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/tee2metron")
		chattyProcessPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/tee2metron/test_helpers/chatty_process")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("prints stdout and stderr and streams them to metron", func() {
		metronReceivedBuffer, port := startFakeMetron()
		dropsondeDestinationFlag := "-dropsondeDestination=127.0.0.1:" + port
		command := exec.Command(tee2MetronPath, dropsondeDestinationFlag, "-sourceInstance=lattice-cell-123", chattyProcessPath, "chattyArg1", "chattyArg2", "-chattyFlag")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Consistently(session.Exited).ShouldNot(BeClosed())

		Eventually(session.Out).Should(gbytes.Say("Hi from stdout. My args are: [chattyArg1 chattyArg2 -chattyFlag]"))
		Eventually(session.Err).Should(gbytes.Say("Oopsie from stderr"))

		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say(chattyProcessPath))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say("lattice-debug"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say("lattice-cell-123"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer), 5).Should(gbytes.Say("Hi from stdout. My args are: [chattyArg1 chattyArg2 -chattyFlag]"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer), 5).Should(gbytes.Say("Oopsie from stderr"))
	})

	Context("With a bad command", func() {
		Context("when the command is missing", func() {
			It("prints and error message and exits", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", "-sourceInstance=lattice-cell-123")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session.Out).Should(gbytes.Say("Command not specified!"))
				Eventually(session.Out).Should(gbytes.Say("Usage: tee2metron -dropsondeDestionation=127.0.0.1:3457 -sourceInstance=lattice-cell-21 COMMAND"))
				Eventually(session.Exited).Should(BeClosed())
				Expect(session.ExitCode()).To(Equal(3))
			})
		})
		Context("when there is an error executing the command", func() {
			It("prints and error message and exits", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", "-sourceInstance=lattice-cell-123", "do-the-fandango-for-me")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session.Out).Should(gbytes.Say(`exec: "do-the-fandango-for-me": executable file not found in \$PATH`))
				Eventually(session.Exited).Should(BeClosed())
				Expect(session.ExitCode()).To(Equal(3))
			})
		})
	})

	Describe("Flags", func() {
		Describe("-dropsondeDestination", func() {
			It("is required", func() {
				command := exec.Command(tee2MetronPath, "-sourceInstance=lattice-cell-123", chattyProcessPath)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

				Eventually(session.Buffer()).Should(gbytes.Say("dropsondeDestination flag is required"))
				Expect(err).ToNot(HaveOccurred())
				Eventually(session.Exited).Should(BeClosed())
				Expect(session.ExitCode()).To(Equal(1))
			})
		})

		Describe("-sourceInstance", func() {
			It("is required", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", chattyProcessPath)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

				Eventually(session.Buffer()).Should(gbytes.Say("sourceInstance flag is required"))
				Expect(err).ToNot(HaveOccurred())
				Eventually(session.Exited).Should(BeClosed())
				Expect(session.ExitCode()).To(Equal(1))
			})
		})
	})
})

func startFakeMetron() (metronReceivedBufferPtr *[]byte, port string) {
	connection, err := net.ListenPacket("udp", "") //listen on some free port

	if err != nil {
		Fail("Error starting the integration test: Could not listen for udp packets on os-assigned port: " + err.Error())
	}

	metronReceivedBuffer := make([]byte, maxUpdDatagramSize)
	go func() {
		defer connection.Close()
		for {
			_, _, err = connection.ReadFrom(metronReceivedBuffer)
			if err != nil {
				panic(err)
			}
			if string(metronReceivedBuffer) != string(make([]byte, maxUpdDatagramSize)) { //output if not empty
				fmt.Fprint(GinkgoWriter, "\nRead UDP Packet: ", string(metronReceivedBuffer), "\n")
			}
		}
	}()

	portMatch := regexp.MustCompile(`\[::\]:(\d+)`)
	port = portMatch.FindStringSubmatch(connection.LocalAddr().String())[1]

	return &metronReceivedBuffer, port
}
