package command_factory_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/docker_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/docker_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/docker_runner/docker_metadata_fetcher/fake_docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"

	app_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
)

var _ = Describe("CommandFactory", func() {
	var (
		fakeAppRunner                 *fake_app_runner.FakeAppRunner
		fakeAppExaminer               *fake_app_examiner.FakeAppExaminer
		outputBuffer                  *gbytes.Buffer
		terminalUI                    terminal.UI
		domain                        string = "192.168.11.11.xip.io"
		fakeClock                     *fakeclock.FakeClock
		fakeDockerMetadataFetcher     *fake_docker_metadata_fetcher.FakeDockerMetadataFetcher
		appRunnerCommandFactoryConfig command_factory.DockerRunnerCommandFactoryConfig
		logger                        lager.Logger
		fakeTailedLogsOutputter       *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		fakeExitHandler               *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		fakeAppRunner = &fake_app_runner.FakeAppRunner{}
		fakeAppExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeDockerMetadataFetcher = &fake_docker_metadata_fetcher.FakeDockerMetadataFetcher{}
		fakeClock = fakeclock.NewFakeClock(time.Now())
		logger = lager.NewLogger("ltc-test")
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("CreateAppCommand", func() {
		var createCommand cli.Command

		BeforeEach(func() {
			env := []string{"SHELL=/bin/bash", "COLOR=Blue"}
			appRunnerCommandFactoryConfig = command_factory.DockerRunnerCommandFactoryConfig{
				AppRunner:   fakeAppRunner,
				AppExaminer: fakeAppExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: fakeDockerMetadataFetcher,
				Domain:                domain,
				Env:                   env,
				Clock:                 fakeClock,
				TailedLogsOutputter:   fakeTailedLogsOutputter,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewDockerRunnerCommandFactory(appRunnerCommandFactoryConfig)
			createCommand = commandFactory.MakeCreateAppCommand()

			fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
		})

		It("creates a Docker based app as specified in the command via the AppRunner", func() {
			args := []string{
				"--cpu-weight=57",
				"--memory-mb=12",
				"--disk-mb=12",
				"--routes=3000:route-3000-yay,1111:route-1111-wahoo,1111:route-1111-me-too",
				"--working-dir=/applications",
				"--run-as-root=true",
				"--instances=22",
				"--env=TIMEZONE=CST",
				`--env=LANG="Chicago English"`,
				`--env=JAVA_OPTS="-Djava.arg=/dev/urandom"`,
				"--env=COLOR",
				"--env=UNSET",
				"--timeout=28s",
				"cool-web-app",
				"superfun/app:mycooltag",
				"--",
				"/start-me-please",
				"AppArg0",
				`--appFlavor="purple"`,
			}
			fakeAppExaminer.RunningAppInstancesInfoReturns(22, false, nil)

			test_helpers.ExecuteCommandWithArgs(createCommand, args)

			Expect(fakeDockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
			Expect(fakeDockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("superfun/app:mycooltag"))

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams.Name).To(Equal("cool-web-app"))
			Expect(createAppParams.StartCommand).To(Equal("/start-me-please"))
			Expect(createAppParams.RootFS).To(Equal("docker:///superfun/app#mycooltag"))
			Expect(createAppParams.AppArgs).To(Equal([]string{"AppArg0", "--appFlavor=\"purple\""}))
			Expect(createAppParams.Instances).To(Equal(22))
			Expect(createAppParams.EnvironmentVariables).To(Equal(map[string]string{
				"TIMEZONE":     "CST",
				"LANG":         `"Chicago English"`,
				"JAVA_OPTS":    `"-Djava.arg=/dev/urandom"`,
				"PROCESS_GUID": "cool-web-app",
				"COLOR":        "Blue",
				"UNSET":        "",
			}))
			Expect(createAppParams.Privileged).To(Equal(true))
			Expect(createAppParams.CPUWeight).To(Equal(uint(57)))
			Expect(createAppParams.MemoryMB).To(Equal(12))
			Expect(createAppParams.DiskMB).To(Equal(12))
			Expect(createAppParams.Monitor.Method).To(Equal(app_runner.PortMonitor))
			Expect(createAppParams.Timeout).To(Equal(time.Second * 28))
			Expect(createAppParams.RouteOverrides).To(ContainExactly(app_runner.RouteOverrides{
				app_runner.RouteOverride{HostnamePrefix: "route-3000-yay", Port: 3000},
				app_runner.RouteOverride{HostnamePrefix: "route-1111-wahoo", Port: 1111},
				app_runner.RouteOverride{HostnamePrefix: "route-1111-me-too", Port: 1111},
			}))
			Expect(createAppParams.NoRoutes).To(BeFalse())
			Expect(createAppParams.WorkingDir).To(Equal("/applications"))

			Expect(createAppParams.Setup).To(BeAssignableToTypeOf(&models.DownloadAction{}))
			reqSetup, ok := createAppParams.Setup.(*models.DownloadAction)
			Expect(ok).To(BeTrue())
			Expect(reqSetup.From).To(Equal("http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz"))
			Expect(reqSetup.To).To(Equal("/tmp"))
			Expect(reqSetup.User).To(Equal("vcap"))

			Expect(outputBuffer).To(test_helpers.SayLine("Creating App: cool-web-app"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("cool-web-app is now running.")))
			Expect(outputBuffer).To(test_helpers.SayLine("App is reachable at:"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://route-3000-yay.192.168.11.11.xip.io")))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://route-1111-wahoo.192.168.11.11.xip.io")))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://route-1111-me-too.192.168.11.11.xip.io")))
		})

		Context("when the PROCESS_GUID is passed in as --env", func() {
			It("sets the PROCESS_GUID to the value passed in", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{""}}, nil)
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
				args := []string{
					"app-to-start",
					"fun-org/app",
					"--env=PROCESS_GUID=MyHappyGuid",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				appEnvVars := createAppParams.EnvironmentVariables
				processGuidEnvVar, found := appEnvVars["PROCESS_GUID"]
				Expect(found).To(BeTrue())
				Expect(processGuidEnvVar).To(Equal("MyHappyGuid"))
			})
		})

		Context("when a malformed routes flag is passed", func() {
			It("errors out when the port is not an int", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--routes=woo:aahh",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MalformedRouteErrorMessage))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("errors out when there is no colon", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--routes=8888",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MalformedRouteErrorMessage))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})
		})

		Describe("Exposed Ports", func() {
			BeforeEach(func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
			})

			It("exposes ports passed by --ports", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--ports=8080,9090",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.ExposedPorts).To(Equal([]uint16{8080, 9090}))
			})

			It("exposes ports from image metadata", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{
					ExposedPorts: []uint16{1200, 2701, 4302},
				}, nil)
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: 1200, 2701, 4302"))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.ExposedPorts).To(Equal([]uint16{1200, 2701, 4302}))
			})

			It("exposes --ports ports when both --ports and EXPOSE metadata exist", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{
					ExposedPorts: []uint16{1200, 2701, 4302},
				}, nil)
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--ports=8080,9090",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.ExposedPorts).To(Equal([]uint16{8080, 9090}))
			})

			Context("when the metadata does not have EXPOSE ports", func() {
				It("exposes the default port 8080", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--no-monitor",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
					Expect(createAppParams.ExposedPorts).To(Equal([]uint16{8080}))
				})
			})

			Context("when malformed --ports flag is passed", func() {
				It("blows up when you pass bad port strings", func() {
					args := []string{
						"--ports=1000,98feh34",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))

				})

				It("errors out when any port is > 65535 (max Linux port number)", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--ports=8080,65536",
						"--monitor-port=8080",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})
			})
		})

		//TODO:  little wonky - this test makes sure we default stuff, but says it's dealing w/ fetcher
		Describe("interactions with the docker metadata fetcher", func() {
			Context("when the docker image is hosted on a docker registry", func() {
				It("creates a Docker based app with sensible defaults and checks for metadata to know the image exists", func() {
					fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
					args := []string{
						"cool-web-app",
						"awesome/app",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say("No port specified, image metadata did not contain exposed ports. Defaulting to 8080.\n"))

					Expect(fakeDockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
					Expect(fakeDockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("awesome/app"))

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
					Expect(createAppParams.Privileged).To(Equal(false))
					Expect(createAppParams.MemoryMB).To(Equal(128))
					Expect(createAppParams.DiskMB).To(Equal(0))
					Expect(createAppParams.Monitor.Port).To(Equal(uint16(8080)))
					Expect(createAppParams.ExposedPorts).To(Equal([]uint16{8080}))
					Expect(createAppParams.Instances).To(Equal(1))
					Expect(createAppParams.WorkingDir).To(Equal("/"))
				})
			})

			Context("when the docker metadata fetcher returns an error", func() {
				It("exposes the error from trying to fetch the Docker metadata", func() {
					fakeDockerMetadataFetcher.FetchMetadataReturns(nil, errors.New("Docker Says No."))
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say("Error fetching image metadata: Docker Says No."))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadDocker}))
				})
			})
		})

		Describe("Monitor Config", func() {
			BeforeEach(func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
			})

			Context("when --no-monitor is passed", func() {
				It("does not monitor", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--no-monitor",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Method).To(Equal(app_runner.NoMonitor))
				})
			})

			Context("when --monitor-port is passed", func() {
				It("port-monitors a specified port", func() {
					args := []string{
						"--ports=1000,2000",
						"--monitor-port=2000",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Method).To(Equal(app_runner.PortMonitor))
					Expect(monitorConfig.Port).To(Equal(uint16(2000)))
				})

				It("prints an error when the monitored port is not exposed", func() {
					args := []string{
						"--ports=1000,1200",
						"--monitor-port=2000",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}
					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MonitorPortNotExposed))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})

			Context("when --monitor-url is passed", func() {
				It("url-monitors a specified url", func() {
					args := []string{
						"--ports=1000,2000",
						"--monitor-url=1000:/sup/yeah",
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Method).To(Equal(app_runner.URLMonitor))
					Expect(monitorConfig.Port).To(Equal(uint16(1000)))
				})

				It("prints an error if the url can't be split", func() {
					args := []string{
						"--ports=1000,2000",
						"--monitor-url=1000/sup/yeah",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})

				It("prints an error if the port is non-numeric", func() {
					args := []string{
						"--ports=1000,2000",
						"--monitor-url=TOTES:/sup/yeah",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidPortErrorMessage))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
				})

				It("prints an error when the monitored url port is not exposed", func() {
					args := []string{
						"--ports=1000,2000",
						"--monitor-url=1200:/sup/yeah",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MonitorPortNotExposed))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
				})
			})

			Context("when no monitoring options are passed", func() {
				It("port-monitors the first exposed port", func() {
					args := []string{
						"--ports=1000,2000",
						"cool-web-app",
						"superfun/app:mycooltag",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Method).To(Equal(app_runner.PortMonitor))
					Expect(monitorConfig.Port).To(Equal(uint16(1000)))
				})

				It("sets a timeout", func() {
					args := []string{
						"--monitor-timeout=5s",
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Timeout).To(Equal(5 * time.Second))
				})
			})

			Context("when multiple monitoring options are passed", func() {
				It("no-monitor takes precedence", func() {
					args := []string{
						"--ports=1200",
						"--monitor-url=1200:/sup/yeah",
						"--no-monitor",
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Method).To(Equal(app_runner.NoMonitor))
				})

				It("monitor-url takes precedence over monitor-port", func() {
					args := []string{
						"--ports=1200",
						"--monitor-url=1200:/sup/yeah",
						"--monitor-port=1200",
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
					monitorConfig := fakeAppRunner.CreateAppArgsForCall(0).Monitor
					Expect(monitorConfig.Method).To(Equal(app_runner.URLMonitor))
					Expect(monitorConfig.Port).To(Equal(uint16(1200)))
				})
			})
		})

		Context("when the --no-routes flag is passed", func() {
			It("calls app runner with NoRoutes equal to true", func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--no-routes",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).NotTo(test_helpers.Say("App is reachable at:"))
				Expect(outputBuffer).NotTo(test_helpers.Say("http://cool-web-app.192.168.11.11.xip.io"))

				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.NoRoutes).To(BeTrue())
			})
		})

		Context("when no working dir is provided, but the metadata has a working dir", func() {
			It("sets the working dir from the Docker metadata", func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{WorkingDir: "/work/it"}, nil)
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.WorkingDir).To(Equal("/work/it"))
			})
		})

		Context("when no start command is provided", func() {
			var args = []string{
				"cool-web-app",
				"fun-org/app",
			}

			BeforeEach(func() {
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
			})

			It("creates a Docker app with the create command retrieved from the docker image metadata", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{WorkingDir: "/this/directory/right/here", StartCommand: []string{"/fetch-start", "arg1", "arg2"}}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("No working directory specified, using working directory from the image metadata...\n"))
				Expect(outputBuffer).To(test_helpers.Say("Working directory is:\n"))
				Expect(outputBuffer).To(test_helpers.Say("/this/directory/right/here\n"))

				Expect(outputBuffer).To(test_helpers.Say("No start command specified, using start command from the image metadata...\n"))
				Expect(outputBuffer).To(test_helpers.Say("Start command is:\n"))
				Expect(outputBuffer).To(test_helpers.Say("/fetch-start arg1 arg2\n"))

				Expect(fakeDockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
				Expect(fakeDockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("fun-org/app"))

				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.StartCommand).To(Equal("/fetch-start"))
				Expect(createAppParams.AppArgs).To(Equal([]string{"arg1", "arg2"}))
				Expect(createAppParams.RootFS).To(Equal("docker:///fun-org/app#latest"))
				Expect(createAppParams.WorkingDir).To(Equal("/this/directory/right/here"))
			})

			It("does not output the working directory if it is not set", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{"/fetch-start"}}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).ToNot(test_helpers.Say("Working directory is:"))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			})

			Context("when the metadata also has no start command", func() {
				It("outputs an error message and exits", func() {
					fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say("Unable to determine start command from image metadata.\n"))
					Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.BadDocker}))
				})
			})
		})

		Context("when the timeout flag is not passed", func() {
			It("defaults the timeout to something reasonable", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{""}}, nil)
				fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)
				args := []string{
					"app-to-timeout",
					"fun-org/app",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
				Expect(createAppParams.Timeout).To(Equal(app_runner_command_factory.DefaultPollingTimeout))
			})
		})

		Describe("polling for the app to start after desiring the app", func() {
			It("polls for the app to start with correct number of instances, outputting logs while the app starts", func() {
				fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				fakeAppExaminer.RunningAppInstancesInfoReturns(0, false, nil)
				args := []string{
					"--instances=10",
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}

				doneChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

				Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("cool-web-app"))

				Expect(fakeAppExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(fakeAppExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				fakeClock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				fakeAppExaminer.RunningAppInstancesInfoReturns(9, false, nil)
				fakeClock.IncrementBySeconds(1)
				Expect(doneChan).ShouldNot(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				fakeAppExaminer.RunningAppInstancesInfoReturns(10, false, nil)
				fakeClock.IncrementBySeconds(1)

				Eventually(doneChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("cool-web-app is now running.")))
				Expect(outputBuffer).To(test_helpers.SayLine("App is reachable at:"))
				Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))

				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
			})

			Context("when the app does not start before the timeout elapses", func() {
				It("alerts the user the app took too long to start", func() {
					fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
					fakeAppExaminer.RunningAppInstancesInfoReturns(0, false, nil)
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

					Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))
					Expect(outputBuffer).To(test_helpers.SayNewLine())

					fakeClock.IncrementBySeconds(120)

					Eventually(doneChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.SayLine(colors.Red("Timed out waiting for the container to come up.")))
					Expect(outputBuffer).To(test_helpers.SayLine("This typically happens because docker layers can take time to download."))
					Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still downloading your application in the background."))
					Expect(outputBuffer).To(test_helpers.SayLine("To view logs:\n\tltc logs cool-web-app"))
					Expect(outputBuffer).To(test_helpers.SayLine("To view status:\n\tltc status cool-web-app"))
					Expect(outputBuffer).To(test_helpers.SayLine("App will be reachable at:"))
					Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))
				})
			})

			Context("when there is a placement error when polling for the app to start", func() {
				It("prints an error message and exits", func() {
					fakeDockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
					fakeAppExaminer.RunningAppInstancesInfoReturns(0, false, nil)
					args := []string{
						"--instances=10",
						"--ports=3000",
						"--working-dir=/applications",
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					doneChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

					Eventually(outputBuffer).Should(test_helpers.Say("Monitoring the app on port 3000..."))
					Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

					Expect(fakeAppExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
					Expect(fakeAppExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

					fakeClock.IncrementBySeconds(1)
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

					fakeAppExaminer.RunningAppInstancesInfoReturns(9, true, nil)
					fakeClock.IncrementBySeconds(1)
					Eventually(doneChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.SayNewLine())
					Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
					Expect(outputBuffer).ToNot(test_helpers.Say("Timed out waiting for the container"))

					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))
				})
			})
		})

		Context("invalid syntax", func() {
			It("validates the CPU weight is in 1-100", func() {
				args := []string{
					"cool-app",
					"greatapp/greatapp",
					"--cpu-weight=0",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Invalid CPU Weight"))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("validates that the name and dockerPath are passed in", func() {
				args := []string{
					"justonearg",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: APP_NAME and DOCKER_IMAGE are required"))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
			})

			It("validates that the terminator -- is passed in when a start command is specified", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"not-the-terminator",
					"start-me-up",
				}
				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: '--' Required before start command"))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
			})
		})

		Context("when the docker repo url is malformed", func() {
			It("outputs an error", func() {
				args := []string{
					"cool-web-app",
					"¥¥¥Bad-Docker¥¥¥",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Invalid repository name (¥¥¥Bad-Docker¥¥¥), only [a-z0-9-_.] are allowed"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when the app runner returns an error", func() {
			It("outputs error messages", func() {
				fakeAppRunner.CreateAppReturns(errors.New("Major Fault"))
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error creating app: Major Fault"))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when a malformed tcp routes flag is passed", func() {
			It("errors out when the container port is not an int", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--tcp-routes=woo:50000",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.InvalidRoutePortErrorMessage))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

			It("errors out when the tcp route is incomplete", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--tcp-routes=5222,50000",
					"--",
					"/start-me-please",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(app_runner_command_factory.MalformedTcpRouteErrorMessage))
				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.InvalidSyntax}))
			})

		})

		It("creates a Docker based app with tcp routes as specified in the command via the AppRunner", func() {
			args := []string{
				"--tcp-routes=5222:50000,5223:50001",
				"cool-web-app",
				"superfun/app:mycooltag",
				"--",
				"/start-me-please",
			}
			fakeAppExaminer.RunningAppInstancesInfoReturns(1, false, nil)

			test_helpers.ExecuteCommandWithArgs(createCommand, args)

			Expect(fakeDockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
			Expect(fakeDockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("superfun/app:mycooltag"))

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)

			Expect(createAppParams.Name).To(Equal("cool-web-app"))
			Expect(createAppParams.StartCommand).To(Equal("/start-me-please"))
			Expect(createAppParams.RootFS).To(Equal("docker:///superfun/app#mycooltag"))
			Expect(createAppParams.Instances).To(Equal(1))

			Expect(createAppParams.TcpRoutes).ShouldNot(BeNil())

			Expect(createAppParams.TcpRoutes).Should(ContainExactly(
				app_runner.TcpRoutes{
					app_runner.TcpRoute{
						ExternalPort: 50000,
						Port:         5222,
					},
					app_runner.TcpRoute{
						ExternalPort: 50001,
						Port:         5223,
					},
				},
			))

			Expect(createAppParams.Monitor.Method).To(Equal(app_runner.PortMonitor))

			Expect(createAppParams.NoRoutes).To(BeFalse())

			Expect(createAppParams.Setup).To(BeAssignableToTypeOf(&models.DownloadAction{}))
			reqSetup, ok := createAppParams.Setup.(*models.DownloadAction)
			Expect(ok).To(BeTrue())
			Expect(reqSetup.From).To(Equal("http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz"))
			Expect(reqSetup.To).To(Equal("/tmp"))
			Expect(reqSetup.User).To(Equal("vcap"))

			Expect(outputBuffer).To(test_helpers.SayLine("Creating App: cool-web-app"))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("cool-web-app is now running.")))
			Expect(outputBuffer).To(test_helpers.SayLine("App is reachable at:"))

			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("External TCP Port 50000 mapped to application port 5222")))
			Expect(outputBuffer).To(test_helpers.SayLine(colors.Green("External TCP Port 50001 mapped to application port 5223")))
		})
	})

})
