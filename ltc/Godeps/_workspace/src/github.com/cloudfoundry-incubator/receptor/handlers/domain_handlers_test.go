package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	fake_legacy_bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Domain Handlers", func() {
	var (
		logger           lager.Logger
		fakeLegacyBBS    *fake_legacy_bbs.FakeReceptorBBS
		fakeBBS          *fake_bbs.FakeClient
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DomainHandler
	)

	BeforeEach(func() {
		fakeLegacyBBS = new(fake_legacy_bbs.FakeReceptorBBS)
		fakeBBS = new(fake_bbs.FakeClient)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDomainHandler(fakeBBS, fakeLegacyBBS, logger)
	})

	Describe("Upsert", func() {
		var domain string
		var ttl time.Duration

		BeforeEach(func() {
			domain = "domain-1"
			ttl = 1000 * time.Second
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
					Expect(fakeBBS.UpsertDomainCallCount()).To(Equal(1))
					d, ttl := fakeBBS.UpsertDomainArgsForCall(0)
					Expect(d).To(Equal(domain))
					Expect(ttl).To(Equal(ttl))
				})

				It("responds with 204 Status NO CONTENT", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
				})

				It("responds with an empty body", func() {
					Expect(responseRecorder.Body.String()).To(Equal(""))
				})
			})

			Context("when the call to the BBS fails", func() {
				BeforeEach(func() {
					fakeBBS.UpsertDomainReturns(errors.New("ka-boom"))
				})

				It("responds with 500 INTERNAL ERROR", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.UnknownError,
						Message: "ka-boom",
					})

					Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
				})
			})

			Context("when the request corresponds to an invalid domain", func() {
				var validationError = models.ValidationError{}

				BeforeEach(func() {
					fakeBBS.UpsertDomainReturns(validationError)
				})

				It("responds with 400 BAD REQUEST", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
				})

				It("responds with a relevant error message", func() {
					expectedBody, _ := json.Marshal(receptor.Error{
						Type:    receptor.InvalidDomain,
						Message: validationError.Error(),
					})

					Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
				})
			})

			Context("when no Cache-Control header is present", func() {
				BeforeEach(func() {
					req.Header.Del("Cache-Control")
				})

				It("sets the TTL to 0 (inifinite)", func() {
					Expect(fakeBBS.UpsertDomainCallCount()).To(Equal(1))
					d, ttl := fakeBBS.UpsertDomainArgsForCall(0)
					Expect(d).To(Equal(domain))
					Expect(ttl).To(Equal(time.Duration(0)))
				})
			})

			Context("when Cache-Control header is present", func() {
				Context("when no max-age is included", func() {
					BeforeEach(func() {
						req.Header.Set("Cache-Control", "public")
					})

					It("fails with an error", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
						expectedBody, _ := json.Marshal(receptor.Error{
							Type:    receptor.InvalidRequest,
							Message: handlers.ErrMaxAgeMissing.Error(),
						})

						Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
					})
				})
			})
		})

		Context("when the request is missing the domain", func() {
			BeforeEach(func() {
				handler.Upsert(responseRecorder, newTestRequest(""))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    receptor.InvalidRequest,
					Message: handlers.ErrDomainMissing.Error(),
				})

				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
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
				Expect(fakeBBS.DomainsCallCount()).To(Equal(1))
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of domains", func() {
				response := []string{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(ConsistOf(domains))
			})
		})

		Context("when the BBS returns no domains", func() {
			BeforeEach(func() {
				fakeBBS.DomainsReturns([]string{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Body.String()).To(Equal("[]"))
			})
		})

		Context("when reading from the BBS fails", func() {
			BeforeEach(func() {
				fakeBBS.DomainsReturns([]string{}, errors.New("Something went wrong"))
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var receptorError receptor.Error
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &receptorError)
				Expect(err).NotTo(HaveOccurred())

				Expect(receptorError).To(Equal(receptor.Error{
					Type:    receptor.UnknownError,
					Message: "Something went wrong",
				}))

			})
		})
	})
})
