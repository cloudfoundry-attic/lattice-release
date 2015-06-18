package main_test

import (
	"net"
	"net/http"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("HealthCheck", func() {
	var (
		server     *ghttp.Server
		serverAddr string
	)

	BeforeEach(func() {
		ip := getNonLoopbackIP()
		server = ghttp.NewUnstartedServer()
		listener, err := net.Listen("tcp", ip+":0")
		Expect(err).NotTo(HaveOccurred())

		server.HTTPTestServer.Listener = listener
		serverAddr = listener.Addr().String()
		server.Start()
	})

	Describe("port healthcheck", func() {
		portHealthCheck := func() *gexec.Session {
			_, port, err := net.SplitHostPort(serverAddr)
			Expect(err).NotTo(HaveOccurred())
			session, err := gexec.Start(exec.Command(healthCheck, "-port", port, "-timeout", "100ms"), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			return session
		}

		Context("when the address is listening", func() {
			It("exits 0 and logs it passed", func() {
				session := portHealthCheck()
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out).To(gbytes.Say("healthcheck passed"))
			})
		})

		Context("when the address is not listening", func() {
			BeforeEach(func() {
				server.Close()
			})

			It("exits 1 and logs it failed", func() {
				session := portHealthCheck()
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Out).To(gbytes.Say("healthcheck failed"))
			})
		})
	})

	Describe("http healthcheck", func() {
		Context("when the healthcheck is properly invoked", func() {
			httpHealthCheck := func() *gexec.Session {
				_, port, err := net.SplitHostPort(serverAddr)
				Expect(err).NotTo(HaveOccurred())
				session, err := gexec.Start(exec.Command(healthCheck, "-uri", "/api/_ping", "-port", port, "-timeout", "100ms"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				return session
			}

			itFailsHttpHealthCheck := func() {
				It("exits 1 and logs it failed", func() {
					session := httpHealthCheck()
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Out).To(gbytes.Say("healthcheck failed"))
				})
			}

			BeforeEach(func() {
				server.RouteToHandler("GET", "/api/_ping", ghttp.VerifyRequest("GET", "/api/_ping"))
			})

			Context("when the address is listening", func() {
				It("exits 0 and logs it passed", func() {
					session := httpHealthCheck()
					Eventually(session).Should(gexec.Exit(0))
					Expect(session.Out).To(gbytes.Say("healthcheck passed"))
				})
			})

			Context("when the address returns error http code", func() {
				BeforeEach(func() {
					server.RouteToHandler("GET", "/api/_ping", ghttp.RespondWith(500, ""))
				})

				itFailsHttpHealthCheck()
			})

			Context("when the address is not listening", func() {
				BeforeEach(func() {
					server.Close()
				})

				itFailsHttpHealthCheck()
			})

			Context("when the server is too slow to respond", func() {
				BeforeEach(func() {
					server.RouteToHandler("GET", "/api/_ping", func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(2 * time.Second)
						w.WriteHeader(http.StatusOK)
					})
				})

				itFailsHttpHealthCheck()
			})
		})
	})
})

func getNonLoopbackIP() string {
	interfaces, err := net.Interfaces()
	Expect(err).NotTo(HaveOccurred())
	for _, intf := range interfaces {
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	Fail("no non-loopback address found")
	panic("non-reachable")
}
