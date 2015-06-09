package receptor_test

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/receptor"
)

var (
	registryHost       string
	fakeReceptorServer *ghttp.Server
	client             receptor.Client
)

var _ = Describe("Receptor Client", func() {

	BeforeEach(func() {
		fakeReceptorServer = ghttp.NewServer()
		client = receptor.NewClient(fakeReceptorServer.URL())
	})

	AfterEach(func() {
		fakeReceptorServer.Close()
	})

	Describe("Client Request", func() {
		var lrpResponse []receptor.DesiredLRPResponse

		BeforeEach(func() {
			lrpResponse = []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{
					ProcessGuid: "some-guid",
					Domain:      "diego",
				},
			}

			fakeReceptorServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/desired_lrps", "domain=diego"),
				ghttp.VerifyContentType(receptor.JSONContentType),
				ghttp.RespondWithJSONEncoded(http.StatusOK, lrpResponse),
			))
		})

		It("sends a valid request from the client to the receptor", func() {
			response, err := client.DesiredLRPsByDomain("diego")

			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(lrpResponse))
		})
	})

	Describe("Content Type Validation and Error Handling", func() {
		var (
			httpHeaders  http.Header
			statusCode   int
			responseBody string
		)

		verifyReceptorError := func(actual error, expected receptor.Error) {
			Expect(actual).To(HaveOccurred())
			Expect(actual).To(BeAssignableToTypeOf(receptor.Error{}))
			Expect(actual.(receptor.Error).Type).To(Equal(expected.Type))
			Expect(actual.(receptor.Error).Message).To(Equal(expected.Message))
		}

		BeforeEach(func() {
			httpHeaders = http.Header{}
			fakeReceptorServer.AppendHandlers(
				ghttp.RespondWithPtr(&statusCode, &responseBody, httpHeaders),
			)
		})

		Context("when the client receives json content", func() {
			BeforeEach(func() {
				httpHeaders[receptor.ContentTypeHeader] = []string{receptor.JSONContentType}
			})

			Context("when the http status code is not successful", func() {
				It("returns a json-encoded error from the server", func() {
					statusCode = http.StatusNotFound
					lrpError := receptor.Error{
						Type:    receptor.DesiredLRPNotFound,
						Message: "Desired LRP with guid 'unicorns' not found",
					}
					responseBytes, err := json.Marshal(lrpError)
					Expect(err).ToNot(HaveOccurred())
					responseBody = string(responseBytes)

					By("triggering a receptor error that returns a 404 status code")
					_, err = client.GetDesiredLRP("unicorns")

					verifyReceptorError(err, lrpError)
				})

				It("returns an invalid json error for invalid json", func() {
					statusCode = http.StatusNotFound
					responseBody = `{"key": "value}`

					_, err := client.GetDesiredLRP("unicorns")

					Expect(err).To(HaveOccurred())
					Expect(err.(receptor.Error).Type).To(Equal(receptor.InvalidJSON))
				})
			})

			Context("when the http status code is successful", func() {
				It("returns an invalid json error for invalid json", func() {
					statusCode = http.StatusOK
					responseBody = `{"key": "value}`

					_, err := client.GetDesiredLRP("unicorns")

					Expect(err).To(HaveOccurred())
					Expect(err.(receptor.Error).Type).To(Equal(receptor.InvalidJSON))
				})
			})
		})

		Context("when the client receives non-json content", func() {
			BeforeEach(func() {
				httpHeaders[receptor.ContentTypeHeader] = []string{"text/plain; charset=utf-8"}
			})

			Context("when the http status code is 404", func() {
				It("returns an invalid response error", func() {
					statusCode = http.StatusNotFound
					responseBody = "404 page not found"

					_, err := client.DesiredLRPs()

					Expect(err).To(HaveOccurred())
					Expect(err.(receptor.Error).Type).To(Equal(receptor.InvalidResponse))
				})

				Context("when there is an x-cf-routererror", func() {
					It("returns a router error", func() {
						statusCode = http.StatusNotFound
						httpHeaders[receptor.XCfRouterErrorHeader] = []string{"unknown_route"}
						expectedErrorMessage := "unknown_route"

						_, err := client.DesiredLRPs()

						verifyReceptorError(err, receptor.Error{receptor.RouterError, expectedErrorMessage})
					})
				})
			})

			Context("when the http status code is not successful and not 404", func() {
				It("returns an invalid response error", func() {
					statusCode = http.StatusGone
					httpHeaders[receptor.ContentTypeHeader] = []string{"image/gif"}
					expectedErrorMessage := "Invalid Response with status code: 410"

					_, err := client.DesiredLRPs()

					verifyReceptorError(err, receptor.Error{receptor.InvalidResponse, expectedErrorMessage})
				})
			})
		})

		Context("when the client receives empty/messy content types", func() {
			Context("when the response is missing the content-type header", func() {
				It("does not return error for successful http status codes", func() {
					statusCode = http.StatusNoContent
					httpHeaders[receptor.ContentTypeHeader] = []string{}
					responseBody = ""

					_, err := client.DesiredLRPs()

					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the content-type is formatted funny", func() {
				It("figures it out and works", func() {
					statusCode = http.StatusOK
					httpHeaders[receptor.ContentTypeHeader] = []string{" aPPLiCaTioN/JSoN  ; charset=ISO-8859-8"}
					lrpResponse := receptor.DesiredLRPResponse{
						ProcessGuid: "some-guid",
						Domain:      "diego",
					}
					responseBytes, err := json.Marshal(lrpResponse)
					Expect(err).ToNot(HaveOccurred())
					responseBody = string(responseBytes)

					response, err := client.GetDesiredLRP("some-guid")

					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(Equal(lrpResponse))
				})
			})
		})
	})
})
