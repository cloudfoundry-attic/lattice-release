package docker_metadata_fetcher_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher/fake_docker_session"
	"github.com/docker/docker/registry"
)

var _ = Describe("DockerMetaDataFetcher", func() {
	var (
		dockerMetadataFetcher docker_metadata_fetcher.DockerMetadataFetcher
		dockerSessionFactory  *fake_docker_session.FakeDockerSessionFactory
		fakeDockerSession     *fake_docker_session.FakeDockerSession
	)

	BeforeEach(func() {
		fakeDockerSession = &fake_docker_session.FakeDockerSession{}
		dockerSessionFactory = &fake_docker_session.FakeDockerSessionFactory{}
		dockerMetadataFetcher = docker_metadata_fetcher.New(dockerSessionFactory)
	})

	Describe("FetchMetadata", func() {

		Context("when fetching metadata from the docker hub registry", func() {
			It("returns the ImageMetadata with the WorkingDir, StartCommand, and PortConfig, and sets the monitored port to the lowest exposed tcp port", func() {
				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)

				imageList := map[string]*registry.ImgData{
					"29d531509fb": &registry.ImgData{
						ID:              "29d531509fb",
						Checksum:        "dsflksdfjlkj",
						ChecksumPayload: "sdflksdjfkl",
						Tag:             "latest",
					},
				}
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   imageList,
						Endpoints: []string{"https://registry-1.docker.io/v1/"},
						Tokens:    []string{"signature=abc,repository=\"cloudfoundry/lattice-app\",access=read"},
					}, nil)

				fakeDockerSession.GetRemoteTagsReturns(map[string]string{"latest": "29d531509fb"}, nil)

				fakeDockerSession.GetRemoteImageJSONReturns(
					[]byte(`{
					"container_config":{ "ExposedPorts":{"28321/tcp":{}, "6923/udp":{}, "27017/tcp":{}} },
				 	"config":{
				 				"WorkingDir":"/home/app",
				 				"Entrypoint":["/lattice-app"],
				 				"Cmd":["--enableAwesomeMode=true","iloveargs"]
							}
						}`),
					0,
					nil)

				dockerImageReference := "cool_user123/sweetapp:latest"
				dockerImageNoTag := "cool_user123/sweetapp"
				imageMetadata, err := dockerMetadataFetcher.FetchMetadata(dockerImageReference)
				Expect(err).ToNot(HaveOccurred())

				Expect(dockerSessionFactory.MakeSessionCallCount()).To(Equal(1))
				Expect(dockerSessionFactory.MakeSessionArgsForCall(0)).To(Equal(dockerImageNoTag))

				Expect(fakeDockerSession.GetRepositoryDataCallCount()).To(Equal(1))
				Expect(fakeDockerSession.GetRepositoryDataArgsForCall(0)).To(Equal(dockerImageNoTag))

				Expect(fakeDockerSession.GetRemoteTagsCallCount()).To(Equal(1))
				registries, repo, tokens := fakeDockerSession.GetRemoteTagsArgsForCall(0)
				Expect(registries).To(Equal([]string{"https://registry-1.docker.io/v1/"}))
				Expect(repo).To(Equal("cool_user123/sweetapp"))
				Expect(tokens).To(Equal([]string{"signature=abc,repository=\"cloudfoundry/lattice-app\",access=read"}))

				Expect(fakeDockerSession.GetRemoteImageJSONCallCount()).To(Equal(1))
				imgIDParam, remoteImageEndpointParam, remoteImageTokensParam := fakeDockerSession.GetRemoteImageJSONArgsForCall(0)
				Expect(imgIDParam).To(Equal("29d531509fb"))
				Expect(remoteImageEndpointParam).To(Equal("https://registry-1.docker.io/v1/"))
				Expect(remoteImageTokensParam).To(Equal([]string{"signature=abc,repository=\"cloudfoundry/lattice-app\",access=read"}))

				Expect(imageMetadata.WorkingDir).To(Equal("/home/app"))
				Expect(imageMetadata.StartCommand).To(Equal([]string{"/lattice-app", "--enableAwesomeMode=true", "iloveargs"}))
				Expect(imageMetadata.Ports.Monitored).To(Equal(uint16(27017)))
				Expect(imageMetadata.Ports.Exposed).To(Equal([]uint16{uint16(27017), uint16(28321)}))
			})
		})

		Context("when fetching metadata from a signed custom registry", func() {
			It("returns the image metadata", func() {

				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)

				imageList := map[string]*registry.ImgData{
					"29d531509fb": &registry.ImgData{
						ID:              "29d531509fb",
						Checksum:        "dsflksdfjlkj",
						ChecksumPayload: "sdflksdjfkl",
						Tag:             "latest",
					},
				}
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   imageList,
						Endpoints: []string{"http://my.custom.registry:5000/v1/"},
						Tokens:    []string{"signature=abc,repository=\"library/savory-app\",access=read"},
					}, nil)

				fakeDockerSession.GetRemoteTagsReturns(map[string]string{"latest": "29d531509fb"}, nil)

				fakeDockerSession.GetRemoteImageJSONReturns(
					[]byte(`{
					"container_config":{ "ExposedPorts":{"4444/tcp":{}, "5555/udp":{}, "3333/tcp":{}} },
				 	"config":{
				 				"WorkingDir":"/home/app",
				 				"Entrypoint":["/savory-app"],
				 				"Cmd":["--pretzels=salty","cheesy"]
							}
						}`),
					0,
					nil)

				dockerImageReference := "my.custom.registry:5000/savory-app"

				imageMetadata, err := dockerMetadataFetcher.FetchMetadata(dockerImageReference)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageMetadata).ToNot(BeNil())

				Expect(dockerSessionFactory.MakeSessionCallCount()).To(Equal(1))
				Expect(dockerSessionFactory.MakeSessionArgsForCall(0)).To(Equal(dockerImageReference))

				Expect(fakeDockerSession.GetRepositoryDataCallCount()).To(Equal(1))
				Expect(fakeDockerSession.GetRepositoryDataArgsForCall(0)).To(Equal("savory-app"))

				Expect(fakeDockerSession.GetRemoteTagsCallCount()).To(Equal(1))
				registries, repo, tokens := fakeDockerSession.GetRemoteTagsArgsForCall(0)
				Expect(registries).To(Equal([]string{"http://my.custom.registry:5000/v1/"}))
				Expect(repo).To(Equal("savory-app"))
				Expect(tokens).To(Equal([]string{"signature=abc,repository=\"library/savory-app\",access=read"}))

				Expect(fakeDockerSession.GetRemoteImageJSONCallCount()).To(Equal(1))
				imgIDParam, remoteImageEndpointParam, remoteImageTokensParam := fakeDockerSession.GetRemoteImageJSONArgsForCall(0)
				Expect(imgIDParam).To(Equal("29d531509fb"))
				Expect(remoteImageEndpointParam).To(Equal("http://my.custom.registry:5000/v1/"))
				Expect(remoteImageTokensParam).To(Equal([]string{"signature=abc,repository=\"library/savory-app\",access=read"}))

				Expect(imageMetadata.WorkingDir).To(Equal("/home/app"))
				Expect(imageMetadata.StartCommand).To(Equal([]string{"/savory-app", "--pretzels=salty", "cheesy"}))
				Expect(imageMetadata.Ports.Monitored).To(Equal(uint16(3333)))
				Expect(imageMetadata.Ports.Exposed).To(Equal([]uint16{uint16(3333), uint16(4444)}))
			})
		})

		Context("when fetching metadata from a insecure custom registry", func() {
			It("retries after getting unknown CA error and returns the image metadata", func() {
				insecureRegistryErrorMessage := "If this private registry supports only HTTP or HTTPS with an unknown CA certificate, please add `--insecure-registry 192.168.11.1:5000` to the daemon's arguments. In the case of HTTPS, if you have access to the registry's CA certificate, no need for the flag; simply place the CA certificate at /etc/docker/certs.d/192.168.11.1:5000/ca.crt"

				dockerSessionFactory.MakeSessionStub = func(reposName string, allowInsecure bool) (docker_metadata_fetcher.DockerSession, error) {
					if !allowInsecure {
						return fakeDockerSession, errors.New(insecureRegistryErrorMessage)
					}

					return fakeDockerSession, nil
				}

				imageList := map[string]*registry.ImgData{
					"29d531509fb": &registry.ImgData{
						ID:              "29d531509fb",
						Checksum:        "dsflksdfjlkj",
						ChecksumPayload: "sdflksdjfkl",
						Tag:             "latest",
					},
				}
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   imageList,
						Endpoints: []string{"http://my.custom.registry:5000/v1/"},
						Tokens:    []string{"signature=abc,repository=\"library/savory-app\",access=read"},
					}, nil)

				fakeDockerSession.GetRemoteTagsReturns(map[string]string{"latest": "29d531509fb"}, nil)

				fakeDockerSession.GetRemoteImageJSONReturns(
					[]byte(`{
					"container_config":{ "ExposedPorts":{"4444/tcp":{}, "5555/udp":{}, "3333/tcp":{}} },
				 	"config":{
				 				"WorkingDir":"/home/app",
				 				"Entrypoint":["/savory-app"],
				 				"Cmd":["--pretzels=salty","cheesy"]
							}
						}`),
					0,
					nil)

				dockerImageReference := "my.custom.registry:5000/savory-app"

				imageMetadata, err := dockerMetadataFetcher.FetchMetadata(dockerImageReference)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageMetadata).ToNot(BeNil())

				Expect(dockerSessionFactory.MakeSessionCallCount()).To(Equal(2))

				reposName, allowInsecure := dockerSessionFactory.MakeSessionArgsForCall(0)
				Expect(reposName).To(Equal(dockerImageReference))
				Expect(allowInsecure).To(BeFalse())

				reposName, allowInsecure = dockerSessionFactory.MakeSessionArgsForCall(1)
				Expect(reposName).To(Equal(dockerImageReference))
				Expect(allowInsecure).To(BeTrue())

				Expect(fakeDockerSession.GetRepositoryDataCallCount()).To(Equal(1))
				Expect(fakeDockerSession.GetRepositoryDataArgsForCall(0)).To(Equal("savory-app"))

				Expect(fakeDockerSession.GetRemoteTagsCallCount()).To(Equal(1))
				registries, repo, tokens := fakeDockerSession.GetRemoteTagsArgsForCall(0)
				Expect(registries).To(Equal([]string{"http://my.custom.registry:5000/v1/"}))
				Expect(repo).To(Equal("savory-app"))
				Expect(tokens).To(Equal([]string{"signature=abc,repository=\"library/savory-app\",access=read"}))

				Expect(fakeDockerSession.GetRemoteImageJSONCallCount()).To(Equal(1))
				imgIDParam, remoteImageEndpointParam, remoteImageTokensParam := fakeDockerSession.GetRemoteImageJSONArgsForCall(0)
				Expect(imgIDParam).To(Equal("29d531509fb"))
				Expect(remoteImageEndpointParam).To(Equal("http://my.custom.registry:5000/v1/"))
				Expect(remoteImageTokensParam).To(Equal([]string{"signature=abc,repository=\"library/savory-app\",access=read"}))

				Expect(imageMetadata.WorkingDir).To(Equal("/home/app"))
				Expect(imageMetadata.StartCommand).To(Equal([]string{"/savory-app", "--pretzels=salty", "cheesy"}))
				Expect(imageMetadata.Ports.Monitored).To(Equal(uint16(3333)))
				Expect(imageMetadata.Ports.Exposed).To(Equal([]uint16{uint16(3333), uint16(4444)}))
			})

			Context("when getting another error after retrying", func() {
				It("returns the error", func() {
					insecureRegistryErrorMessage := "If this private registry supports only HTTP or HTTPS with an unknown CA certificate, please add `--insecure-registry 192.168.11.1:5000` to the daemon's arguments. In the case of HTTPS, if you have access to the registry's CA certificate, no need for the flag; simply place the CA certificate at /etc/docker/certs.d/192.168.11.1:5000/ca.crt"
					dockerImageReference := "verybad/apple"

					dockerSessionFactory.MakeSessionReturns(fakeDockerSession, errors.New(insecureRegistryErrorMessage))
					_, err := dockerMetadataFetcher.FetchMetadata(dockerImageReference)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(MatchRegexp("private registry supports only HTTP or HTTPS with an unknown CA certificate"))

					Expect(dockerSessionFactory.MakeSessionCallCount()).To(Equal(2))

					reposName, allowInsecure := dockerSessionFactory.MakeSessionArgsForCall(0)
					Expect(reposName).To(Equal(dockerImageReference))
					Expect(allowInsecure).To(BeFalse())

					reposName, allowInsecure = dockerSessionFactory.MakeSessionArgsForCall(1)
					Expect(reposName).To(Equal(dockerImageReference))
					Expect(allowInsecure).To(BeTrue())
				})
			})
		})

		Context("when exposed ports are null in the docker metadata", func() {
			It("doesn't blow up, and returns zero values", func() {

				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)

				imageList := map[string]*registry.ImgData{
					"29d531509fb": &registry.ImgData{
						ID:              "29d531509fb",
						Checksum:        "dsflksdfjlkj",
						ChecksumPayload: "sdflksdjfkl",
						Tag:             "latest",
					},
				}
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   imageList,
						Endpoints: []string{"https://registry-1.docker.io/v1/"},
						Tokens:    []string{"signature=abc,repository=\"cloudfoundry/lattice-app\",access=read"},
					}, nil)

				fakeDockerSession.GetRemoteTagsReturns(map[string]string{"latest": "29d531509fb"}, nil)

				fakeDockerSession.GetRemoteImageJSONReturns(
					[]byte(`{
					"container_config":{ "ExposedPorts":null },
				 	"config":{
				 				"WorkingDir":"/home/app",
				 				"Entrypoint":["/lattice-app"],
				 				"Cmd":["--enableAwesomeMode=true","iloveargs"]
							}
						}`),
					0,
					nil)

				repoName := "cool_user123/sweetapp"
				imageMetadata, err := dockerMetadataFetcher.FetchMetadata(repoName)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageMetadata.Ports.Monitored).To(Equal(uint16(0)))
				Expect(imageMetadata.Ports.Exposed).To(Equal([]uint16{}))

			})
		})

		Context("when there is an error parsing the docker image reference", func() {
			It("returns an error", func() {
				_, err := dockerMetadataFetcher.FetchMetadata("bad/appName")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid namespace name (bad). Cannot be fewer than 4 or more than 30 characters."))
			})
		})

		Context("when there is an error making the session", func() {
			It("returns an error", func() {
				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, errors.New("Couldn't make a session."))
				_, err := dockerMetadataFetcher.FetchMetadata("verybad/apple")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Couldn't make a session."))
			})
		})

		Context("when there is an error getting the repository data", func() {
			It("returns an error", func() {
				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)
				fakeDockerSession.GetRepositoryDataReturns(&registry.RepositoryData{}, errors.New("We floundered getting your repo data."))

				_, err := dockerMetadataFetcher.FetchMetadata("cloud_flounder/fishy")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("We floundered getting your repo data."))
			})
		})

		Context("when there is an error getting remote tags", func() {
			It("returns an error", func() {
				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   map[string]*registry.ImgData{},
						Endpoints: []string{},
						Tokens:    []string{},
					}, nil)

				fakeDockerSession.GetRemoteTagsReturns(nil, errors.New("Can't get tags!"))

				_, err := dockerMetadataFetcher.FetchMetadata("tagless/inseattle")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Can't get tags!"))
			})
		})

		Context("When the requested tag does not exist", func() {
			It("returns an error", func() {
				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)

				imageList := map[string]*registry.ImgData{
					"29d531509fb": &registry.ImgData{
						ID:              "29d531509fb",
						Checksum:        "dsflksdfjlkj",
						ChecksumPayload: "sdflksdjfkl",
						Tag:             "latest",
					},
				}
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   imageList,
						Endpoints: []string{"https://registry-1.docker.io/v1/"},
						Tokens:    []string{"signature=abc,repository=\"cloudfoundry/lattice-app\",access=read"},
					}, nil)

				fakeDockerSession.GetRemoteTagsReturns(map[string]string{"latest": "29d531509fb"}, nil)

				_, err := dockerMetadataFetcher.FetchMetadata("wiggle/app:some-unknown-tag-v3245")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Unknown tag: wiggle/app:some-unknown-tag-v3245"))
			})
		})

		Describe("Handling image JSON errors", func() {
			BeforeEach(func() {
				dockerSessionFactory.MakeSessionReturns(fakeDockerSession, nil)

				imageList := map[string]*registry.ImgData{
					"29d531509fb": &registry.ImgData{
						ID:              "29d531509fb",
						Checksum:        "dsflksdfjlkj",
						ChecksumPayload: "sdflksdjfkl",
						Tag:             "latest",
					},
				}
				fakeDockerSession.GetRepositoryDataReturns(
					&registry.RepositoryData{
						ImgList:   imageList,
						Endpoints: []string{"https://registry-1.docker.io/v1/"},
						Tokens:    []string{"signature=abc,repository=\"cloudfoundry/lattice-app\",access=read"},
					}, nil)
				fakeDockerSession.GetRemoteTagsReturns(map[string]string{"latest": "29d531509fb"}, nil)
			})

			Context("when there is an error getting the remote image json", func() {
				It("returns an error", func() {
					fakeDockerSession.GetRemoteImageJSONReturns([]byte{}, 0, errors.New("JSON? What's that!???"))

					_, err := dockerMetadataFetcher.FetchMetadata("wiggle/app")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("JSON? What's that!???"))
				})
			})

			Context("when there is an error parsing the remote image json", func() {
				It("returns an error", func() {
					fakeDockerSession.GetRemoteImageJSONReturns([]byte("i'm not valid json"), 0, nil)

					_, err := dockerMetadataFetcher.FetchMetadata("wiggle/app")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Error parsing remote image json for specified docker image:\ninvalid character 'i' looking for beginning of value"))
				})
			})

			Context("When Config is missing from the image Json", func() {
				It("returns an error", func() {
					fakeDockerSession.GetRemoteImageJSONReturns([]byte("{}"), 0, nil)

					_, err := dockerMetadataFetcher.FetchMetadata("wiggle/app")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Parsing start command failed"))
				})
			})
		})
	})
})
