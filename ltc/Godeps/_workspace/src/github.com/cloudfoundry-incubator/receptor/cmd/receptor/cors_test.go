package main_test

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/cmd/receptor/testrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/rata"
)

var _ = Describe("CORS support", func() {
	JustBeforeEach(func() {
		receptorRunner = testrunner.New(receptorBinPath, receptorArgs)
		receptorProcess = ginkgomon.Invoke(receptorRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
	})

	Context("when CORS support is enabled", func() {
		BeforeEach(func() {
			receptorArgs.CORSEnabled = true
		})

		Context("when a request contains a valid ORIGIN header", func() {
			var res *http.Response
			var req *http.Request

			JustBeforeEach(func() {
				req, res = doGetRequest()
			})

			It("responds to with a ACAO header containg the ORIGIN header value", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				value := res.Header.Get("Access-Control-Allow-Origin")
				Expect(value).To(Equal(req.Header.Get("Origin")))
			})
		})

		Context("when a request with the OPTION method is made against any endpoint", func() {
			var res *http.Response
			var req *http.Request

			JustBeforeEach(func() {
				req, res = doOptionsRequest()
			})

			It("returns 200 OK", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Context("when CORS support is disabled", func() {
		BeforeEach(func() {
			receptorArgs.CORSEnabled = false
		})

		Context("when a request contains a valid ORIGIN header", func() {
			var res *http.Response
			var req *http.Request

			JustBeforeEach(func() {
				req, res = doGetRequest()
			})

			It("responds to without a ACAO header", func() {
				_, isSet := res.Header["Access-Control-Allow-Origin"]
				Expect(isSet).To(BeFalse())
			})
		})
	})
})

func doGetRequest() (*http.Request, *http.Response) {
	reqGen := rata.NewRequestGenerator("http://"+receptorAddress, receptor.Routes)

	req, err := reqGen.CreateRequest(receptor.TasksRoute, nil, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Set("Origin", "example.com")
	req.SetBasicAuth(username, password)
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	Expect(err).NotTo(HaveOccurred())

	return req, res
}

func doOptionsRequest() (*http.Request, *http.Response) {
	req, err := http.NewRequest("OPTIONS", "http://"+receptorAddress, nil)
	Expect(err).NotTo(HaveOccurred())
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	return req, res
}
