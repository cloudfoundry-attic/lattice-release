package docker_metadata_fetcher_test

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/cli/app_runner/docker_metadata_fetcher"
	"github.com/docker/docker/registry"
)

var _ = Describe("DockerSessionFactory", func() {
	Describe("MakeSession", func() {
		var registryHost string
		var dockerRegistryServer *ghttp.Server

		BeforeEach(func() {
			dockerRegistryServer = ghttp.NewServer()
		})

		AfterEach(func() {
			dockerRegistryServer.Close()
		})

		Describe("Happy Path", func() {
			BeforeEach(func() {
				parts, _ := url.Parse(dockerRegistryServer.URL())
				registryHost = parts.Host

				dockerRegistryServer.RouteToHandler("GET", "/v1/_ping", ghttp.VerifyRequest("GET", "/v1/_ping"))
				dockerRegistryServer.RouteToHandler("GET", "/v2/", ghttp.VerifyRequest("GET", "/v2/"))
			})

			It("creates a registry session for the given repo", func() {
				sessionFactory := docker_metadata_fetcher.NewDockerSessionFactory()
				session, err := sessionFactory.MakeSession(registryHost + "/lattice-mappppppppppppappapapa")

				Expect(err).ToNot(HaveOccurred())
				registrySession, ok := session.(*registry.Session)
				Expect(ok).To(Equal(true))

				Expect(*registrySession.GetAuthConfig(true)).To(Equal(registry.AuthConfig{}))

			})
		})

		Context("When resolving the repo name fails", func() {
			It("returns errors from resolving the repo name", func() {
				sessionFactory := docker_metadata_fetcher.NewDockerSessionFactory()
				_, err := sessionFactory.MakeSession("¥Not-A-Valid-Repo-Name¥" + "/lattice-mappppppppppppappapapa")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("Error resolving Docker repository name:\nInvalid namespace name"))

			})
		})

		Context("when creating a new endpoint fails", func() {
			It("returns an error", func() {
				sessionFactory := docker_metadata_fetcher.NewDockerSessionFactory()
				_, err := sessionFactory.MakeSession("nonexistantregistry.example.com/lattice-mappppppppppppappapapa")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("Error Connecting to Docker registry:\ninvalid registry endpoint"))
			})
		})

	})
})
