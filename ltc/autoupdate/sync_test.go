package autoupdate_test

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/autoupdate"
	"github.com/cloudfoundry-incubator/lattice/ltc/autoupdate/mocks"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("Sync", func() {
	var (
		fakeFileSwapper *mocks.FakeFileSwapper

		fakeServer *ghttp.Server
		config     *config_package.Config
		sync       *autoupdate.Sync

		ltcTempFile *os.File
	)

	BeforeEach(func() {
		fakeFileSwapper = &mocks.FakeFileSwapper{}

		fakeServer = ghttp.NewServer()
		fakeServerURL, err := url.Parse(fakeServer.URL())
		Expect(err).NotTo(HaveOccurred())

		fakeServerHost, fakeServerPort, err := net.SplitHostPort(fakeServerURL.Host)
		Expect(err).NotTo(HaveOccurred())

		ltcTempFile, err = ioutil.TempFile("", "fake-ltc")
		Expect(err).NotTo(HaveOccurred())

		config = config_package.New(nil)
		config.SetTarget(fakeServerHost + ".xip.io:" + fakeServerPort)
		sync = autoupdate.NewSync(fakeFileSwapper)
	})

	AfterEach(func() {
		Expect(os.Remove(ltcTempFile.Name())).To(Succeed())
	})

	Describe("#SyncLTC", func() {
		It("should download ltc from the target and swap it with ltc", func() {
			fakeServer.RouteToHandler("GET", "/v1/sync/amiga/ltc", ghttp.CombineHandlers(
				ghttp.RespondWith(200, []byte{0x01, 0x02, 0x03}, http.Header{
					"Content-Type":   []string{"application/octet-stream"},
					"Content-Length": []string{"3"},
				}),
			))

			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(tmpFile.Name())

			fakeFileSwapper.GetTempFileReturns(tmpFile, nil)

			sync.SyncLTC(ltcTempFile.Name(), "amiga", config)

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))

			Expect(fakeFileSwapper.GetTempFileCallCount()).To(Equal(1))

			Expect(fakeFileSwapper.SwapTempFileCallCount()).To(Equal(1))
			actualDest, actualSrc := fakeFileSwapper.SwapTempFileArgsForCall(0)
			Expect(actualDest).To(Equal(ltcTempFile.Name()))
			Expect(actualSrc).To(Equal(tmpFile.Name()))

			tmpFile, err = os.OpenFile(tmpFile.Name(), os.O_RDONLY, 0)
			Expect(err).NotTo(HaveOccurred())

			bytes, err := ioutil.ReadAll(tmpFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytes).To(Equal([]byte{0x01, 0x02, 0x03}))
		})

		Context("when the http request returns a non-200 status", func() {
			It("should return an error", func() {
				fakeServer.RouteToHandler("GET", "/v1/sync/amiga/ltc", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "", nil),
				))

				err := sync.SyncLTC(ltcTempFile.Name(), "amiga", config)
				Expect(err).To(MatchError(HavePrefix("failed to download ltc")))
			})
		})

		Context("when the http request fails", func() {
			It("should return an error", func() {
				config.SetTarget("localhost:1")

				err := sync.SyncLTC(ltcTempFile.Name(), "amiga", config)
				Expect(err).To(MatchError(HavePrefix("failed to connect to receptor")))
			})
		})

		Context("when opening the temp file fails", func() {
			It("should return an error", func() {
				fakeServer.RouteToHandler("GET", "/v1/sync/amiga/ltc", ghttp.CombineHandlers(
					ghttp.RespondWith(200, []byte{0x01, 0x02, 0x03}, http.Header{
						"Content-Type":   []string{"application/octet-stream"},
						"Content-Length": []string{"3"},
					}),
				))

				fakeFileSwapper.GetTempFileReturns(nil, errors.New("boom"))

				err := sync.SyncLTC(ltcTempFile.Name(), "amiga", config)
				Expect(err).To(MatchError("failed to open temp file: boom"))
			})
		})

		Context("when the file copy fails", func() {
			It("should return an error", func() {
				fakeServer.RouteToHandler("GET", "/v1/sync/amiga/ltc", ghttp.CombineHandlers(
					ghttp.RespondWith(200, []byte{0x01, 0x02, 0x03}, http.Header{
						"Content-Type":   []string{"application/octet-stream"},
						"Content-Length": []string{"3"},
					}),
				))

				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tmpFile.Name())
				tmpFile, err = os.OpenFile(tmpFile.Name(), os.O_RDONLY, 0)
				Expect(err).NotTo(HaveOccurred())

				fakeFileSwapper.GetTempFileReturns(tmpFile, nil)

				err = sync.SyncLTC(ltcTempFile.Name(), "amiga", config)
				Expect(err).To(MatchError(HavePrefix("failed to write to temp ltc")))
			})
		})

		Context("when swapping the files fails", func() {
			It("should return an error", func() {
				fakeServer.RouteToHandler("GET", "/v1/sync/amiga/ltc", ghttp.CombineHandlers(
					ghttp.RespondWith(200, []byte{0x01, 0x02, 0x03}, http.Header{
						"Content-Type":   []string{"application/octet-stream"},
						"Content-Length": []string{"3"},
					}),
				))

				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tmpFile.Name())

				fakeFileSwapper.GetTempFileReturns(tmpFile, nil)
				fakeFileSwapper.SwapTempFileReturns(errors.New("failed"))

				err = sync.SyncLTC(ltcTempFile.Name(), "amiga", config)
				Expect(err).To(MatchError(HavePrefix("failed to swap ltc")))
			})
		})
	})
})
