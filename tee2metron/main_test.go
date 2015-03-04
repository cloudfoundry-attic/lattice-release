package main_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"net"
	"os/exec"

	"regexp"
	//    "errors""strings"
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

	It("prints stdout to stdout and stderr to stderr, while logging a stdout log message to ", func() {
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
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say("Hi from stdout. My args are: [chattyArg1 chattyArg2 -chattyFlag]"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say("Oopsie from stderr"))
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
