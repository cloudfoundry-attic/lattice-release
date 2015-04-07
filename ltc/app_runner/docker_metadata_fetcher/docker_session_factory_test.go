package docker_metadata_fetcher_test

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/docker/docker/registry"
)

var _ = Describe("DockerSessionFactory", func() {
	Describe("MakeSession", func() {
		var (
			registryHost         string
			dockerRegistryServer *ghttp.Server
			sessionFactory       docker_metadata_fetcher.DockerSessionFactory
		)

		BeforeEach(func() {
			sessionFactory = docker_metadata_fetcher.NewDockerSessionFactory()

			dockerRegistryServer = ghttp.NewServer()
		})

		AfterEach(func() {
			dockerRegistryServer.Close()
		})

		Describe("creating registry sessions", func() {
			BeforeEach(func() {
				parts, _ := url.Parse(dockerRegistryServer.URL())
				registryHost = parts.Host

				dockerRegistryServer.RouteToHandler("GET", "/v1/_ping", ghttp.VerifyRequest("GET", "/v1/_ping"))
				dockerRegistryServer.RouteToHandler("GET", "/v2/", ghttp.VerifyRequest("GET", "/v2/"))
			})

			Context("when connecting to a secure registry", func() {
				It("creates a registry session for the given repo", func() {
					session, err := sessionFactory.MakeSession(registryHost+"/lattice-mappppppppppppappapapa", false)
					Expect(err).ToNot(HaveOccurred())

					registrySession, ok := session.(*registry.Session)
					Expect(ok).To(BeTrue())

					Expect(*registrySession.GetAuthConfig(true)).To(Equal(registry.AuthConfig{}))
				})
			})

			Context("when connecting to an insecure registry", func() {
				It("creates a registry session for the given repo", func() {
					session, err := sessionFactory.MakeSession(registryHost+"/lattice-mappppppppppppappapapa", true)
					Expect(err).ToNot(HaveOccurred())

					registrySession, ok := session.(*registry.Session)
					Expect(ok).To(BeTrue())

					Expect(*registrySession.GetAuthConfig(true)).To(Equal(registry.AuthConfig{}))
				})
			})
		})

		Context("when resolving the repo name fails", func() {
			It("returns errors from resolving the repo name", func() {
				_, err := sessionFactory.MakeSession("¥Not-A-Valid-Repo-Name¥"+"/lattice-mappppppppppppappapapa", false)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("Error resolving Docker repository name:\nInvalid namespace name"))
			})
		})

		Context("when creating a new endpoint fails", func() {
			It("returns an error", func() {
				_, err := sessionFactory.MakeSession("nonexistantregistry.example.com/lattice-mappppppppppppappapapa", false)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("Error Connecting to Docker registry:\ninvalid registry endpoint"))
			})
		})

	})
})
