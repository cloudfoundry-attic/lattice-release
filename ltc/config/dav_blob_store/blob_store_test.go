package dav_blob_store_test

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
)

var _ = Describe("BlobStore", func() {
	var (
		blobStore  *dav_blob_store.BlobStore
		fakeServer *ghttp.Server
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		fakeServerURL, err := url.Parse(fakeServer.URL())
		Expect(err).NotTo(HaveOccurred())

		serverHost, serverPort, err := net.SplitHostPort(fakeServerURL.Host)
		Expect(err).NotTo(HaveOccurred())

		blobTargetInfo := dav_blob_store.Config{
			Host:     serverHost,
			Port:     serverPort,
			Username: "user",
			Password: "pass",
		}

		blobStore = dav_blob_store.New(blobTargetInfo)
	})

	AfterEach(func() {
		if fakeServer != nil {
			fakeServer.Close()
		}
	})

	Describe("#List", func() {
		var responseBodyRoot string
		BeforeEach(func() {
			responseBodyRoot = `
				<?xml version="1.0" encoding="utf-8"?>
				<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/b</D:href>
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
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/a</D:href>
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
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/c</D:href>
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

		It("lists objects", func() {
			responseBodyA := `
				<?xml version="1.0" encoding="utf-8"?>
				<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/a/bits.zip</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/zip-compressed</D:getcontenttype>
						<D:getlastmodified ns0:dt="dateTime.rfc1123">Wed, 29 Jul 2015 18:43:36 GMT</D:getlastmodified>
						<D:resourcetype>
						  <D:collection/>
						</D:resourcetype>
					  </D:prop>
					  <D:status>HTTP/1.1 200 OK</D:status>
					</D:propstat>
				  </D:response>
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/a/</D:href>
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
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/a/droplet.tgz</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/x-gtar-compressed</D:getcontenttype>
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

			responseBodyB := `
				<?xml version="1.0" encoding="utf-8"?>
				<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/b/bits.zip</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/zip-compressed</D:getcontenttype>
						<D:getlastmodified ns0:dt="dateTime.rfc1123">Wed, 29 Jul 2015 18:43:36 GMT</D:getlastmodified>
						<D:resourcetype>
						  <D:collection/>
						</D:resourcetype>
					  </D:prop>
					  <D:status>HTTP/1.1 200 OK</D:status>
					</D:propstat>
				  </D:response>
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/b/</D:href>
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
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/b/droplet.tgz</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/x-gtar-compressed</D:getcontenttype>
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

			responseBodyC := `
				<?xml version="1.0" encoding="utf-8"?>
				<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/c/bits.zip</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/zip-compressed</D:getcontenttype>
						<D:getlastmodified ns0:dt="dateTime.rfc1123">Wed, 29 Jul 2015 18:43:36 GMT</D:getlastmodified>
						<D:resourcetype>
						  <D:collection/>
						</D:resourcetype>
					  </D:prop>
					  <D:status>HTTP/1.1 200 OK</D:status>
					</D:propstat>
				  </D:response>
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/c/</D:href>
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
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/c/droplet.tgz</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/x-gtar-compressed</D:getcontenttype>
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

			responseBodyA = strings.Replace(responseBodyA, "http://192.168.11.11:8444", fakeServer.URL(), -1)
			responseBodyB = strings.Replace(responseBodyB, "http://192.168.11.11:8444", fakeServer.URL(), -1)
			responseBodyC = strings.Replace(responseBodyC, "http://192.168.11.11:8444", fakeServer.URL(), -1)

			fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
				ghttp.VerifyHeaderKV("Depth", "1"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			fakeServer.RouteToHandler("PROPFIND", "/blobs/a", ghttp.CombineHandlers(
				ghttp.VerifyHeaderKV("Depth", "1"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(207, responseBodyA, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			fakeServer.RouteToHandler("PROPFIND", "/blobs/b", ghttp.CombineHandlers(
				ghttp.VerifyHeaderKV("Depth", "1"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(207, responseBodyB, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			fakeServer.RouteToHandler("PROPFIND", "/blobs/c", ghttp.CombineHandlers(
				ghttp.VerifyHeaderKV("Depth", "1"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(207, responseBodyC, http.Header{"Content-Type": []string{"application/xml"}}),
			))

			expectedTime, err := time.Parse(time.RFC1123, "Wed, 29 Jul 2015 18:43:36 GMT")
			Expect(err).NotTo(HaveOccurred())

			Expect(blobStore.List()).To(ConsistOf(
				dav_blob_store.Blob{Path: "b/bits.zip", Size: 4096, Created: expectedTime},
				dav_blob_store.Blob{Path: "b/droplet.tgz", Size: 4096, Created: expectedTime},
				dav_blob_store.Blob{Path: "a/bits.zip", Size: 4096, Created: expectedTime},
				dav_blob_store.Blob{Path: "a/droplet.tgz", Size: 4096, Created: expectedTime},
				dav_blob_store.Blob{Path: "c/bits.zip", Size: 4096, Created: expectedTime},
				dav_blob_store.Blob{Path: "c/droplet.tgz", Size: 4096, Created: expectedTime},
			))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(4))
		})

		Context("when the root call fails", func() {
			It("returns an error it can't connect to the server", func() {
				fakeServer.Close()
				fakeServer = nil

				_, err := blobStore.List()
				Expect(err).To(HaveOccurred())
			})

			It("returns an error when we fail to retrieve the objects from DAV", func() {
				fakeServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("PROPFIND", "/blobs"),
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.VerifyBasicAuth("user", "pass"),
					ghttp.RespondWith(http.StatusInternalServerError, nil, http.Header{"Content-Type": []string{"application/xml"}}),
				))

				_, err := blobStore.List()
				Expect(err).To(MatchError(ContainSubstring("Internal Server Error")))

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error when it fails to parse the XML", func() {
				fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.VerifyBasicAuth("user", "pass"),
					ghttp.RespondWith(207, `<D:bad`, http.Header{"Content-Type": []string{"application/xml"}}),
				))

				_, err := blobStore.List()
				Expect(err).To(MatchError("XML syntax error on line 1: unexpected EOF"))

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when it fails to parse the time", func() {
				It("returns an error", func() {
					responseBodyRoot = strings.Replace(responseBodyRoot, "Wed, 29 Jul 2015 18:43:36 GMT", "ABC", -1)

					fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Depth", "1"),
						ghttp.VerifyBasicAuth("user", "pass"),
						ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
					))

					_, err := blobStore.List()
					Expect(err).To(MatchError(ContainSubstring(`cannot parse "ABC"`)))

					Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Context("when the child call fails", func() {
			BeforeEach(func() {
				fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.VerifyBasicAuth("user", "pass"),
					ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
				))
			})

			It("returns an error it can't connect to the server", func() {
				doneChan := make(chan struct{})

				go func() {
					defer GinkgoRecover()
					Eventually(doneChan).Should(Receive())
					fakeServer.Close()
					fakeServer = nil
				}()

				fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.VerifyBasicAuth("user", "pass"),
					ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
					func(_ http.ResponseWriter, _ *http.Request) {
						defer func() {
							doneChan <- struct{}{}
						}()
					},
				))

				_, err := blobStore.List()
				Expect(err).To(HaveOccurred())
			})

			It("returns an error when we fail to retrieve the objects from DAV", func() {
				fakeServer.RouteToHandler("PROPFIND", "/blobs/b", ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.VerifyBasicAuth("user", "pass"),
					ghttp.RespondWith(http.StatusInternalServerError, nil, http.Header{"Content-Type": []string{"application/xml"}}),
				))

				_, err := blobStore.List()
				Expect(err).To(MatchError(ContainSubstring("Internal Server Error")))

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(2))
			})

			It("returns an error when it fails to parse the XML", func() {
				fakeServer.RouteToHandler("PROPFIND", "/blobs/b", ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Depth", "1"),
					ghttp.VerifyBasicAuth("user", "pass"),
					ghttp.RespondWith(207, `<D:bad`, http.Header{"Content-Type": []string{"application/xml"}}),
				))

				_, err := blobStore.List()
				Expect(err).To(MatchError("XML syntax error on line 1: unexpected EOF"))

				Expect(fakeServer.ReceivedRequests()).To(HaveLen(2))
			})

			Context("when the root call points to a bad child url", func() {
				BeforeEach(func() {
					responseBodyRoot = strings.Replace(responseBodyRoot, fakeServer.URL(), "#%Z", -1)
					fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Depth", "1"),
						ghttp.VerifyBasicAuth("user", "pass"),
						ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
					))
				})
				It("returns an error", func() {
					_, err := blobStore.List()
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})

			Context("when the child call points to a bad url", func() {
				BeforeEach(func() {
					responseBodyB := `
						<?xml version="1.0" encoding="utf-8"?>
						<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
						  <D:response>
							<D:href>http://192.168.11.11:8444/blobs/b/bits.zip</D:href>
							<D:propstat>
							  <D:prop>
								<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
								<D:getcontentlanguage>en</D:getcontentlanguage>
								<D:getcontentlength>4096</D:getcontentlength>
								<D:getcontenttype>application/zip-compressed</D:getcontenttype>
								<D:getlastmodified ns0:dt="dateTime.rfc1123">Wed, 29 Jul 2015 18:43:36 GMT</D:getlastmodified>
								<D:resourcetype>
								  <D:collection/>
								</D:resourcetype>
							  </D:prop>
							  <D:status>HTTP/1.1 200 OK</D:status>
							</D:propstat>
						  </D:response>
						  <D:response>
							<D:href>http://192.168.11.11:8444/blobs/b/</D:href>
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
						  <D:response>
							<D:href>http://192.168.11.11:8444/blobs/b/droplet.tgz</D:href>
							<D:propstat>
							  <D:prop>
								<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
								<D:getcontentlanguage>en</D:getcontentlanguage>
								<D:getcontentlength>4096</D:getcontentlength>
								<D:getcontenttype>application/x-gtar-compressed</D:getcontenttype>
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

					responseBodyB = strings.Replace(responseBodyB, "http://192.168.11.11:8444", "#%Z", -1)

					fakeServer.RouteToHandler("PROPFIND", "/blobs", ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Depth", "1"),
						ghttp.VerifyBasicAuth("user", "pass"),
						ghttp.RespondWith(207, responseBodyRoot, http.Header{"Content-Type": []string{"application/xml"}}),
					))
					fakeServer.RouteToHandler("PROPFIND", "/blobs/b", ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Depth", "1"),
						ghttp.VerifyBasicAuth("user", "pass"),
						ghttp.RespondWith(207, responseBodyB, http.Header{"Content-Type": []string{"application/xml"}}),
					))
				})
				It("returns an error", func() {
					_, err := blobStore.List()
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})
		})
	})

	Describe("#Upload", func() {
		BeforeEach(func() {
			fakeServer.RouteToHandler("PROPFIND", "/blobs/some-path", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(404, "", http.Header{}),
			))
			fakeServer.RouteToHandler("MKCOL", "/blobs/some-path", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(http.StatusCreated, "", http.Header{}),
			))
		})

		It("uploads the provided reader into the bucket", func() {
			fakeServer.RouteToHandler("PUT", "/blobs/some-path/some-object", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				func(_ http.ResponseWriter, request *http.Request) {
					Expect(ioutil.ReadAll(request.Body)).To(Equal([]byte("some data")))
				},
				ghttp.RespondWith(http.StatusCreated, "", http.Header{}),
			))

			Expect(blobStore.Upload("some-path/some-object", strings.NewReader("some data"))).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(3))
		})

		It("uploads the provided reader into the bucket which already exists", func() {
			collectionResponseBody := `
				<?xml version="1.0" encoding="utf-8"?>
				<D:multistatus xmlns:D="DAV:" xmlns:ns0="urn:uuid:c2f41010-65b3-11d1-a29f-00aa00c14882/">
				  <D:response>
					<D:href>http://192.168.11.11:8444/blobs/some-path/xxx</D:href>
					<D:propstat>
					  <D:prop>
						<D:creationdate ns0:dt="dateTime.tz">2015-07-29T18:43:50Z</D:creationdate>
						<D:getcontentlanguage>en</D:getcontentlanguage>
						<D:getcontentlength>4096</D:getcontentlength>
						<D:getcontenttype>application/zip-compressed</D:getcontenttype>
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

			fakeServer.RouteToHandler("PROPFIND", "/blobs/some-path", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(207, collectionResponseBody, http.Header{}),
			))
			fakeServer.RouteToHandler("MKCOL", "/blobs/some-path", func(_ http.ResponseWriter, _ *http.Request) {
				Fail("MKCOL shouldn't be called if the collection already exists.")
			})
			fakeServer.RouteToHandler("PUT", "/blobs/some-path/some-object", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				func(_ http.ResponseWriter, request *http.Request) {
					Expect(ioutil.ReadAll(request.Body)).To(Equal([]byte("some data")))
				},
				ghttp.RespondWith(http.StatusCreated, "", http.Header{}),
			))

			Expect(blobStore.Upload("some-path/some-object", strings.NewReader("some data"))).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(2))
		})

		It("uploads the provided reader into the bucket if the remote path doesn't exist", func() {
			fakeServer.RouteToHandler("PUT", "/blobs/some-path/some-object", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				func(_ http.ResponseWriter, request *http.Request) {
					Expect(ioutil.ReadAll(request.Body)).To(Equal([]byte("some data")))
				},
				ghttp.RespondWith(http.StatusCreated, "", http.Header{}),
			))

			Expect(blobStore.Upload("some-path/some-object", strings.NewReader("some data"))).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(3))
		})

		It("uploads the provided reader into the bucket if the remote path already exists", func() {
			fakeServer.RouteToHandler("PUT", "/blobs/some-path/some-object", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("user", "pass"),
				func(_ http.ResponseWriter, request *http.Request) {
					Expect(ioutil.ReadAll(request.Body)).To(Equal([]byte("some data")))
				},
				ghttp.RespondWith(http.StatusOK, "", http.Header{}),
			))

			Expect(blobStore.Upload("some-path/some-object", strings.NewReader("some data"))).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(3))
		})

		It("returns an error when DAV fails to receive the object", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/blobs/some-path/some-object"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
			))

			err := blobStore.Upload("some-path/some-object", strings.NewReader("some data"))
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(3))
		})

		It("returns an error when the DAV client cannot connect", func() {
			fakeServer.Close()
			fakeServer = nil

			err := blobStore.Upload("some-path/some-object", strings.NewReader("some data"))
			Expect(err).To(MatchError(ContainSubstring("connection refused")))
		})

		It("returns an error when MKCOL is required but fails", func() {
			fakeServer.RouteToHandler("MKCOL", "/blobs/some-path",
				ghttp.RespondWith(http.StatusInternalServerError, "", nil))

			err := blobStore.Upload("some-path/some-object", strings.NewReader("some data"))
			Expect(err).To(MatchError("500 Internal Server Error"))

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("#Download", func() {
		It("dowloads the requested path", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/blobs/some-path/some-object"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(http.StatusOK, "some data", http.Header{"Content-length": []string{"9"}}),
			))

			pathReader, err := blobStore.Download("some-path/some-object")
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.ReadAll(pathReader)).To(Equal([]byte("some data")))
			Expect(pathReader.Close()).To(Succeed())

			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when DAV fails to retrieve the object", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/blobs/some-path/some-object"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
			))

			_, err := blobStore.Download("some-path/some-object")
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})

		It("returns an error when the DAV client cannot connect", func() {
			fakeServer.Close()
			fakeServer = nil

			_, err := blobStore.Download("some-path/some-object")
			Expect(err).To(MatchError(ContainSubstring("connection refused")))
		})
	})

	Describe("#Delete", func() {
		It("deletes the object at the provided path", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/blobs/some-path/some-object"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(http.StatusNoContent, ""),
			))
			Expect(blobStore.Delete("some-path/some-object")).NotTo(HaveOccurred())
			Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when DAV fails to delete the object", func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/blobs/some-path/some-object"),
				ghttp.VerifyBasicAuth("user", "pass"),
				ghttp.RespondWith(http.StatusInternalServerError, "", http.Header{}),
			))

			err := blobStore.Delete("some-path/some-object")
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})

		It("returns an error when the DAV client cannot connect", func() {
			fakeServer.Close()
			fakeServer = nil

			err := blobStore.Delete("some-path/some-object")
			Expect(err).To(MatchError(ContainSubstring("connection refused")))
		})
	})
})
