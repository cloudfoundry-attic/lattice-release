package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("davtool", func() {
	var davtoolPath string

	BeforeSuite(func() {
		var err error
		davtoolPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/davtool")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	Describe("invalid action", func() {
		It("prints an error message and exits", func() {
			command := exec.Command(davtoolPath, "invalid")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(3))
			Expect(session.Out).To(gbytes.Say("Usage: davtool \\[put\\|delete\\] arguments..."))
		})
	})

	Describe("delete", func() {
		var (
			httpStatusCode              int
			fakeServer                  *ghttp.Server
			fakeServerURL, sanitizedURL string
		)

		BeforeEach(func() {
			httpStatusCode = http.StatusNoContent
			fakeServer = ghttp.NewServer()
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/blobs/path"),
				ghttp.RespondWithPtr(&httpStatusCode, nil),
				ghttp.VerifyBasicAuth("user", "pass"),
			))
			fakeServerURL = fmt.Sprintf("http://%s:%s@%s%s", "user", "pass", fakeServer.Addr(), "/blobs/path")
			sanitizedURL = fmt.Sprintf("http://%s%s", fakeServer.Addr(), "/blobs/path")
		})

		AfterEach(func() {
			fakeServer.Close()
		})

		It("does a DELETE to the DAV server to delete the file", func() {
			command := exec.Command(davtoolPath, "delete", fakeServerURL)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("Deleted %s.", sanitizedURL))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when the DAV delete fails", func() {
			It("prints an error message and exits", func() {
				httpStatusCode = 404
				command := exec.Command(davtoolPath, "delete", fakeServerURL)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error deleting %s: ", sanitizedURL))

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when arguments for the delete action are invalid", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(davtoolPath, "delete")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("Usage: davtool delete url"))
			})
		})
	})

	Describe("put", func() {
		var (
			httpStatusCode              int
			fakeServer                  *ghttp.Server
			httpBody                    []byte
			tmpFile                     *os.File
			fakeServerURL, sanitizedURL string
		)

		BeforeEach(func() {
			httpStatusCode = http.StatusCreated
			fakeServer = ghttp.NewServer()
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/blobs/path"),
				ghttp.RespondWithPtr(&httpStatusCode, nil),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.VerifyHeaderKV("Content-Length", "18"),
				func(_ http.ResponseWriter, request *http.Request) {
					var err error
					httpBody, err = ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
				},
			))

			fakeServerURL = fmt.Sprintf("http://%s:%s@%s%s", "user", "pass", fakeServer.Addr(), "/blobs/path")
			sanitizedURL = fmt.Sprintf("http://%s%s", fakeServer.Addr(), "/blobs/path")

			var err error
			tmpFile, err = ioutil.TempFile(os.TempDir(), "fileToUpload")
			Expect(err).NotTo(HaveOccurred())

			tmpFile.Write([]byte("some-file-contents"))
			Expect(tmpFile.Close()).To(Succeed())
		})

		AfterEach(func() {
			fakeServer.Close()
			Expect(os.Remove(tmpFile.Name())).To(Succeed())
		})

		It("does a PUT to the DAV server to create a new file", func() {
			command := exec.Command(davtoolPath, "put", fakeServerURL, tmpFile.Name())
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Uploaded %s to %s.", tmpFile.Name(), sanitizedURL))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			Expect(httpBody).To(Equal([]byte("some-file-contents")))
		})

		It("does a PUT to the DAV server to replace an existing file", func() {
			httpStatusCode = 200

			command := exec.Command(davtoolPath, "put", fakeServerURL, tmpFile.Name())
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Uploaded %s to %s.", tmpFile.Name(), sanitizedURL))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			Expect(httpBody).To(Equal([]byte("some-file-contents")))
		})

		Context("when the DAV upload fails", func() {
			It("prints an error message and exits", func() {
				httpStatusCode = 404
				command := exec.Command(davtoolPath, "put", fakeServerURL, tmpFile.Name())
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error uploading " + tmpFile.Name() + ": 404 Not Found"))

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the source file cannot be read", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(davtoolPath, "put", fakeServerURL, "some-missing-file")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error opening some-missing-file: open some-missing-file: no such file or directory"))
			})
		})

		Context("when the command is missing", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(davtoolPath, "put", "invalid")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("Usage: davtool put url fileToUpload"))
			})
		})
	})
})
