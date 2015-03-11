package command_factory_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher/fake_docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("CommandFactory", func() {

	var (
		appRunner                     *fake_app_runner.FakeAppRunner
		outputBuffer                  *gbytes.Buffer
		timeout                       time.Duration = 10 * time.Second
		domain                        string        = "192.168.11.11.xip.io"
		clock                         *fakeclock.FakeClock
		dockerMetadataFetcher         *fake_docker_metadata_fetcher.FakeDockerMetadataFetcher
		appRunnerCommandFactoryConfig command_factory.AppRunnerCommandFactoryConfig
		logger                        lager.Logger
		fakeTailedLogsOutputter       *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		fakeExitHandler               *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		appRunner = &fake_app_runner.FakeAppRunner{}
		outputBuffer = gbytes.NewBuffer()
		dockerMetadataFetcher = &fake_docker_metadata_fetcher.FakeDockerMetadataFetcher{}
		logger = lager.NewLogger("ltc-test")
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("CreateAppCommand", func() {

		var createCommand cli.Command

		BeforeEach(func() {
			env := []string{"SHELL=/bin/bash", "COLOR=Blue"}
			clock = fakeclock.NewFakeClock(time.Now())
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:             appRunner,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Output:                output.New(outputBuffer),
				Timeout:               timeout,
				Domain:                domain,
				Env:                   env,
				Clock:                 clock,
				Logger:                logger,
				TailedLogsOutputter:   fakeTailedLogsOutputter,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			createCommand = commandFactory.MakeCreateAppCommand()
		})

		It("creates a Docker based app as specified in the command via the AppRunner", func() {
			args := []string{
				"--memory-mb=12",
				"--disk-mb=12",
				"--ports=3000,2000,1111",
				"--monitored-port=3000",
				"--routes=3000:route-3000-yay,1111:route-1111-wahoo,1111:route-1111-me-too",
				"--working-dir=/applications",
				"--run-as-root=true",
				"--instances=22",
				"--env=TIMEZONE=CST",
				"--env=LANG=\"Chicago English\"",
				"--env=COLOR",
				"--env=UNSET",
				"cool-web-app",
				"superfun/app:mycooltag",
				"--",
				"/start-me-please",
				"AppArg0",
				"--appFlavor=\"purple\"",
			}
			appRunner.RunningAppInstancesInfoReturns(22, false, nil)

			test_helpers.ExecuteCommandWithArgs(createCommand, args)

			Expect(dockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
			Expect(dockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("superfun/app:mycooltag"))

			Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
			createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
			Expect(createDockerAppParameters.Name).To(Equal("cool-web-app"))
			Expect(createDockerAppParameters.StartCommand).To(Equal("/start-me-please"))
			Expect(createDockerAppParameters.DockerImagePath).To(Equal("superfun/app:mycooltag"))
			Expect(createDockerAppParameters.AppArgs).To(Equal([]string{"AppArg0", "--appFlavor=\"purple\""}))
			Expect(createDockerAppParameters.Instances).To(Equal(22))
			Expect(createDockerAppParameters.EnvironmentVariables).To(Equal(map[string]string{"TIMEZONE": "CST", "LANG": "\"Chicago English\"", "COLOR": "Blue", "UNSET": ""}))
			Expect(createDockerAppParameters.Privileged).To(Equal(true))
			Expect(createDockerAppParameters.MemoryMB).To(Equal(12))
			Expect(createDockerAppParameters.DiskMB).To(Equal(12))
			Expect(createDockerAppParameters.Monitor).To(Equal(true))

			Expect(createDockerAppParameters.Ports.Monitored).To(Equal(uint16(3000)))
			Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{1111, 2000, 3000}))
			Expect(createDockerAppParameters.RouteOverrides).To(ContainExactly(docker_app_runner.RouteOverrides{
				docker_app_runner.RouteOverride{HostnamePrefix: "route-3000-yay", Port: 3000},
				docker_app_runner.RouteOverride{HostnamePrefix: "route-1111-wahoo", Port: 1111},
				docker_app_runner.RouteOverride{HostnamePrefix: "route-1111-me-too", Port: 1111},
			}))
			Expect(createDockerAppParameters.WorkingDir).To(Equal("/applications"))

			Expect(outputBuffer).To(test_helpers.Say("Creating App: cool-web-app\n"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://route-3000-yay.192.168.11.11.xip.io\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://route-1111-wahoo.192.168.11.11.xip.io\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://route-1111-me-too.192.168.11.11.xip.io\n")))
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
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MalformedRouteErrorMessage))
			})

			It("errors out when there is no colon", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--routes=8888",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MalformedRouteErrorMessage))
			})
		})

		Describe("interactions with the docker metadata fetcher", func() {
			Context("when the docker image is hosted on the docker hub registry", func() {
				It("creates a Docker based app with sensible defaults and checks for metadata to know the image exists", func() {
					args := []string{
						"cool-web-app",
						"awesome/app",
						"--",
						"/start-me-please",
					}
					appRunner.RunningAppInstancesInfoReturns(1, false, nil)
					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(dockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
					Expect(dockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("awesome/app"))

					Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
					createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
					Expect(outputBuffer).To(test_helpers.Say("No port specified, image metadata did not contain exposed ports. Defaulting to 8080.\n"))
					Expect(createDockerAppParameters.Privileged).To(Equal(false))
					Expect(createDockerAppParameters.MemoryMB).To(Equal(128))
					Expect(createDockerAppParameters.DiskMB).To(Equal(1024))
					Expect(createDockerAppParameters.Ports.Monitored).To(Equal(uint16(8080)))
					Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{8080}))
					Expect(createDockerAppParameters.Instances).To(Equal(1))
					Expect(createDockerAppParameters.WorkingDir).To(Equal("/"))
				})
			})

			Context("when the docker image is hosted on a custom registry", func() {
				It("creates a docker based app with sensible defaults and checks for metadata to know the image exists", func() {
					args := []string{
						"cool-web-app",
						"super.fun.time:5000/mega-app:hallo",
						"--",
						"/start-me-please",
					}
					appRunner.RunningAppInstancesInfoReturns(1, false, nil)
					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(dockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
					Expect(dockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("super.fun.time:5000/mega-app:hallo"))

					Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
					createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
					Expect(outputBuffer).To(test_helpers.Say("No port specified, image metadata did not contain exposed ports. Defaulting to 8080.\n"))
					Expect(createDockerAppParameters.Privileged).To(Equal(false))
					Expect(createDockerAppParameters.MemoryMB).To(Equal(128))
					Expect(createDockerAppParameters.DiskMB).To(Equal(1024))
					Expect(createDockerAppParameters.Ports.Monitored).To(Equal(uint16(8080)))
					Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{8080}))
					Expect(createDockerAppParameters.Instances).To(Equal(1))
					Expect(createDockerAppParameters.WorkingDir).To(Equal("/"))
				})
			})

			Context("when the docker metadata fetcher returns an error", func() {
				It("exposes the error from trying to fetch the Docker metadata", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}
					dockerMetadataFetcher.FetchMetadataReturns(nil, errors.New("Docker Says No."))

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
					Expect(outputBuffer).To(test_helpers.Say("Error fetching image metadata: Docker Says No."))
				})
			})
		})

		Describe("exposed/monitored port behavior", func() {
			It("blows up when you pass bad port strings", func() {
				args := []string{
					"--ports=1000,98feh34",
					"--monitored-port=1000",
					"cool-web-app",
					"superfun/app:mycooltag",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.InvalidPortErrorMessage))
			})

			It("errors out when any port is > 65535 (max Linux port number)", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--ports=8080,65536",
					"--monitored-port=8080",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.InvalidPortErrorMessage))
			})

			It("is an error when the user does not pass a monitored port", func() {
				args := []string{
					"--ports=1000,1234",
					"cool-web-app",
					"superfun/app:mycooltag",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MustSetMonitoredPortErrorMessage))
			})

			It("defaults the monitored port if the user only specifies one exposed port", func() {
				args := []string{
					"--ports=1234",
					"cool-web-app",
					"superfun/app:mycooltag",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appRunner.RunningAppInstancesInfoReturns(1, false, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
				createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
				Expect(createDockerAppParameters.Ports.Monitored).To(Equal(uint16(1234)))
			})

			Context("when the --no-monitor flag is passed", func() {
				Context("when multiple ports are specified", func() {
					It("still works", func() {
						args := []string{
							"cool-web-app",
							"superfun/app",
							"--ports=8080,9090",
							"--no-monitor",
							"--",
							"/start-me-please",
						}
						appRunner.RunningAppInstancesInfoReturns(1, false, nil)
						dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

						test_helpers.ExecuteCommandWithArgs(createCommand, args)
						Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
						createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)

						Expect(createDockerAppParameters.Monitor).To(Equal(false))
						Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{8080, 9090}))
					})

				})
				Context("when the metadata does not have ports", func() {
					It("exposes 8080 but does not monitor it", func() {
						args := []string{
							"cool-web-app",
							"superfun/app",
							"--no-monitor",
							"--",
							"/start-me-please",
						}
						appRunner.RunningAppInstancesInfoReturns(1, false, nil)
						dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

						test_helpers.ExecuteCommandWithArgs(createCommand, args)
						createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)

						Expect(createDockerAppParameters.Monitor).To(Equal(false))
						Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{8080}))
					})
				})
				Context("when the docker metadata has ports", func() {
					It("exposes the ports from the metadata but does not monitor them", func() {
						args := []string{
							"cool-web-app",
							"superfun/app",
							"--no-monitor",
							"--",
							"/start-me-please",
						}
						appRunner.RunningAppInstancesInfoReturns(1, false, nil)
						dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{
							Ports: docker_app_runner.PortConfig{
								Monitored: 1200,
								Exposed:   []uint16{1200, 2701, 4302},
							}}, nil)

						test_helpers.ExecuteCommandWithArgs(createCommand, args)
						createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)

						Expect(createDockerAppParameters.Monitor).To(Equal(false))
						Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{1200, 2701, 4302}))
						Expect(outputBuffer).To(test_helpers.Say("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: 1200, 2701, 4302\n"))
						Expect(outputBuffer).To(test_helpers.Say("No ports will be monitored."))
					})
				})
			})

		})

		Context("when no working dir is provided, but the metadata has a working dir", func() {
			It("sets the working dir from the Docker metadata", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}
				appRunner.RunningAppInstancesInfoReturns(1, false, nil)
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{WorkingDir: "/work/it"}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
				Expect(createDockerAppParameters.WorkingDir).To(Equal("/work/it"))
			})
		})

		Context("when no port is provided, but the metadata has expose ports", func() {
			It("sets the ports from the Docker metadata", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}
				appRunner.RunningAppInstancesInfoReturns(1, false, nil)
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{
					Ports: docker_app_runner.PortConfig{
						Monitored: 2701,
						Exposed:   []uint16{1200, 2701, 4302},
					},
				}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
				Expect(outputBuffer).To(test_helpers.Say("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: 1200, 2701, 4302\n"))
				Expect(outputBuffer).To(test_helpers.Say("Monitoring the app on port 2701...\n"))
				Expect(createDockerAppParameters.Ports.Monitored).To(Equal(uint16(2701)))
				Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{1200, 2701, 4302}))
			})
		})

		Context("when the --no-monitor flag is passed", func() {
			Context("when the metadata does not have ports", func() {
				It("exposes 8080 but does not monitor it", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--no-monitor",
						"--",
						"/start-me-please",
					}
					appRunner.RunningAppInstancesInfoReturns(1, false, nil)
					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
					Expect(createDockerAppParameters.Monitor).To(Equal(false))
					Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{8080}))
				})
			})

			Context("when the docker metadata has ports", func() {
				It("exposes the ports from the metadata but does not monitor them", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--no-monitor",
						"--",
						"/start-me-please",
					}
					appRunner.RunningAppInstancesInfoReturns(1, false, nil)
					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{
						Ports: docker_app_runner.PortConfig{
							Monitored: 1200,
							Exposed:   []uint16{1200, 2701, 4302},
						}}, nil)

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)
					Expect(createDockerAppParameters.Monitor).To(Equal(false))
					Expect(createDockerAppParameters.Ports.Exposed).To(Equal([]uint16{1200, 2701, 4302}))
					Expect(outputBuffer).To(test_helpers.Say("No port specified, using exposed ports from the image metadata.\n\tExposed Ports: 1200, 2701, 4302\n"))
					Expect(outputBuffer).To(test_helpers.Say("No ports will be monitored."))
				})
			})
		})

		Context("when no start command is provided", func() {
			var args = []string{
				"cool-web-app",
				"fun-org/app",
			}

			JustBeforeEach(func() {
				appRunner.RunningAppInstancesInfoReturns(1, false, nil)
			})

			It("creates a Docker app with the create command retrieved from the docker image metadata", func() {
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{WorkingDir: "/this/directory/right/here", StartCommand: []string{"/fetch-start", "arg1", "arg2"}}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(dockerMetadataFetcher.FetchMetadataCallCount()).To(Equal(1))
				Expect(dockerMetadataFetcher.FetchMetadataArgsForCall(0)).To(Equal("fun-org/app"))

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
				createDockerAppParameters := appRunner.CreateDockerAppArgsForCall(0)

				Expect(createDockerAppParameters.StartCommand).To(Equal("/fetch-start"))
				Expect(createDockerAppParameters.AppArgs).To(Equal([]string{"arg1", "arg2"}))
				Expect(createDockerAppParameters.DockerImagePath).To(Equal("fun-org/app"))
				Expect(createDockerAppParameters.WorkingDir).To(Equal("/this/directory/right/here"))

				Expect(outputBuffer).To(test_helpers.Say("No working directory specified, using working directory from the image metadata...\n"))
				Expect(outputBuffer).To(test_helpers.Say("Working directory is:\n"))
				Expect(outputBuffer).To(test_helpers.Say("/this/directory/right/here\n"))

				Expect(outputBuffer).To(test_helpers.Say("No start command specified, using start command from the image metadata...\n"))
				Expect(outputBuffer).To(test_helpers.Say("Start command is:\n"))
				Expect(outputBuffer).To(test_helpers.Say("/fetch-start arg1 arg2\n"))
			})

			It("does not output the working directory if it is not set", func() {
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{"/fetch-start"}}, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).ToNot(test_helpers.Say("Working directory is:"))
			})
		})

		Context("polling for the app to start after desiring the app", func() {
			It("polls for the app to start with correct number of instances, outputting logs while the app starts", func() {
				args := []string{
					"--instances=10",
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}

				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appRunner.RunningAppInstancesInfoReturns(0, false, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

				Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("cool-web-app"))

				Expect(appRunner.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(appRunner.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				clock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				appRunner.RunningAppInstancesInfoReturns(9, false, nil)
				clock.IncrementBySeconds(1)
				Expect(commandFinishChan).ShouldNot(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				appRunner.RunningAppInstancesInfoReturns(10, false, nil)
				clock.IncrementBySeconds(1)

				Eventually(commandFinishChan).Should(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
				Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io\n")))
			})

			It("alerts the user if the app does not start", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appRunner.RunningAppInstancesInfoReturns(0, false, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

				clock.IncrementBySeconds(10)

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.Say(colors.Red("cool-web-app took too long to start.")))
			})

			Context("when there is a placement error when polling for the app to start", func() {
				It("Prints an error message and exits", func() {
					args := []string{
						"--instances=10",
						"--ports=3000",
						"--working-dir=/applications",
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}

					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
					appRunner.RunningAppInstancesInfoReturns(0, false, nil)

					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

					Eventually(outputBuffer).Should(test_helpers.Say("Monitoring the app on port 3000..."))
					Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

					Expect(appRunner.RunningAppInstancesInfoCallCount()).To(Equal(1))
					Expect(appRunner.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

					clock.IncrementBySeconds(1)
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

					appRunner.RunningAppInstancesInfoReturns(9, true, nil)
					clock.IncrementBySeconds(1)
					Eventually(commandFinishChan).Should(BeClosed())
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))

					Expect(outputBuffer).To(test_helpers.SayNewLine())
					Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
					Expect(outputBuffer).ToNot(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
					Expect(outputBuffer).ToNot(test_helpers.Say(colors.Red("cool-web-app took too long to start.")))
				})
			})
		})

		Context("invalid syntax", func() {
			It("validates that the name and dockerImage are passed in", func() {
				args := []string{
					"justonearg",
				}

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: APP_NAME and DOCKER_IMAGE are required"))
				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
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
				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
			})
		})

		Context("when the app runner returns an error", func() {
			It("outputs error messages", func() {
				args := []string{
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appRunner.CreateDockerAppReturns(errors.New("Major Fault"))

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error Creating App: Major Fault"))
			})
		})
	})

	Describe("ScaleAppCommand", func() {
		var scaleCommand cli.Command
		BeforeEach(func() {
			clock = fakeclock.NewFakeClock(time.Now())

			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:             appRunner,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Output:                output.New(outputBuffer),
				Timeout:               timeout,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
				TailedLogsOutputter:   fakeTailedLogsOutputter,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			scaleCommand = commandFactory.MakeScaleAppCommand()
		})

		It("scales an with the specified number of instances", func() {
			args := []string{
				"cool-web-app",
				"22",
			}

			appRunner.RunningAppInstancesInfoReturns(22, false, nil)

			test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("App Scaled Successfully")))

			Expect(appRunner.ScaleAppCallCount()).To(Equal(1))
			name, instances := appRunner.ScaleAppArgsForCall(0)
			Expect(name).To(Equal("cool-web-app"))
			Expect(instances).To(Equal(22))
		})

		It("polls until the required number of instances are running", func() {
			args := []string{
				"cool-web-app",
				"22",
			}

			appRunner.RunningAppInstancesInfoReturns(1, false, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))

			Expect(appRunner.RunningAppInstancesInfoCallCount()).To(Equal(1))
			Expect(appRunner.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appRunner.RunningAppInstancesInfoReturns(22, false, nil)
			clock.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("App Scaled Successfully")))
		})

		It("alerts the user if the app does not scale succesfully", func() {
			appRunner.RunningAppInstancesInfoReturns(1, false, nil)

			args := []string{
				"cool-web-app",
				"22",
			}

			dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))

			clock.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("cool-web-app took too long to scale.")))
		})

		Context("when the receptor returns errors", func() {
			It("outputs error messages", func() {
				args := []string{
					"cool-web-app",
					"22",
				}

				appRunner.ScaleAppReturns(errors.New("Major Fault"))
				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error Scaling App to 22 instances: Major Fault"))
			})
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{
					"",
				}

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'"))
				Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
			})

			It("validates that the number of instances is passed in", func() {
				args := []string{
					"cool-web-app",
				}

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc scale APP_NAME NUMBER_OF_INSTANCES'"))
				Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
			})

			It("validates that the number of instances is an integer", func() {
				args := []string{
					"cool-web-app",
					"twenty-two",
				}

				test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Number of Instances must be an integer"))
				Expect(appRunner.ScaleAppCallCount()).To(Equal(0))
			})
		})

		Context("when there is a placement error when polling for the app to scale", func() {
			It("Prints an error message and exits", func() {
				args := []string{
					"cool-web-app",
					"3",
				}

				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appRunner.RunningAppInstancesInfoReturns(0, false, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 3 instances"))

				Expect(appRunner.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(appRunner.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				clock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

				appRunner.RunningAppInstancesInfoReturns(2, true, nil)
				clock.IncrementBySeconds(1)
				Eventually(commandFinishChan).Should(BeClosed())

				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))

				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
				Expect(outputBuffer).ToNot(test_helpers.Say(colors.Green("App Scaled Successfully.")))
				Expect(outputBuffer).ToNot(test_helpers.Say(colors.Red("cool-web-app took too long to scale.")))
			})
		})
	})

	Describe("UpdateRoutesCommand", func() {
		var updateRoutesCommand cli.Command

		BeforeEach(func() {
			clock = fakeclock.NewFakeClock(time.Now())

			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:             appRunner,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Output:                output.New(outputBuffer),
				Timeout:               timeout,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			updateRoutesCommand = commandFactory.MakeUpdateRoutesCommand()
		})

		It("updates the routes", func() {
			args := []string{
				"cool-web-app",
				"8080:foo.com,9090:bar.com",
			}

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

			test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Updating cool-web-app routes. You can check this app's current routes by running 'ltc status cool-web-app'"))

			Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(1))

			name, routeOverrides := appRunner.UpdateAppRoutesArgsForCall(0)

			Expect(name).To(Equal("cool-web-app"))
			Expect(routeOverrides).To(Equal(expectedRouteOverrides))
		})

		Context("when the receptor returns errors", func() {
			It("outputs error messages", func() {
				args := []string{
					"cool-web-app",
					"8080:foo.com",
				}

				appRunner.UpdateAppRoutesReturns(errors.New("Major Fault"))
				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error updating routes: Major Fault"))
				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(1))
			})
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{
					"",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc update-routes APP_NAME NEW_ROUTES'"))
				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
			})

			It("validates that the routes are passed in", func() {
				args := []string{
					"cool-web-app",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: Please enter 'ltc update-routes APP_NAME NEW_ROUTES'"))
				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
			})
		})

		Context("malformed route", func() {
			It("errors out when the port is not an int", func() {
				args := []string{
					"cool-web-app",
					"woo:aahh",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MalformedRouteErrorMessage))

			})

			It("errors out when there is no colon", func() {
				args := []string{
					"cool-web-app",
					"8888",
				}

				test_helpers.ExecuteCommandWithArgs(updateRoutesCommand, args)

				Expect(appRunner.UpdateAppRoutesCallCount()).To(Equal(0))
				Expect(outputBuffer).To(test_helpers.Say(command_factory.MalformedRouteErrorMessage))
			})
		})

	})

	Describe("RemoveAppCommand", func() {
		var removeCommand cli.Command

		BeforeEach(func() {
			clock = fakeclock.NewFakeClock(time.Now())
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:             appRunner,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Output:                output.New(outputBuffer),
				Timeout:               timeout,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			removeCommand = commandFactory.MakeRemoveAppCommand()
		})

		It("removes a app", func() {
			args := []string{
				"cool",
			}

			appRunner.AppExistsReturns(false, nil)

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool"))
			Eventually(outputBuffer).Should(test_helpers.Say(colors.Green("Successfully Removed cool.")))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("cool"))
		})

		It("polls until the app is removed", func() {
			args := []string{
				"cool",
			}

			appRunner.AppExistsReturns(true, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool"))

			Expect(appRunner.AppExistsCallCount()).To(Equal(1))
			Expect(appRunner.AppExistsArgsForCall(0)).To(Equal("cool"))

			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appRunner.AppExistsReturns(false, nil)
			clock.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Eventually(outputBuffer).Should(test_helpers.SayNewLine())
			Eventually(outputBuffer).Should(test_helpers.Say(colors.Green("Successfully Removed cool.")))
		})

		It("alerts the user if the app does not remove", func() {
			appRunner.AppExistsReturns(true, nil)

			args := []string{
				"cool-web-app",
			}

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool-web-app"))

			clock.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.Say(colors.Red("Failed to remove cool-web-app.")))
		})

		It("alerts the user if the app runner returns an error", func() {
			appRunner.AppExistsReturns(false, errors.New("Something Bad"))

			args := []string{
				"cool-web-app",
			}

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool-web-app"))

			clock.IncrementBySeconds(10)

			Eventually(commandFinishChan).Should(BeClosed())
			Expect(outputBuffer).To(test_helpers.Say(colors.Red("Failed to remove cool-web-app.")))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
			Expect(appRunner.RemoveAppCallCount()).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.RemoveAppReturns(errors.New("Major Fault"))
			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Error Stopping App: Major Fault"))
		})
	})

})
