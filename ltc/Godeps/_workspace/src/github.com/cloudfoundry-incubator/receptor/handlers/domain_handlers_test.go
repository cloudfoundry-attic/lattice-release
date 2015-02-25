package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Domain Handlers", func() {
	var (
		logger           lager.Logger
		fakeBBS          *fake_bbs.FakeReceptorBBS
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DomainHandler
	)

	BeforeEach(func() {
		fakeBBS = new(fake_bbs.FakeReceptorBBS)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDomainHandler(fakeBBS, logger)
	})

	Describe("Upsert", func() {
		var domain string
		var ttlInSeconds int

		BeforeEach(func() {
			domain = "domain-1"
			ttlInSeconds = 1000
		})

		Context("with a structured request", func() {
			var req *http.Request

			BeforeEach(func() {
				req = newTestRequest("")
				req.URL.RawQuery = url.Values{":domain": []string{domain}}.Encode()
				req.Header["Cache-Control"] = []string{"public", "max-age=1000"}
			})

			JustBeforeEach(func() {
				handler.Upsert(responseRecorder, req)
			})

			Context("when the call to the BBS succeeds", func() {
				It("calls Upsert on the BBS", func() {
					Ω(fakeBBS.UpsertDomainCallCount()).Should(Equal(1))
					d, ttl := fakeBBS.UpsertDomainArgsForCall(0)
					Ω(d).To(Equal(domain))
					Ω(ttl).To(Equal(ttlInSeconds))
				})

				It("responds with 204 Status NO CONTENT", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusNoContent))
				})

				It("responds with an empty body", func() {
					Ω(responseRecorder.Body.String()).Should(Equal(""))
				})
			})

			Context("when the call to the BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.UpsertDomainReturns(errors.New("ka-boom"))
				})

				It("responds with 500 INTERNAL ERROR", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.UnknownError,
						Message: "ka-boom",
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})

			Context("when the request corresponds to an invalid domain", func() {
				var validationError = models.ValidationError{}

				BeforeEach(func() {
					fakeBBS.UpsertDomainReturns(validationError)
				})

				It("responds with 400 BAD REQUEST", func() {
					Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.InvalidDomain,
						Message: validationError.Error(),
					})

					Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
				})
			})

			Context("when no Cache-Control header is present", func() {
				BeforeEach(func() {
					req.Header.Del("Cache-Control")
				})

				It("sets the TTL to 0 (inifinite)", func() {
					Ω(fakeBBS.UpsertDomainCallCount()).Should(Equal(1))
					d, ttl := fakeBBS.UpsertDomainArgsForCall(0)
					Ω(d).To(Equal(domain))
					Ω(ttl).To(Equal(0))
				})
			})

			Context("when Cache-Control header is present", func() {
				Context("when no max-age is included", func() {
					BeforeEach(func() {
						req.Header.Set("Cache-Control", "public")
					})

					It("fails with an error", func() {
						Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
						expectedBody, _ := json.Marshal(receptor.Error{
							Type:    receptor.InvalidRequest,
							Message: handlers.ErrMaxAgeMissing.Error(),
						})

						Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
					})
				})
			})
		})

		Context("when the request is missing the domain", func() {
			BeforeEach(func() {
				handler.Upsert(responseRecorder, newTestRequest(""))
			})

			It("responds with 400 BAD REQUEST", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: handlers.ErrDomainMissing.Error(),
				})

				Ω(responseRecorder.Body.String()).Should(Equal(string(expectedBody)))
			})
		})
	})

	Describe("GetAll", func() {
		var domains []string

		BeforeEach(func() {
			domains = []string{"domain-a", "domain-b"}
		})

		JustBeforeEach(func() {
			handler.GetAll(responseRecorder, newTestRequest(""))
		})

		Context("when reading domains from BBS succeeds", func() {
			BeforeEach(func() {
				fakeBBS.DomainsReturns(domains, nil)
			})

			It("call the BBS to retrieve the domains", func() {
				Ω(fakeBBS.DomainsCallCount()).Should(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns a list of domains", func() {
				response := []string{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(ConsistOf(domains))
			})
		})

		Context("when the BBS returns no domains", func() {
			BeforeEach(func() {
				fakeBBS.DomainsReturns([]string{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Ω(responseRecorder.Body.String()).Should(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.DomainsReturns([]string{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Ω(responseRecorder.Code).Should(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var receptorError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &receptorError)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(receptorError).Should(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))
			})
		})
	})
})
