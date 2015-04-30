package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Auth Cookie Handler", func() {
	var (
		logger           *lagertest.TestLogger
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.AuthCookieHandler
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewAuthCookieHandler(logger)
	})

	Describe("GenerateCookie", func() {
		const basicAuthValue = "some-base64-data"
		var request *http.Request

		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("", "", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			handler.GenerateCookie(responseRecorder, request)
		})

		Context("when the Authorization header is present", func() {
			BeforeEach(func() {
				request.Header.Set("Authorization", basicAuthValue)
				request.Header.Set("Host", "receptor.diego.com")
			})

			It("sends a Cookie with the Authenication creds", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
				response := http.Response{Header: responseRecorder.Header()}
				Expect(response.Cookies()).To(HaveLen(1))
				Expect(response.Cookies()[0].String()).To(Equal(
					(&http.Cookie{
						Name:     receptor.AuthorizationCookieName,
						Value:    basicAuthValue,
						Domain:   "",
						MaxAge:   0,
						HttpOnly: true,
					}).String(),
				))
			})
		})

		Context("when the Authorization header is not set", func() {
			BeforeEach(func() {
				request.Header.Set("Host", "receptor.diego.com")
			})

			It("responds with a 204 without setting a cookie", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
				response := http.Response{Header: responseRecorder.Header()}
				Expect(response.Cookies()).To(HaveLen(0))
			})
		})
	})
})
