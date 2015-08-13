package main_test

import (
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

var _ = Describe("s3tool", func() {
	Describe("invalid action", func() {
		It("prints an error message and exits", func() {
			command := exec.Command(s3toolPath, "invalid")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(3))
			Expect(session.Out).To(gbytes.Say("Usage: s3tool \\[get\\|put\\|delete\\] arguments..."))
		})
	})

	Describe("delete", func() {
		var (
			httpStatusCode int
			fakeServer     *ghttp.Server
		)

		BeforeEach(func() {
			httpStatusCode = 200
			fakeServer = ghttp.NewServer()
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/bucket/key"),
				ghttp.RespondWithPtr(&httpStatusCode, nil),
				func(_ http.ResponseWriter, req *http.Request) {
					auth := req.Header.Get(http.CanonicalHeaderKey("Authorization"))
					Expect(auth).To(HavePrefix("AWS "))
				},
			))
		})

		AfterEach(func() {
			fakeServer.Close()
		})

		It("does a DELETE to the S3 server to delete the file", func() {
			command := exec.Command(s3toolPath, "delete", "access", "secret", fakeServer.URL(), "bucket", "key")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Deleted s3://bucket/key."))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when the S3 delete fails", func() {
			It("prints an error message and exits", func() {
				httpStatusCode = 404

				command := exec.Command(s3toolPath, "delete", "access", "secret", fakeServer.URL(), "bucket", "key")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error deleting s3://bucket/key: "))
			})
		})

		Context("when arguments for the delete action are invalid", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(s3toolPath, "delete", "invalid")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("Usage: s3tool delete s3AccessKey s3SecretKey blobStoreURL s3Bucket s3Path"))
			})
		})
	})

	Describe("get", func() {
		var (
			httpStatusCode int
			httpResponse   string
			fakeServer     *ghttp.Server
		)

		BeforeEach(func() {
			fakeServer = ghttp.NewServer()
			httpStatusCode = 200
			httpResponse = ""
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/bucket/key"),
				ghttp.RespondWithPtr(&httpStatusCode, &httpResponse),
				func(_ http.ResponseWriter, req *http.Request) {
					auth := req.Header.Get(http.CanonicalHeaderKey("Authorization"))
					Expect(auth).To(HavePrefix("AWS "))
				},
			))
		})

		AfterEach(func() {
			fakeServer.Close()
		})

		It("does a GET to the S3 server and saves the resulting response", func() {
			httpResponse = "some-file-contents"

			tmpFile, err := ioutil.TempFile(os.TempDir(), "downloadedFile")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(tmpFile.Name())

			command := exec.Command(s3toolPath, "get", "access", "secret", fakeServer.URL(), "bucket", "key", tmpFile.Name())
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Downloaded s3://bucket/key to %s", tmpFile.Name()))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			Expect(ioutil.ReadFile(tmpFile.Name())).To(Equal([]byte("some-file-contents")))
		})

		Context("when the S3 download fails", func() {
			It("prints an error message and exits", func() {
				httpStatusCode = 404

				command := exec.Command(s3toolPath, "get", "access", "secret", fakeServer.URL(), "bucket", "key", "some-missing-file")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error downloading s3://bucket/key:"))

				_, err = os.Stat("some-missing-file")
				Expect(err).To(MatchError("stat some-missing-file: no such file or directory"))
			})
		})

		Context("when the destination open fails", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(s3toolPath, "get", "access", "secret", fakeServer.URL(), "bucket", "key", "/")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error opening /: "))
			})
		})

		Context("when the command is missing", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(s3toolPath, "get", "invalid")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("Usage: s3tool get s3AccessKey s3SecretKey blobStoreURL s3Bucket s3Path destinationFilePath"))
			})
		})
	})

	Describe("put", func() {
		var (
			httpStatusCode int
			fakeServer     *ghttp.Server
			httpBody       []byte
			tmpFile        *os.File
		)

		BeforeEach(func() {
			httpStatusCode = 200
			fakeServer = ghttp.NewServer()
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/bucket/key"),
				ghttp.RespondWithPtr(&httpStatusCode, nil),
				func(_ http.ResponseWriter, req *http.Request) {
					auth := req.Header.Get(http.CanonicalHeaderKey("Authorization"))
					Expect(auth).To(HavePrefix("AWS "))
					var err error
					httpBody, err = ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
				},
			))

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

		It("does a PUT to the S3 server to upload the file", func() {
			command := exec.Command(s3toolPath, "put", "access", "secret", fakeServer.URL(), "bucket", "key", tmpFile.Name())
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Uploaded " + tmpFile.Name() + " to s3://bucket/key."))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			Expect(httpBody).To(Equal([]byte("some-file-contents")))
		})

		Context("when the S3 upload fails", func() {
			It("prints an error message and exits", func() {
				httpStatusCode = 404

				command := exec.Command(s3toolPath, "put", "access", "secret", fakeServer.URL(), "bucket", "key", tmpFile.Name())
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error uploading " + tmpFile.Name() + ": 404NotFound: 404 Not Found"))
			})
		})

		Context("when the source file cannot be read", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(s3toolPath, "put", "access", "secret", fakeServer.URL(), "bucket", "key", "some-missing-file")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(2))
				Expect(session.Out).To(gbytes.Say("Error opening some-missing-file: open some-missing-file: no such file or directory"))
			})
		})

		Context("when the command is missing", func() {
			It("prints an error message and exits", func() {
				command := exec.Command(s3toolPath, "put", "invalid")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(3))
				Expect(session.Out).To(gbytes.Say("Usage: s3tool put s3AccessKey s3SecretKey blobStoreURL s3Bucket s3Path fileToUpload"))
			})
		})
	})
})
