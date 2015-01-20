package docker_metadata_fetcher_test

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/docker/docker/registry"
	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_metadata_fetcher"
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
				Expect(err.Error()).To(Equal("Error resolving Docker repository name:\nInvalid namespace name (¥Not-A-Valid-Repo-Name¥), only [a-z0-9_] are allowed, size between 4 and 30"))

			})
		})

		Context("when creating a new endpoint fails", func() {
			It("returns an error", func() {
				sessionFactory := docker_metadata_fetcher.NewDockerSessionFactory()
				_, err := sessionFactory.MakeSession(registryHost + "/lattice-mappppppppppppappapapa")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("Error Connecting to Docker registry:\nInvalid registry endpoint"))
			})
		})

	})
})
