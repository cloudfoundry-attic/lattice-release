package docker_metadata_fetcher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
	"net/http"
	"net/url"
)

var _ = Describe("DockerMetaDataFetcher", func() {
	Describe("FetchMetadata", func() {
		var registryHost string
		var invalidHost string

		BeforeEach(func() {
			server := ghttp.NewServer()
			invalidServer := ghttp.NewServer()

			parts, _ := url.Parse(server.URL())
			registryHost = parts.Host

			parts, _ = url.Parse(invalidServer.URL())
			invalidHost = parts.Host

			server.RouteToHandler("GET", "/v1/_ping", ghttp.VerifyRequest("GET", "/v1/_ping"))

			server.RouteToHandler("GET", "/v1/repositories/ouruser/ourrepo/images", ghttp.CombineHandlers(
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("X-Docker-Token", "token-1,token-2")
					w.Write([]byte(`[
			                            {"id": "id-1", "checksum": "sha-1"},
			                            {"id": "id-2", "checksum": "sha-2"},
			                            {"id": "id-3", "checksum": "sha-3"}
			                        ]`))
				}),
			))

			server.RouteToHandler("GET", "/v1/repositories/ouruser/ourrepo/tags", ghttp.CombineHandlers(
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte(`{
								   "latest": "id-1",
								   "some-tag": "id-2"
							   }`))
				}),
			))

			server.RouteToHandler("GET", "/v1/images/id-1/json", ghttp.CombineHandlers(
				ghttp.CombineHandlers(
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("X-Docker-Size", "789")
						w.Write([]byte(`{"id":"layer-1","parent":"parent-1","Config":{"Entrypoint": ["/dockerapp", "entryarg"], "Cmd":["-foobar", "bazbot"], "WorkingDir": "/workingdir"}}`))
					}),
				),
			))

		})

		It("returns the ImageMetadata with the WorkingDir and StartCommand", func() {
			dockerMetaDataFetcher := docker_metadata_fetcher.New()

			repoName := registryHost + "/ouruser/ourrepo"
			imageMetaData, err := dockerMetaDataFetcher.FetchMetadata(repoName, "latest")

			Expect(err).ToNot(HaveOccurred())
			Expect(imageMetaData.WorkingDir).To(Equal("/workingdir"))
			Expect(imageMetaData.StartCommand).To(Equal([]string{"/dockerapp", "entryarg", "-foobar", "bazbot"}))
		})

		Context("with invalid arguemnts", func() {
			It("handles missing tags", func() {
				dockerMetaDataFetcher := docker_metadata_fetcher.New()

				repoName := registryHost + "/ouruser/ourrepo"
				_, err := dockerMetaDataFetcher.FetchMetadata(repoName, "nonexistanttag")

				Expect(err).To(HaveOccurred())
			})

			It("handles invalid repo names", func() {
				dockerMetaDataFetcher := docker_metadata_fetcher.New()

				_, err := dockerMetaDataFetcher.FetchMetadata("INVALID REPO NAME", "latest")

				Expect(err).To(HaveOccurred())
			})

			It("handles invalid hosts", func() {
				dockerMetaDataFetcher := docker_metadata_fetcher.New()

				repoName := "qwer:5123/ouruser/ourrepo"
				_, err := dockerMetaDataFetcher.FetchMetadata(repoName, "latest")

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
