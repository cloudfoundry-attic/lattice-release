package main_test

import (
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("s3deleter", func() {
	var (
		s3DeleterPath  string
		httpStatusCode int
		fakeServer     *ghttp.Server
	)

	BeforeSuite(func() {
		var err error
		s3DeleterPath, err = gexec.Build("github.com/cloudfoundry-incubator/lattice/cell-helpers/s3deleter")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		var emptyResponse = ""
		fakeServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("DELETE", ContainSubstring("/bucket/key")),
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

	It("does a DELETE to the S3 server to delete the file", func() {
		httpStatusCode = 200

		command := exec.Command(s3DeleterPath, "access", "secret", fakeServer.URL(), "bucket", "key")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("Deleted s3://bucket/key."))

		Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))

		Eventually(session.Exited).Should(BeClosed())
		Expect(session.ExitCode()).To(Equal(0))
	})

	Context("when the S3 delete fails", func() {
		It("prints an error message and exits", func() {
			httpStatusCode = 404

			command := exec.Command(s3DeleterPath, "access", "secret", fakeServer.URL(), "bucket", "key")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Error deleting s3://bucket/key: "))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(2))
		})
	})

	Context("when the command is missing", func() {
		It("prints an error message and exits", func() {
			command := exec.Command(s3DeleterPath)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(gbytes.Say("Usage: s3deleter s3AccessKey s3SecretKey httpProxy s3Bucket s3Path"))
			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(3))
		})
	})

})
