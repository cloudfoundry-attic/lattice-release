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

var _ = Describe("s3uploader", func() {
	var (
		s3uploaderPath string
		httpStatusCode int
		fakeServer     *ghttp.Server
	)

	BeforeSuite(func() {
		var err error
		s3uploaderPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/s3uploader")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		var emptyResponse = ""
		fakeServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("PUT", ContainSubstring("/bucket/")),
			ghttp.RespondWithPtr(&httpStatusCode, &emptyResponse),
			func(w http.ResponseWriter, req *http.Request) {
				auth, ok := req.Header[http.CanonicalHeaderKey("Authorization")]
				Expect(ok).To(BeTrue())
				Expect(auth).To(ConsistOf(HavePrefix("AWS access:")))
			},
		))
	})

	AfterEach(func() {
		fakeServer.Close()
	})

	It("does a PUT to the S3 server to upload the file", func() {
		httpStatusCode = 200

		tmpFile, err := ioutil.TempFile(os.TempDir(), "fileToUpload")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		tmpFile.Write([]byte("abcd"))
		tmpFile.Close()

		command := exec.Command(s3uploaderPath, "access", "secret", fakeServer.URL(), "bucket", "key", tmpFile.Name())

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("Uploaded " + tmpFile.Name() + " to s3://bucket/key."))

		Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))

		Eventually(session.Exited).Should(BeClosed())
		Expect(session.ExitCode()).To(Equal(0))
	})

	Context("when the S3 upload fails", func() {
		It("prints an error message and exits", func() {
			httpStatusCode = 404

			command := exec.Command(s3uploaderPath, "access", "secret", fakeServer.URL(), "bucket", "key", "fileToUpload")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Error stat'ing fileToUpload: "))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(2))
		})
	})

	Context("when the command is missing", func() {
		It("prints an error message and exits", func() {
			command := exec.Command(s3uploaderPath)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Usage: s3uploader s3AccessKey s3SecretKey httpProxy s3Bucket s3Path fileToUpload"))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(3))
		})
	})

})
