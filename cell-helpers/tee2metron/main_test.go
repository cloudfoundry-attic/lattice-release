package main_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const maxUpdDatagramSize = 65507

var _ = Describe("tee2metron", func() {
	It("prints stdout and stderr and streams them to metron", func() {
		metronReceivedBuffer, port := startFakeMetron()
		dropsondeDestinationFlag := "-dropsondeDestination=127.0.0.1:" + port
		command := exec.Command(tee2MetronPath, dropsondeDestinationFlag, "-sourceInstance=cell-123", chattyProcessPath, "chattyArg1", "chattyArg2", "-chattyFlag")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Consistently(session.Exited).ShouldNot(BeClosed())
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say(chattyProcessPath))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say("lattice-debug"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer)).Should(gbytes.Say("cell-123"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer), 5).Should(gbytes.Say("Hi from stdout. My args are: [chattyArg1 chattyArg2 -chattyFlag]"))
		Eventually(gbytes.BufferWithBytes(*metronReceivedBuffer), 5).Should(gbytes.Say("Oopsie from stderr"))

		Eventually(session.Terminate()).Should(gexec.Exit(2))
		Expect(session.Out).To(gbytes.Say("Hi from stdout. My args are: [chattyArg1 chattyArg2 -chattyFlag]"))
		Expect(session.Err).To(gbytes.Say("Oopsie from stderr"))
	})

	Context("with a bad command", func() {
		Context("when the command is missing", func() {
			It("prints and error message and exits", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", "-sourceInstance=cell-123")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("Command not specified!"))
				Expect(session.Out).To(gbytes.Say("Usage: tee2metron -dropsondeDestionation=127.0.0.1:3457 -sourceInstance=cell-21 COMMAND"))
			})
		})
		Context("when there is an error executing the command", func() {
			It("prints and error message and exits", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", "-sourceInstance=cell-123", "do-the-fandango-for-me")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say(`exec: "do-the-fandango-for-me": executable file not found in \$PATH`))
			})
		})
		Context("when there is an no permission to execute", func() {
			BeforeEach(func() {
				Expect(os.Chmod(chattyProcessPath, 0644)).To(Succeed())
			})
			AfterEach(func() {
				Expect(os.Chmod(chattyProcessPath, 0755)).To(Succeed())
			})
			It("prints and error message and exits", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", "-sourceInstance=cell-123", chattyProcessPath, "chattyArg1", "chattyArg2", "-chattyFlag")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("chatty_process: permission denied"))
			})
		})
	})

	Describe("Flags", func() {
		Context("when -dropsondeDestination is not passed", func() {
			It("should specify the flag is required", func() {
				command := exec.Command(tee2MetronPath, "-sourceInstance=cell-123", chattyProcessPath)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out).To(gbytes.Say("dropsondeDestination flag is required"))
			})
		})

		Context("when -sourceInstance is not passed", func() {
			It("should specify the flag is required", func() {
				command := exec.Command(tee2MetronPath, "-dropsondeDestination=127.0.0.1:4000", chattyProcessPath)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out).To(gbytes.Say("sourceInstance flag is required"))
			})
		})
	})
})

func startFakeMetron() (metronReceivedBufferPtr *[]byte, port string) {
	connection, err := net.ListenPacket("udp", "")
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
			if string(metronReceivedBuffer) != string(make([]byte, maxUpdDatagramSize)) {
				fmt.Fprintf(GinkgoWriter, "\nRead UDP Packet: %s\n", metronReceivedBuffer)
			}
		}
	}()

	_, listenPort, err := net.SplitHostPort(connection.LocalAddr().String())
	Expect(err).NotTo(HaveOccurred())

	return &metronReceivedBuffer, listenPort
}
