package main_test

import (
	"encoding/base64"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Auth", func() {
	JustBeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Context("when a request without auth is made", func() {
		var req *http.Request
		var res *http.Response

		BeforeEach(func() {
			var err error
			req, err = http.NewRequest("GET", "http://"+receptorAddress, nil)
			Ω(err).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			var err error

			res, err = http.DefaultClient.Do(req)
			Ω(err).ShouldNot(HaveOccurred())

			res.Body.Close()
		})

		Context("when the username and password have been set", func() {
			It("returns 401 for all requests", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusUnauthorized))
			})
		})

		Context("when the username and password have been set via the receptor_authorization cookie", func() {
			BeforeEach(func() {
				req.AddCookie(&http.Cookie{
					Name:  receptor.AuthorizationCookieName,
					Value: basicAuthHeader(username, password),
				})
			})

			It("does not return 401", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})

		Context("and the username and password have not been set", func() {
			BeforeEach(func() {
				receptorArgs.Username = ""
				receptorArgs.Password = ""
				receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
			})

			It("does not return 401", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
	})
})

func basicAuthHeader(username, password string) string {
	credentials := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}
