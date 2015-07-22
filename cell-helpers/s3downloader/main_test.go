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

var _ = Describe("s3downloader", func() {
	var (
		s3downloaderPath string
		httpStatusCode   int
		httpResponse     string
		fakeServer       *ghttp.Server
	)

	BeforeSuite(func() {
		var err error
		s3downloaderPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/s3downloader")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		fakeServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/bucket/key"),
			ghttp.RespondWithPtr(&httpStatusCode, &httpResponse),
			func(w http.ResponseWriter, req *http.Request) {
				auth, ok := req.Header[http.CanonicalHeaderKey("Authorization")]
				Expect(ok).To(BeTrue())
				Expect(auth).To(ConsistOf(HavePrefix("AWS ")))
			},
		))
	})

	AfterEach(func() {
		fakeServer.Close()
	})

	It("does a GET to the S3 server to download the url", func() {
		httpStatusCode = 200
		httpResponse = "xyz"

		tmpFile, err := ioutil.TempFile(os.TempDir(), "downloadedFile")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		command := exec.Command(s3downloaderPath, "access", "secret", fakeServer.URL(), "bucket", "key", tmpFile.Name())

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("Downloaded s3://bucket/key to " + tmpFile.Name() + "."))

		Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))

		Eventually(session.Exited).Should(BeClosed())
		Expect(session.ExitCode()).To(Equal(0))
	})

	It("works with parameters that start with a leading -", func() {
		httpStatusCode = 200
		httpResponse = "xyz"

		tmpFile, err := ioutil.TempFile(os.TempDir(), "downloadedFile")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		command := exec.Command(s3downloaderPath, "-access", "-secret", fakeServer.URL(), "bucket", "key", tmpFile.Name())

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("Downloaded s3://bucket/key to " + tmpFile.Name() + "."))

		Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))

		Eventually(session.Exited).Should(BeClosed())
		Expect(session.ExitCode()).To(Equal(0))
	})

	Context("when the S3 download fails", func() {
		It("prints an error message and exits", func() {
			httpStatusCode = 404
			httpResponse = ""

			command := exec.Command(s3downloaderPath, "access", "secret", fakeServer.URL(), "bucket", "key", "downloadFilePath")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Error downloading s3://bucket/key: "))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(2))

			_, err = os.Stat("downloadFilePath")
			Expect(os.IsExist(err)).To(BeFalse())
		})
	})

	Context("when the destination open fails", func() {
		It("prints an error message and exits", func() {
			httpStatusCode = 200
			httpResponse = "xyz"

			command := exec.Command(s3downloaderPath, "access", "secret", fakeServer.URL(), "bucket", "key", "/")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Error opening /: "))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(2))
		})
	})

	Context("when the command is missing", func() {
		It("prints an error message and exits", func() {
			command := exec.Command(s3downloaderPath)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Usage: s3downloader s3AccessKey s3SecretKey httpProxy s3Bucket s3Path destinationFilePath"))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(3))
		})
	})

})
