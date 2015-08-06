package app_runner_test

import (
	"encoding/json"
	"errors"
	"time"

	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/reserved_app_ids"
	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("AppRunner", func() {

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		appRunner          app_runner.AppRunner
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		appRunner = app_runner.New(fakeReceptorClient, "myDiegoInstall.com")
	})

	Describe("CreateApp", func() {
		var createAppParams app_runner.CreateAppParams

		BeforeEach(func() {
			appArgs := []string{"app", "arg1", "--app", "arg 2"}
			appEnv := map[string]string{"APPROOT": "/root/env/path"}
			createAppParams = app_runner.CreateAppParams{
				AppEnvironmentParams: app_runner.AppEnvironmentParams{
					EnvironmentVariables: appEnv,
					Monitor: app_runner.MonitorConfig{
						Method: app_runner.PortMonitor,
						Port:   2000,
					},
					Instances:    22,
					CPUWeight:    67,
					MemoryMB:     128,
					DiskMB:       1024,
					ExposedPorts: []uint16{2000, 4000},
					WorkingDir:   "/user/web/myappdir",
				},

				Name:         "americano-app",
				StartCommand: "/app-run-statement",
				RootFS:       "/runtest/runner",
				AppArgs:      appArgs,
				Annotation:   "some annotation",

				Setup: &models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
					To:   "/tmp",
					User: "vcap",
				},
			}
		})

		It("Upserts lattice domain so that it is always fresh, then starts the Docker App", func() {
			err := appRunner.CreateApp(createAppParams)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
			domain, ttl := fakeReceptorClient.UpsertDomainArgsForCall(0)
			Expect(domain).To(Equal("lattice"))
			Expect(ttl).To(Equal(time.Duration(0)))

			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
			req := fakeReceptorClient.CreateDesiredLRPArgsForCall(0)
			Expect(req.ProcessGuid).To(Equal("americano-app"))
			Expect(req.Domain).To(Equal("lattice"))
			Expect(req.RootFS).To(Equal("/runtest/runner"))
			Expect(req.Instances).To(Equal(22))
			Expect(req.EnvironmentVariables).To(ConsistOf(
				receptor.EnvironmentVariable{Name: "APPROOT", Value: "/root/env/path"},
				receptor.EnvironmentVariable{Name: "PORT", Value: "2000"},
			))
			Expect(req.Routes).To(Equal(route_helpers.AppRoutes{
				route_helpers.AppRoute{Hostnames: []string{"americano-app.myDiegoInstall.com", "americano-app-2000.myDiegoInstall.com"}, Port: 2000},
				route_helpers.AppRoute{Hostnames: []string{"americano-app-4000.myDiegoInstall.com"}, Port: 4000},
			}.RoutingInfo()))
			Expect(req.CPUWeight).To(Equal(uint(67)))
			Expect(req.MemoryMB).To(Equal(128))
			Expect(req.DiskMB).To(Equal(1024))
			Expect(req.Privileged).To(BeFalse())
			Expect(req.Ports).To(ConsistOf(uint16(2000), uint16(4000)))
			Expect(req.LogGuid).To(Equal("americano-app"))
			Expect(req.LogSource).To(Equal("APP"))
			Expect(req.MetricsGuid).To(Equal("americano-app"))
			Expect(req.Annotation).To(Equal("some annotation"))

			Expect(req.Setup).To(BeAssignableToTypeOf(&models.DownloadAction{}))
			reqSetup, ok := req.Setup.(*models.DownloadAction)
			Expect(ok).To(BeTrue())
			Expect(reqSetup.From).To(Equal("http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz"))
			Expect(reqSetup.To).To(Equal("/tmp"))
			Expect(reqSetup.User).To(Equal("vcap"))

			Expect(req.Action).To(BeAssignableToTypeOf(&models.RunAction{}))
			reqAction, ok := req.Action.(*models.RunAction)
			Expect(ok).To(BeTrue())
			Expect(reqAction.Path).To(Equal("/app-run-statement"))
			Expect(reqAction.Args).To(Equal([]string{"app", "arg1", "--app", "arg 2"}))
			Expect(reqAction.Dir).To(Equal("/user/web/myappdir"))
			Expect(reqAction.User).To(Equal("vcap"))

			Expect(req.Monitor).To(BeAssignableToTypeOf(&models.RunAction{}))
			reqMonitor, ok := req.Monitor.(*models.RunAction)
			Expect(ok).To(BeTrue())
			Expect(reqMonitor.Path).To(Equal("/tmp/healthcheck"))
			Expect(reqMonitor.Args).To(Equal([]string{"-port", "2000"}))
			Expect(reqMonitor.LogSource).To(Equal("HEALTH"))
			Expect(reqMonitor.User).To(Equal("vcap"))
		})

		Context("when 'lattice-debug' is passed as the appId", func() {
			It("is an error because that id is reserved for the lattice-debug log stream", func() {
				createAppParams = app_runner.CreateAppParams{
					Name: reserved_app_ids.LatticeDebugLogStreamAppId,
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).To(MatchError(app_runner.AttemptedToCreateLatticeDebugErrorMessage))
			})
		})

		Context("when Privileged is true on the CreateAppParams", func() {
			It("sets Privileged=true and User=root on the lrp request and RunActions, respectively", func() {
				createAppParams.Privileged = true

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				req := fakeReceptorClient.CreateDesiredLRPArgsForCall(0)
				Expect(req.Privileged).To(BeTrue())

				reqAction, ok := req.Action.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqAction.User).To(Equal("root"))

				reqMonitor, ok := req.Monitor.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqMonitor.User).To(Equal("root"))
			})
		})

		Context("when Privileged is false on the CreateAppParams", func() {
			It("sets Privileged=false and User=vcap on the lrp request and RunActions, respectively", func() {
				createAppParams.Privileged = false

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				req := fakeReceptorClient.CreateDesiredLRPArgsForCall(0)
				Expect(req.Privileged).To(BeFalse())

				reqAction, ok := req.Action.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqAction.User).To(Equal("vcap"))

				reqMonitor, ok := req.Monitor.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqMonitor.User).To(Equal("vcap"))
			})
		})

		Context("when tcp routes are not empty", func() {
			BeforeEach(func() {
				createAppParams.AppEnvironmentParams.TcpRoutes = app_runner.TcpRoutes{
					app_runner.TcpRoute{ExternalPort: 60000, Port: 2000},
					app_runner.TcpRoute{ExternalPort: 60010, Port: 2000},
					app_runner.TcpRoute{ExternalPort: 60020, Port: 3000},
				}
			})

			Context("and when route overrides are not empty", func() {
				BeforeEach(func() {
					createAppParams.AppEnvironmentParams.RouteOverrides = app_runner.RouteOverrides{
						app_runner.RouteOverride{HostnamePrefix: "wiggle", Port: 2000},
						app_runner.RouteOverride{HostnamePrefix: "swang", Port: 2000},
						app_runner.RouteOverride{HostnamePrefix: "shuffle", Port: 4000},
					}
				})

				It("uses the tcp routes", func() {
					err := appRunner.CreateApp(createAppParams)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
					routes := route_helpers.RoutesFromRoutingInfo(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Routes)

					Expect(routes.TcpRoutes).ShouldNot(BeNil())
					Expect(routes.TcpRoutes).Should(ContainExactly(
						route_helpers.TcpRoutes{
							route_helpers.TcpRoute{
								ExternalPort: 60000,
								Port:         2000,
							},
							route_helpers.TcpRoute{
								ExternalPort: 60010,
								Port:         2000,
							},
							route_helpers.TcpRoute{
								ExternalPort: 60020,
								Port:         3000,
							},
						},
					))

					Expect(routes.AppRoutes).To(ContainExactly(
						route_helpers.AppRoutes{
							route_helpers.AppRoute{
								Hostnames: []string{"wiggle.myDiegoInstall.com", "swang.myDiegoInstall.com"},
								Port:      2000,
							},
							route_helpers.AppRoute{
								Hostnames: []string{"shuffle.myDiegoInstall.com"},
								Port:      4000,
							},
						}))
				})
			})

			Context("and when route overrides are empty", func() {
				It("uses the tcp routes", func() {
					err := appRunner.CreateApp(createAppParams)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
					routes := route_helpers.RoutesFromRoutingInfo(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Routes)

					Expect(routes.TcpRoutes).ShouldNot(BeNil())
					Expect(routes.TcpRoutes).Should(ContainExactly(
						route_helpers.TcpRoutes{
							route_helpers.TcpRoute{
								ExternalPort: 60000,
								Port:         2000,
							},
							route_helpers.TcpRoute{
								ExternalPort: 60010,
								Port:         2000,
							},
							route_helpers.TcpRoute{
								ExternalPort: 60020,
								Port:         3000,
							},
						},
					))

					Expect(routes.AppRoutes).To(ContainExactly(
						route_helpers.AppRoutes{
							route_helpers.AppRoute{
								Hostnames: []string{"americano-app.myDiegoInstall.com", "americano-app-2000.myDiegoInstall.com"},
								Port:      2000,
							},
							route_helpers.AppRoute{
								Hostnames: []string{"americano-app-4000.myDiegoInstall.com"},
								Port:      4000,
							},
						}))
				})
			})
		})

		Context("and when NoRoutes is true", func() {
			It("does not register any routes for the app", func() {
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						NoRoutes: true,
						TcpRoutes: app_runner.TcpRoutes{
							app_runner.TcpRoute{ExternalPort: 60000, Port: 2000},
							app_runner.TcpRoute{ExternalPort: 60010, Port: 2000},
							app_runner.TcpRoute{ExternalPort: 60020, Port: 3000},
						},
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Routes).To(
					Equal(route_helpers.Routes{AppRoutes: route_helpers.AppRoutes{}}.RoutingInfo()))
			})
		})

		Context("when route overrides are not empty", func() {
			It("uses the overriden routes instead of the defaults", func() {
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						RouteOverrides: app_runner.RouteOverrides{
							app_runner.RouteOverride{HostnamePrefix: "wiggle", Port: 2000},
							app_runner.RouteOverride{HostnamePrefix: "swang", Port: 2000},
							app_runner.RouteOverride{HostnamePrefix: "shuffle", Port: 4000},
						},
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				appRoutes := route_helpers.AppRoutesFromRoutingInfo(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Routes)
				Expect(appRoutes).To(ContainExactly(
					route_helpers.AppRoutes{
						route_helpers.AppRoute{
							Hostnames: []string{"wiggle.myDiegoInstall.com", "swang.myDiegoInstall.com"},
							Port:      2000,
						},
						route_helpers.AppRoute{
							Hostnames: []string{"shuffle.myDiegoInstall.com"},
							Port:      4000,
						},
					}))
			})
		})

		Context("when NoRoutes is true", func() {
			It("does not register any routes for the app", func() {
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						RouteOverrides: app_runner.RouteOverrides{
							app_runner.RouteOverride{HostnamePrefix: "wiggle", Port: 2000},
						},
						NoRoutes: true,
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Routes).To(Equal(route_helpers.AppRoutes{}.RoutingInfo()))
			})
		})

		Context("when Monitor is NoMonitor", func() {
			It("does not pass a monitor action, regardless of whether or not a monitor port is passed", func() {
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						Monitor: app_runner.MonitorConfig{
							Method: app_runner.NoMonitor,
							Port:   uint16(4444),
						},
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0).Monitor).To(BeExactlyNil())
			})
		})

		Context("when monitoring a port", func() {
			It("sets the timeout for the monitor", func() {
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, nil)
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						Monitor: app_runner.MonitorConfig{
							Method:  app_runner.PortMonitor,
							Port:    2345,
							Timeout: 15 * time.Second,
						},
						ExposedPorts: []uint16{2345},
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				req := fakeReceptorClient.CreateDesiredLRPArgsForCall(0)

				Expect(req.Monitor).To(BeAssignableToTypeOf(&models.RunAction{}))
				reqMonitor, ok := req.Monitor.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqMonitor.Path).To(Equal("/tmp/healthcheck"))
				Expect(reqMonitor.Args).To(Equal([]string{"-timeout", "15s", "-port", "2345"}))
				Expect(reqMonitor.LogSource).To(Equal("HEALTH"))
				Expect(reqMonitor.User).To(Equal("vcap"))
			})
		})

		Context("when monitoring a url", func() {
			It("passes a monitor action", func() {
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						Monitor: app_runner.MonitorConfig{
							Method: app_runner.URLMonitor,
							Port:   1234,
							URI:    "/healthy/endpoint",
						},
						ExposedPorts: []uint16{1234},
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				req := fakeReceptorClient.CreateDesiredLRPArgsForCall(0)

				Expect(req.Monitor).To(BeAssignableToTypeOf(&models.RunAction{}))
				reqMonitor, ok := req.Monitor.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqMonitor.Path).To(Equal("/tmp/healthcheck"))
				Expect(reqMonitor.Args).To(Equal([]string{"-port", "1234", "-uri", "/healthy/endpoint"}))
				Expect(reqMonitor.LogSource).To(Equal("HEALTH"))
				Expect(reqMonitor.User).To(Equal("vcap"))
			})

			It("sets the timeout for the monitor", func() {
				createAppParams = app_runner.CreateAppParams{
					AppEnvironmentParams: app_runner.AppEnvironmentParams{
						Monitor: app_runner.MonitorConfig{
							Method:  app_runner.URLMonitor,
							Port:    1234,
							URI:     "/healthy/endpoint",
							Timeout: 20 * time.Second,
						},
						ExposedPorts: []uint16{1234},
					},
				}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
				req := fakeReceptorClient.CreateDesiredLRPArgsForCall(0)

				Expect(req.Monitor).To(BeAssignableToTypeOf(&models.RunAction{}))
				reqMonitor, ok := req.Monitor.(*models.RunAction)
				Expect(ok).To(BeTrue())
				Expect(reqMonitor.Path).To(Equal("/tmp/healthcheck"))
				Expect(reqMonitor.Args).To(Equal([]string{"-timeout", "20s", "-port", "1234", "-uri", "/healthy/endpoint"}))
				Expect(reqMonitor.LogSource).To(Equal("HEALTH"))
				Expect(reqMonitor.User).To(Equal("vcap"))
			})
		})

		It("returns errors if the app is already desired", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "app-already-desired", Instances: 1}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			createAppParams = app_runner.CreateAppParams{
				Name: "app-already-desired",
			}

			err := appRunner.CreateApp(createAppParams)
			Expect(err).To(MatchError("app-already-desired is already running"))

			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Context("when the receptor returns errors", func() {
			It("returns upsert domain errors", func() {
				upsertError := errors.New("You're not that fresh, buddy.")
				fakeReceptorClient.UpsertDomainReturns(upsertError)
				createAppParams = app_runner.CreateAppParams{}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).To(MatchError(upsertError))
			})

			It("returns desiring lrp errors", func() {
				receptorError := errors.New("error - Desiring an LRP")
				fakeReceptorClient.CreateDesiredLRPReturns(receptorError)
				createAppParams = app_runner.CreateAppParams{}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).To(MatchError(receptorError))
			})

			It("returns existing count errors", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)
				createAppParams = app_runner.CreateAppParams{}

				err := appRunner.CreateApp(createAppParams)
				Expect(err).To(MatchError(receptorError))
			})
		})
	})

	Describe("SubmitLrp", func() {
		It("Creates an app from JSON", func() {
			desiredLRP := receptor.DesiredLRPCreateRequest{
				ProcessGuid: "americano-app",
				Domain:      "lattice",
				RootFS:      "docker:///runtest/runner#latest",
				Instances:   22,
				EnvironmentVariables: []receptor.EnvironmentVariable{
					receptor.EnvironmentVariable{Name: "APPROOT", Value: "/root/env/path"},
					receptor.EnvironmentVariable{Name: "PORT", Value: "2000"},
				},
				Routes: route_helpers.AppRoutes{
					route_helpers.AppRoute{Hostnames: []string{"americano-app.myDiegoInstall.com", "americano-app-2000.myDiegoInstall.com"}, Port: 2000},
					route_helpers.AppRoute{Hostnames: []string{"americano-app-4000.myDiegoInstall.com"}, Port: 4000},
				}.RoutingInfo(),
				CPUWeight:   67,
				MemoryMB:    128,
				DiskMB:      1024,
				Privileged:  true,
				Ports:       []uint16{2000, 4000},
				LogGuid:     "americano-app",
				LogSource:   "APP",
				MetricsGuid: "americano-app",
				Setup: &models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
					To:   "/tmp",
					User: "vcap",
				},
				Action: &models.RunAction{
					Path: "/app-run-statement",
					Args: []string{"app", "arg1", "--app", "arg 2"},
					Dir:  "/user/web/myappdir",
					User: "vcap",
				},
				Monitor: &models.RunAction{
					Path:      "/tmp/healthcheck",
					Args:      []string{"-port", "2000"},
					LogSource: "HEALTH",
					User:      "vcap",
				},
			}

			lrpJson, err := json.Marshal(desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			lrpName, err := appRunner.SubmitLrp(lrpJson)
			Expect(err).NotTo(HaveOccurred())
			Expect(lrpName).To(Equal("americano-app"))

			Expect(fakeReceptorClient.UpsertDomainCallCount()).To(Equal(1))
			domain, ttl := fakeReceptorClient.UpsertDomainArgsForCall(0)
			Expect(domain).To(Equal("lattice"))
			Expect(ttl).To(BeZero())

			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0)).To(Equal(desiredLRP))
		})

		It("returns errors if the app is already desired", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{ProcessGuid: "app-already-desired", Instances: 1},
			}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			desiredLRP := receptor.DesiredLRPCreateRequest{
				ProcessGuid: "app-already-desired",
			}

			lrpJSON, err := json.Marshal(desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			lrpName, err := appRunner.SubmitLrp(lrpJSON)
			Expect(err).To(MatchError("app-already-desired is already running"))
			Expect(lrpName).To(Equal("app-already-desired"))

			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(0))
		})

		Context("when 'lattice-debug' is passed as the appId", func() {
			It("is an error because that id is reserved for the lattice-debug log stream", func() {
				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "lattice-debug",
				}
				lrpJSON, err := json.Marshal(desiredLRP)
				Expect(err).NotTo(HaveOccurred())

				lrpName, err := appRunner.SubmitLrp(lrpJSON)
				Expect(err).To(MatchError(app_runner.AttemptedToCreateLatticeDebugErrorMessage))
				Expect(lrpName).To(Equal("lattice-debug"))

				Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(0))
			})
		})

		It("returns an error for invalid JSON", func() {
			lrpName, err := appRunner.SubmitLrp([]byte(`{"Value":"test value`))
			Expect(err).To(MatchError("unexpected end of JSON input"))
			Expect(lrpName).To(BeEmpty())

			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(0))
		})

		Context("when the receptor returns errors", func() {
			It("returns existing count errors", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)
				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "nescafe-app",
				}
				lrpJSON, err := json.Marshal(desiredLRP)
				Expect(err).NotTo(HaveOccurred())

				lrpName, err := appRunner.SubmitLrp(lrpJSON)
				Expect(err).To(MatchError(receptorError))
				Expect(lrpName).To(Equal("nescafe-app"))
			})

			It("returns upsert domain errors", func() {
				upsertError := errors.New("You're not that fresh, buddy.")
				fakeReceptorClient.UpsertDomainReturns(upsertError)
				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "whatever-app",
				}
				lrpJSON, err := json.Marshal(desiredLRP)
				Expect(err).NotTo(HaveOccurred())

				lrpName, err := appRunner.SubmitLrp(lrpJSON)
				Expect(err).To(MatchError(upsertError))
				Expect(lrpName).To(Equal("whatever-app"))
			})

			It("returns existing count errors", func() {
				receptorError := errors.New("error - some error creating app")
				fakeReceptorClient.CreateDesiredLRPReturns(receptorError)
				desiredLRP := receptor.DesiredLRPCreateRequest{
					ProcessGuid: "nescafe-app",
				}
				lrpJSON, err := json.Marshal(desiredLRP)
				Expect(err).NotTo(HaveOccurred())

				lrpName, err := appRunner.SubmitLrp(lrpJSON)
				Expect(err).To(MatchError(receptorError))
				Expect(lrpName).To(Equal("nescafe-app"))
			})
		})
	})

	Describe("ScaleApp", func() {
		It("scales a Docker App", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1},
			}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			instanceCount := 25

			err := appRunner.ScaleApp("americano-app", instanceCount)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReceptorClient.UpdateDesiredLRPCallCount()).To(Equal(1))
			processGuid, updateRequest := fakeReceptorClient.UpdateDesiredLRPArgsForCall(0)
			Expect(processGuid).To(Equal("americano-app"))
			Expect(*updateRequest.Instances).To(Equal(instanceCount))
		})

		It("returns errors if the app is NOT already started", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1},
			}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.ScaleApp("app-not-running", 15)
			Expect(err).To(MatchError("app-not-running is not started."))

			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Context("returning errors from the receptor", func() {
			It("returns desiring lrp errors", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
				receptorError := errors.New("error - Updating an LRP")
				fakeReceptorClient.UpdateDesiredLRPReturns(receptorError)

				err := appRunner.ScaleApp("americano-app", 17)
				Expect(err).To(MatchError(receptorError))
			})

			It("returns errors fetching the existing lrp count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.ScaleApp("nescafe-app", 2)
				Expect(err).To(MatchError(receptorError))
			})
		})
	})

	Describe("UpdateAppRoutes", func() {
		It("updates the Routes", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{ProcessGuid: "americano-app"},
			}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			expectedRouteOverrides := app_runner.RouteOverrides{
				app_runner.RouteOverride{
					HostnamePrefix: "foo.com",
					Port:           8080,
				},
				app_runner.RouteOverride{
					HostnamePrefix: "bar.com",
					Port:           9090,
				},
			}

			err := appRunner.UpdateAppRoutes("americano-app", expectedRouteOverrides)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReceptorClient.UpdateDesiredLRPCallCount()).To(Equal(1))
			processGuid, updateRequest := fakeReceptorClient.UpdateDesiredLRPArgsForCall(0)
			Expect(processGuid).To(Equal("americano-app"))
			expectedRoutes := route_helpers.AppRoutes{
				route_helpers.AppRoute{Hostnames: []string{"foo.com.myDiegoInstall.com"}, Port: 8080},
				route_helpers.AppRoute{Hostnames: []string{"bar.com.myDiegoInstall.com"}, Port: 9090},
			}
			Expect(route_helpers.AppRoutesFromRoutingInfo(updateRequest.Routes)).To(ContainExactly(expectedRoutes))
		})

		Context("when an empty routes is passed", func() {
			It("deregisters the routes", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app"}}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

				err := appRunner.UpdateAppRoutes("americano-app", app_runner.RouteOverrides{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeReceptorClient.UpdateDesiredLRPCallCount()).To(Equal(1))
				processGuid, updateRequest := fakeReceptorClient.UpdateDesiredLRPArgsForCall(0)
				Expect(processGuid).To(Equal("americano-app"))
				Expect(updateRequest.Routes).To(Equal(route_helpers.AppRoutes{}.RoutingInfo()))
			})
		})

		It("returns errors if the app is NOT already started", func() {
			expectedRouteOverrides := app_runner.RouteOverrides{
				app_runner.RouteOverride{
					HostnamePrefix: "foo.com",
					Port:           8080,
				},
				app_runner.RouteOverride{
					HostnamePrefix: "bar.com",
					Port:           9090,
				},
			}
			desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{ProcessGuid: "americano-app"}}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.UpdateAppRoutes("app-not-running", expectedRouteOverrides)
			Expect(err).To(MatchError("app-not-running is not started."))

			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Context("returning errors from the receptor", func() {
			It("returns desiring lrp errors", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{receptor.DesiredLRPResponse{
					ProcessGuid: "americano-app", Instances: 1},
				}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
				receptorError := errors.New("error - Updating an LRP")
				fakeReceptorClient.UpdateDesiredLRPReturns(receptorError)

				err := appRunner.ScaleApp("americano-app", 17)
				Expect(err).To(MatchError(receptorError))
			})

			It("returns errors fetching the existing lrp count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.UpdateAppRoutes("nescafe-app", nil)
				Expect(err).To(MatchError(receptorError))
			})
		})
	})

	Describe("RemoveApp", func() {
		It("Removes a Docker App", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1},
			}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
			fakeReceptorClient.DeleteDesiredLRPReturns(nil)

			err := appRunner.RemoveApp("americano-app")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReceptorClient.DeleteDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.DeleteDesiredLRPArgsForCall(0)).To(Equal("americano-app"))
		})

		It("returns errors if the app is NOT already started", func() {
			desiredLRPs := []receptor.DesiredLRPResponse{
				receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1},
			}
			fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)

			err := appRunner.RemoveApp("app-not-running")
			Expect(err).To(MatchError("app-not-running is not started."))

			Expect(fakeReceptorClient.DesiredLRPsCallCount()).To(Equal(1))
		})

		Describe("returning errors from the receptor", func() {
			It("returns deleting lrp errors", func() {
				desiredLRPs := []receptor.DesiredLRPResponse{
					receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1},
				}
				fakeReceptorClient.DesiredLRPsReturns(desiredLRPs, nil)
				deletingError := errors.New("deleting failed")
				fakeReceptorClient.DeleteDesiredLRPReturns(deletingError)

				err := appRunner.RemoveApp("americano-app")
				Expect(err).To(MatchError(deletingError))
			})

			It("returns errors fetching the existing count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.DesiredLRPsReturns([]receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.RemoveApp("nescafe-app")
				Expect(err).To(MatchError(receptorError))
			})
		})
	})
})
