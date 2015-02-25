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
				Ω(res.StatusCode).Should(Equal(http.StatusOK))
				value := res.Header.Get("Access-Control-Allow-Origin")
				Ω(value).Should(Equal(req.Header.Get("Origin")))
			})
		})

		Context("when a request with the OPTION method is made against any endpoint", func() {
			var res *http.Response
			var req *http.Request

			JustBeforeEach(func() {
				req, res = doOptionsRequest()
			})

			It("returns 200 OK", func() {
				Ω(res.StatusCode).Should(Equal(http.StatusOK))
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
				Ω(isSet).Should(BeFalse())
			})
		})
	})
})

func doGetRequest() (*http.Request, *http.Response) {
	reqGen := rata.NewRequestGenerator("http://"+receptorAddress, receptor.Routes)

	req, err := reqGen.CreateRequest(receptor.TasksRoute, nil, nil)
	Ω(err).ShouldNot(HaveOccurred())

	req.Header.Set("Origin", "example.com")
	req.SetBasicAuth(username, password)
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	Ω(err).ShouldNot(HaveOccurred())

	return req, res
}

func doOptionsRequest() (*http.Request, *http.Response) {
	req, err := http.NewRequest("OPTIONS", "http://"+receptorAddress, nil)
	Ω(err).ShouldNot(HaveOccurred())
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	Ω(err).ShouldNot(HaveOccurred())
	return req, res
}
