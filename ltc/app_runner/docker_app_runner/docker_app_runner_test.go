package docker_app_runner_test

import (
	"encoding/json"
	"errors"
	"time"

	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("DockerAppRunner", func() {

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		appRunner          docker_app_runner.AppRunner
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		appRunner = docker_app_runner.New(fakeReceptorClient, "myDiegoInstall.com")

	})

	Describe("PortConfig", func() {
		Describe("IsEmpty", func() {
			It("returns true if the port config has no exposed ports", func() {
				portConfig := docker_app_runner.PortConfig{
					Monitored: uint16(0),
					Exposed:   []uint16{},
				}
				Expect(portConfig.IsEmpty()).To(BeTrue())

			})

			It("returns false if the port config has exposed ports", func() {
				portConfig := docker_app_runner.PortConfig{
					Monitored: uint16(0),
					Exposed:   []uint16{uint16(1234)},
				}
				Expect(portConfig.IsEmpty()).To(BeFalse())
			})
		})
	})

	Describe("CreateDockerApp", func() {
		It("Upserts lattice domain so that it is always fresh, then starts the Docker App", func() {
			fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)

			args := []string{"app", "arg1", "--app", "arg 2"}
			envs := map[string]string{"APPROOT": "/root/env/path"}
			err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
				Name:                 "americano-app",
				StartCommand:         "/app-run-statement",
				DockerImagePath:      "runtest/runner",
				AppArgs:              args,
				EnvironmentVariables: envs,
				Privileged:           true,
				Monitor:              true,
				Instances:            22,
				CPUWeight:            67,
				MemoryMB:             128,
				DiskMB:               1024,
				Ports:                docker_app_runner.PortConfig{Exposed: []uint16{2000, 4000}, Monitored: 2000},
				WorkingDir:           "/user/web/myappdir",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
			domain, ttl := fakeReceptorClient.UpsertDomainArgsForCall(0)
			Expect(domain).To(Equal("lattice"))
			Expect(ttl).To(Equal(time.Duration(0)))

			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0)).To(Equal(receptor.DesiredLRPCreateRequest{
				ProcessGuid:          "americano-app",
				Domain:               "lattice",
				RootFS:               "docker:///runtest/runner#latest",
				Instances:            22,
				EnvironmentVariables: []receptor.EnvironmentVariable{receptor.EnvironmentVariable{Name: "APPROOT", Value: "/root/env/path"}, receptor.EnvironmentVariable{Name: "PORT", Value: "2000"}},
				Routes: route_helpers.AppRoutes{
					route_helpers.AppRoute{Hostnames: []string{"americano-app.myDiegoInstall.com", "americano-app-2000.myDiegoInstall.com"}, Port: 2000},
					route_helpers.AppRoute{Hostnames: []string{"americano-app-4000.myDiegoInstall.com"}, Port: 4000},
				}.RoutingInfo(),
				CPUWeight:  67,
				MemoryMB:   128,
				DiskMB:     1024,
				Privileged: true,
				Ports:      []uint16{2000, 4000},
				LogGuid:    "americano-app",
				LogSource:  "APP",
				Setup: &models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
					To:   "/tmp",
				},
				Action: &models.RunAction{
					Path:       "/app-run-statement",
					Args:       []string{"app", "arg1", "--app", "arg 2"},
					Privileged: true,
					Dir:        "/user/web/myappdir",
				},
				Monitor: &models.RunAction{
					Path:      "/tmp/healthcheck",
					Args:      []string{"-port", "2000"},
					LogSource: "HEALTH",
				},
			}))
		})

		Context("when 'lattice-debug' is passed as the appId", func() {
			It("is an error because that id is reserved for the lattice-debug log stream", func() {
				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:            reserved_app_ids.LatticeDebugLogStreamAppId,
					StartCommand:    "/app-run-statement",
					DockerImagePath: "runtest/runner",
					AppArgs:         []string{},
				})

				Expect(err.Error()).To(Equal(docker_app_runner.AttemptedToCreateLatticeDebugErrorMessage))
			})
		})

		Context("when overrideRoutes is not empty", func() {
			It("uses the override Routes instead of the defaults", func() {
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)

				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:            "americano-app",
					StartCommand:    "/app-run-statement",
					DockerImagePath: "runtest/runner",
					AppArgs:         []string{},
					Ports:           docker_app_runner.PortConfig{Exposed: []uint16{2000, 3000, 4000}, Monitored: 2000},
					RouteOverrides: docker_app_runner.RouteOverrides{
						docker_app_runner.RouteOverride{HostnamePrefix: "wiggle", Port: 2000},
						docker_app_runner.RouteOverride{HostnamePrefix: "swang", Port: 2000},
						docker_app_runner.RouteOverride{HostnamePrefix: "shuffle", Port: 4000},
					},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				Expect(route_helpers.AppRoutesFromRoutingInfo(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Routes)).To(ContainExactly(route_helpers.AppRoutes{
					route_helpers.AppRoute{Hostnames: []string{"wiggle.myDiegoInstall.com", "swang.myDiegoInstall.com"}, Port: 2000},
					route_helpers.AppRoute{Hostnames: []string{"shuffle.myDiegoInstall.com"}, Port: 4000},
				}))
			})
		})

		Context("when Monitor is false", func() {
			It("Does not pass a monitor action, regardless of whether or not a monitor port is passed", func() {
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)

				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:            "americano-app",
					StartCommand:    "/app-run-statement",
					DockerImagePath: "runtest/runner",
					AppArgs:         []string{},
					Monitor:         false,
					Ports:           docker_app_runner.PortConfig{Monitored: 1234, Exposed: []uint16{1234}},
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Monitor).To(BeExactlyNil())
			})
		})

		It("returns errors if the app is already desired", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "app-already-desired", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
				Name:                 "app-already-desired",
				StartCommand:         "faily/boom",
				DockerImagePath:      "/app-bork-statement",
				AppArgs:              []string{},
				EnvironmentVariables: map[string]string{},
				Privileged:           false,
				Instances:            1,
				MemoryMB:             128,
				DiskMB:               1024,
				Ports:                docker_app_runner.PortConfig{Monitored: 8080, Exposed: []uint16{8080}},
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("app-already-desired is already running"))
			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Context("when the docker repo url is malformed", func() {
			It("Returns an error", func() {
				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:                 "nescafe-app",
					StartCommand:         "/app",
					DockerImagePath:      "¥¥¥Bad-Docker¥¥¥",
					AppArgs:              []string{},
					EnvironmentVariables: map[string]string{},
					Privileged:           false,
					Instances:            1,
					MemoryMB:             128,
					DiskMB:               1024,
					Ports:                docker_app_runner.PortConfig{Exposed: []uint16{8080}, Monitored: 8080},
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid repository name (¥¥¥Bad-Docker¥¥¥), only [a-z0-9-_.] are allowed"))
			})
		})

		Context("when the receptor returns errors", func() {
			It("returns upsert domain errors", func() {
				upsertError := errors.New("You're not that fresh, buddy.")
				fakeReceptorClient.UpsertDomainReturns(upsertError)

				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:                 "nescafe-app",
					StartCommand:         "faily/boom",
					DockerImagePath:      "borked_app",
					AppArgs:              []string{},
					EnvironmentVariables: map[string]string{},
					Privileged:           false,
					Instances:            1,
					MemoryMB:             128,
					DiskMB:               1024,
					Ports:                docker_app_runner.PortConfig{Exposed: []uint16{8080}, Monitored: 8080},
				})

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(upsertError))
			})

			It("returns desiring lrp errors", func() {
				receptorError := errors.New("error - Desiring an LRP")
				fakeReceptorClient.CreateDesiredLRPReturns(receptorError)

				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:                 "nescafe-app",
					StartCommand:         "faily/boom",
					DockerImagePath:      "borked_app",
					AppArgs:              []string{},
					EnvironmentVariables: map[string]string{},
					Privileged:           false,
					Instances:            1,
					MemoryMB:             128,
					DiskMB:               1024,
					Ports:                docker_app_runner.PortConfig{Exposed: []uint16{8080}, Monitored: 8080},
				})

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(receptorError))
			})

			It("returns existing count errors", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.CreateDockerApp(docker_app_runner.CreateDockerAppParams{
					Name:                 "nescafe-app",
					StartCommand:         "faily/boom",
					DockerImagePath:      "/app-bork-statement",
					AppArgs:              []string{},
					EnvironmentVariables: map[string]string{},
					Privileged:           false,
					Instances:            1,
					MemoryMB:             128,
					DiskMB:               1024,
					Ports:                docker_app_runner.PortConfig{Exposed: []uint16{8080}, Monitored: 8080},
				})

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(receptorError))
			})
		})
	})

	Describe("CreateAppFromJson", func() {

		It("Creates an app from JSON", func() {
			fakeReceptorClient.CreateDesiredLRPReturns(nil)

			desiredLRP := receptor.DesiredLRPCreateRequest{
				ProcessGuid:          "americano-app",
				Domain:               "lattice",
				RootFS:               "docker:///runtest/runner#latest",
				Instances:            22,
				EnvironmentVariables: []receptor.EnvironmentVariable{receptor.EnvironmentVariable{Name: "APPROOT", Value: "/root/env/path"}, receptor.EnvironmentVariable{Name: "PORT", Value: "2000"}},
				Routes: route_helpers.AppRoutes{
					route_helpers.AppRoute{Hostnames: []string{"americano-app.myDiegoInstall.com", "americano-app-2000.myDiegoInstall.com"}, Port: 2000},
					route_helpers.AppRoute{Hostnames: []string{"americano-app-4000.myDiegoInstall.com"}, Port: 4000},
				}.RoutingInfo(),
				CPUWeight:  67,
				MemoryMB:   128,
				DiskMB:     1024,
				Privileged: true,
				Ports:      []uint16{2000, 4000},
				LogGuid:    "americano-app",
				LogSource:  "APP",
				Setup: &models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
					To:   "/tmp",
				},
				Action: &models.RunAction{
					Path:       "/app-run-statement",
					Args:       []string{"app", "arg1", "--app", "arg 2"},
					Privileged: true,
					Dir:        "/user/web/myappdir",
				},
				Monitor: &models.RunAction{
					Path:      "/tmp/healthcheck",
					Args:      []string{"-port", "2000"},
					LogSource: "HEALTH",
				},
			}

			lrpJson, marshalErr := json.Marshal(desiredLRP)
			Expect(marshalErr).ToNot(HaveOccurred())

			lrpName, err := appRunner.CreateLrp(lrpJson)

			Expect(err).ToNot(HaveOccurred())
			Expect(lrpName).To(Equal("americano-app"))

			Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
			domain, ttl := fakeReceptorClient.UpsertDomainArgsForCall(0)
			Expect(domain).To(Equal("lattice"))
			Expect(ttl).To(Equal(time.Duration(0)))

			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0)).To(Equal(desiredLRP))
		})

		It("returns errors if the app is already desired", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "app-already-desired", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			desiredLRP := receptor.DesiredLRPCreateRequest{
				ProcessGuid: "app-already-desired",
			}
			lrpJson, marshalErr := json.Marshal(desiredLRP)
			Expect(marshalErr).ToNot(HaveOccurred())

			lrpName, err := appRunner.CreateLrp(lrpJson)

			Expect(err).To(HaveOccurred())
			Expect(lrpName).To(Equal("app-already-desired"))

			Expect(err.Error()).To(Equal("app-already-desired is already running"))
			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(0))
		})

		Context("when 'lattice-debug' is passed as the appId", func() {
			It("is an error because that id is reserved for the lattice-debug log stream", func() {
				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "lattice-debug",
				}

				lrpJson, marshalErr := json.Marshal(desiredLRP)
				Expect(marshalErr).ToNot(HaveOccurred())

				lrpName, err := appRunner.CreateLrp(lrpJson)

				Expect(err).To(HaveOccurred())
				Expect(lrpName).To(Equal("lattice-debug"))
				Expect(err.Error()).To(Equal(docker_app_runner.AttemptedToCreateLatticeDebugErrorMessage))
				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(0))
			})
		})

		It("returns an error for invalid JSON", func() {
			lrpName, err := appRunner.CreateLrp([]byte(`{"Value":"test value`))

			Expect(err).To(HaveOccurred())
			Expect(lrpName).To(BeEmpty())
			Expect(err.Error()).To(Equal("unexpected end of JSON input"))
			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(0))
		})

		Context("when the receptor returns errors", func() {

			It("returns existing count errors", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "nescafe-app",
				}
				lrpJson, marshalErr := json.Marshal(desiredLRP)
				Expect(marshalErr).ToNot(HaveOccurred())

				lrpName, err := appRunner.CreateLrp(lrpJson)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(receptorError))
				Expect(lrpName).To(Equal("nescafe-app"))
			})

			It("returns upsert domain errors", func() {
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)

				upsertError := errors.New("You're not that fresh, buddy.")
				fakeReceptorClient.UpsertDomainReturns(upsertError)

				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "whatever-app",
				}
				lrpJson, marshalErr := json.Marshal(desiredLRP)
				Expect(marshalErr).ToNot(HaveOccurred())

				lrpName, err := appRunner.CreateLrp(lrpJson)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(upsertError))
				Expect(lrpName).To(Equal("whatever-app"))
			})

			It("returns existing count errors", func() {
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)
				fakeReceptorClient.UpsertDomainReturns(nil)

				receptorError := errors.New("error - some error creating app")
				fakeReceptorClient.CreateDesiredLRPReturns(receptorError)

				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "nescafe-app",
				}
				lrpJson, marshalErr := json.Marshal(desiredLRP)
				Expect(marshalErr).ToNot(HaveOccurred())

				lrpName, err := appRunner.CreateLrp(lrpJson)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(receptorError))
				Expect(lrpName).To(Equal("nescafe-app"))
			})
		})

	})

	Describe("ScaleApp", func() {

		It("Scales a Docker App", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			instanceCount := 25

			err := appRunner.ScaleApp("americano-app", instanceCount)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeReceptorClient.UpdateDesiredLRPCallCount()).To(Equal(1))
			processGuid, updateRequest := fakeReceptorClient.UpdateDesiredLRPArgsForCall(0)
			Expect(processGuid).To(Equal("americano-app"))

			Expect(updateRequest).To(Equal(receptor.DesiredLRPUpdateRequest{Instances: &instanceCount}))
		})

		It("returns errors if the app is NOT already started", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.ScaleApp("app-not-running", 15)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("app-not-running is not started."))
			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Context("returning errors from the receptor", func() {
			It("returns desiring lrp errors", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

				receptorError := errors.New("error - Updating an LRP")
				fakeReceptorClient.UpdateDesiredLRPReturns(receptorError)

				err := appRunner.ScaleApp("americano-app", 17)
				Expect(err).To(Equal(receptorError))
			})

			It("returns errors fetching the existing lrp count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.ScaleApp("nescafe-app", 2)
				Expect(err).To(Equal(receptorError))
			})
		})
	})

	Describe("UpdateAppRoutes", func() {

		It("Updates the Routes", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app"}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			expectedRouteOverrides := docker_app_runner.RouteOverrides{
				docker_app_runner.RouteOverride{
					HostnamePrefix: "foo.com",
					Port:           8080,
				},
				docker_app_runner.RouteOverride{
					HostnamePrefix: "bar.com",
					Port:           9090,
				},
			}

			err := appRunner.UpdateAppRoutes("americano-app", expectedRouteOverrides)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeReceptorClient.UpdateDesiredLRPCallCount()).To(Equal(1))
			processGuid, updateRequest := fakeReceptorClient.UpdateDesiredLRPArgsForCall(0)
			Expect(processGuid).To(Equal("americano-app"))

			expectedRoutes := route_helpers.AppRoutes{
				route_helpers.AppRoute{Hostnames: []string{"foo.com.myDiegoInstall.com"}, Port: 8080},
				route_helpers.AppRoute{Hostnames: []string{"bar.com.myDiegoInstall.com"}, Port: 9090},
			}

			Expect(route_helpers.AppRoutesFromRoutingInfo(updateRequest.Routes)).To(ContainExactly(expectedRoutes))
		})

		It("returns errors if the app is NOT already started", func() {
			expectedRouteOverrides := docker_app_runner.RouteOverrides{
				docker_app_runner.RouteOverride{
					HostnamePrefix: "foo.com",
					Port:           8080,
				},
				docker_app_runner.RouteOverride{
					HostnamePrefix: "bar.com",
					Port:           9090,
				},
			}

			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app"}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.UpdateAppRoutes("app-not-running", expectedRouteOverrides)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("app-not-running is not started."))
			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Context("returning errors from the receptor", func() {
			It("returns desiring lrp errors", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

				receptorError := errors.New("error - Updating an LRP")
				fakeReceptorClient.UpdateDesiredLRPReturns(receptorError)

				err := appRunner.ScaleApp("americano-app", 17)
				Expect(err).To(Equal(receptorError))
			})

			It("returns errors fetching the existing lrp count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.UpdateAppRoutes("nescafe-app", nil)
				Expect(err).To(Equal(receptorError))
			})
		})
	})

	Describe("RemoveApp", func() {
		It("Removes a Docker App", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			fakeReceptorClient.DeleteDesiredLRPReturns(nil)

			err := appRunner.RemoveApp("americano-app")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeReceptorClient.DeleteDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.DeleteDesiredLRPArgsForCall(0)).To(Equal("americano-app"))
		})

		It("returns errors if the app is NOT already started", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.RemoveApp("app-not-running")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("app-not-running is not started."))
			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))

		})

		Describe("returning errors from the receptor", func() {
			It("returns deleting lrp errors", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

				deletingError := errors.New("deleting failed")
				fakeReceptorClient.DeleteDesiredLRPReturns(deletingError)

				err := appRunner.RemoveApp("americano-app")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deletingError))
			})

			It("returns errors fetching the existing count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.RemoveApp("nescafe-app")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(receptorError))
			})
		})
	})
})
