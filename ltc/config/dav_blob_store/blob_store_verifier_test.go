package dav_blob_store_test

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
)

var _ = Describe("BlobStore", func() {
	var (
		verifier               dav_blob_store.Verifier
		fakeServer             *ghttp.Server
		serverHost, serverPort string
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()

		fakeServerURL, err := url.Parse(fakeServer.URL())
		Expect(err).NotTo(HaveOccurred())

		serverHost, serverPort, err = net.SplitHostPort(fakeServerURL.Host)
		Expect(err).NotTo(HaveOccurred())

		verifier = dav_blob_store.Verifier{}
	})

	AfterEach(func() {
		if fakeServer != nil {
			fakeServer.Close()
		}
	})

	Describe("Verify", func() {
		var responseBodyRoot string
		BeforeEach(func() {
			responseBodyRoot = `
				<?xml version="1.0" encoding="utf-8"?>
				<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>httpd/unix-directory</D:getcontenttype>
						<D:getlastmodified ns0:dt="dateTime.rfc1123">Wed, 29 Jul 2015 18:43:36 GMT</D:getlastmodified>
						<D:resourcetype>
						  <D:collection/>
						</D:resourcetype>
					  </D:prop>
					  <D:status>HTTP/1.1 200 OK</D:status>
					</D:propstat>
				  </D:response>
				</D:multistatus>
			`

			responseBodyRoot = strings.Replace(responseBodyRoot, "http://192.168.11.11:8444", fakeServer.URL(), -1)
		})

		It("verifies a blob store with valid credentials", func() {
			blobTargetInfo := dav_blob_store.Config{
				Host:     serverHost,
				Port:     serverPort,
				Username: "",
				Password: "",
			}

			fakeServer.RouteToHandler("PROPFIND", "/blobs/", ghttp.CombineHandlers(
				ghttp.VerifyHeaderKV("Depth", "1"),
				ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			authorized, err := verifier.Verify(blobTargetInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(authorized).To(BeTrue())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when the blob store credentials are incorrect", func() {
			It("rejects a blob store with invalid credentials", func() {
				blobTargetInfo := dav_blob_store.Config{
					Host:     serverHost,
					Port:     serverPort,
					Username: "",
					Password: "",
				}

				responseBody := `<?xml version="1.0" encoding="iso-8859-1"?>
				<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
				         "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
				<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
				 <head>
				  <title>401 - Unauthorized</title>
				 </head>
				 <body>
				  <h1>401 - Unauthorized</h1>
				 </body>
				</html>`

				fakeServer.RouteToHandler("PROPFIND", "/blobs/", ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.RespondWith(401, responseBody, http.Header{"Content-Type": []string{"application/xml"}}),
				))

				authorized, err := verifier.Verify(blobTargetInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(authorized).To(BeFalse())

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the blob store is inaccessible", func() {
			It("returns an error", func() {
				blobTargetInfo := dav_blob_store.Config{
					Host:     serverHost,
					Port:     serverPort,
					Username: "",
					Password: "",
				}

				fakeServer.Close()
				fakeServer = nil

				_, err := verifier.Verify(blobTargetInfo)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
