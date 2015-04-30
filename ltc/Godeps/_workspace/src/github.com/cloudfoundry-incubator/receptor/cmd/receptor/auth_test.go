package main_test

import (
	"encoding/base64"
	"net/http"
	"net/url"

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
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			var err error

			res, err = http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())

			res.Body.Close()
		})

		Context("when the username and password have been set", func() {
			It("returns 401 for all requests", func() {
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
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
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Context("and the username and password have not been set", func() {
			BeforeEach(func() {
				receptorArgs.Username = ""
				receptorArgs.Password = ""
				receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
			})

			It("does not return 401", func() {
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("AuthCookie", func() {
			BeforeEach(func() {
				req.URL.Path = "/v1/auth_cookie"
				req.URL.User = url.UserPassword(username, password)
				req.Method = "POST"
			})

			It("returns an authenication cookie", func() {
				Expect(res.StatusCode).To(Equal(http.StatusNoContent))
				Expect(res.Cookies()).To(HaveLen(1))
				Expect(res.Cookies()[0].String()).To(Equal((&http.Cookie{
					Name:     receptor.AuthorizationCookieName,
					Value:    "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
					MaxAge:   0,
					HttpOnly: true,
				}).String()))

				By("Using the cookie to make a request")

				req2, err := http.NewRequest("GET", "http://"+receptorAddress, nil)
				Expect(err).NotTo(HaveOccurred())

				req2.AddCookie(res.Cookies()[0])

				res2, err := http.DefaultClient.Do(req2)
				Expect(err).NotTo(HaveOccurred())

				Expect(res2.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})
})

func basicAuthHeader(username, password string) string {
	credentials := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}
