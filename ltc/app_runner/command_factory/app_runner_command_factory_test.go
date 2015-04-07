package command_factory_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_metadata_fetcher/fake_docker_metadata_fetcher"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("CommandFactory", func() {

	var (
		appRunner                     *fake_app_runner.FakeAppRunner
		appExaminer                   *fake_app_examiner.FakeAppExaminer
		outputBuffer                  *gbytes.Buffer
		terminalUI                    terminal.UI
		domain                        string = "192.168.11.11.xip.io"
		clock                         *fakeclock.FakeClock
		dockerMetadataFetcher         *fake_docker_metadata_fetcher.FakeDockerMetadataFetcher
		appRunnerCommandFactoryConfig command_factory.AppRunnerCommandFactoryConfig
		logger                        lager.Logger
		fakeTailedLogsOutputter       *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		fakeExitHandler               *fake_exit_handler.FakeExitHandler
	)

	BeforeEach(func() {
		appRunner = &fake_app_runner.FakeAppRunner{}
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		dockerMetadataFetcher = &fake_docker_metadata_fetcher.FakeDockerMetadataFetcher{}
		clock = fakeclock.NewFakeClock(time.Now())
		logger = lager.NewLogger("ltc-test")
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("CreateAppCommand", func() {
		var createCommand cli.Command

		BeforeEach(func() {
			env := []string{"SHELL=/bin/bash", "COLOR=Blue"}
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
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
				"--cpu-weight=57",
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
				"--timeout=28s",
				"cool-web-app",
				"superfun/app:mycooltag",
				"--",
				"/start-me-please",
				"AppArg0",
				"--appFlavor=\"purple\"",
			}
			appExaminer.RunningAppInstancesInfoReturns(22, false, nil)

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
			Expect(createDockerAppParameters.EnvironmentVariables).To(Equal(map[string]string{"TIMEZONE": "CST", "LANG": "\"Chicago English\"", "PROCESS_GUID": "cool-web-app", "COLOR": "Blue", "UNSET": ""}))
			Expect(createDockerAppParameters.Privileged).To(Equal(true))
			Expect(createDockerAppParameters.CPUWeight).To(Equal(uint(57)))
			Expect(createDockerAppParameters.MemoryMB).To(Equal(12))
			Expect(createDockerAppParameters.DiskMB).To(Equal(12))
			Expect(createDockerAppParameters.Monitor).To(Equal(true))
			Expect(createDockerAppParameters.Timeout).To(Equal(time.Second * 28))

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
			Expect(outputBuffer).To(test_helpers.Say("App is reachable at:\n"))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://route-3000-yay.192.168.11.11.xip.io\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://route-1111-wahoo.192.168.11.11.xip.io\n")))
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://route-1111-me-too.192.168.11.11.xip.io\n")))
		})

		Context("when the PROCESS_GUID is passed in as --env", func() {
			It("sets the PROCESS_GUID to the value passed in", func() {
				args := []string{
					"app-to-start",
					"fun-org/app",
					"--env=PROCESS_GUID=MyHappyGuid",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{""}}, nil)
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
				createDockerAppParams := appRunner.CreateDockerAppArgsForCall(0)
				appEnvVars := createDockerAppParams.EnvironmentVariables
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
					appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
					appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)

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
						appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
						appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
						appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
					appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
					appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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

			BeforeEach(func() {
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
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
				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
			})

			Context("when the metadata also has no start command", func() {
				It("outputs an error message and exits", func() {
					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

					test_helpers.ExecuteCommandWithArgs(createCommand, args)

					Expect(outputBuffer).To(test_helpers.Say("Unable to determine start command from image metadata.\n"))
					Expect(appRunner.CreateDockerAppCallCount()).To(BeZero())
				})
			})
		})

		Context("when the timeout flag is not passed", func() {
			It("defaults the timeout to something reasonable", func() {
				args := []string{
					"app-to-timeout",
					"fun-org/app",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{StartCommand: []string{""}}, nil)
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)

				test_helpers.ExecuteCommandWithArgs(createCommand, args)

				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(1))
				createDockerAppParams := appRunner.CreateDockerAppArgsForCall(0)
				Expect(createDockerAppParams.Timeout).To(Equal(command_factory.DefaultPollingTimeout))
			})
		})

		Describe("polling for the app to start after desiring the app", func() {
			It("polls for the app to start with correct number of instances, outputting logs while the app starts", func() {
				args := []string{
					"--instances=10",
					"cool-web-app",
					"superfun/app",
					"--",
					"/start-me-please",
				}

				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
				appExaminer.RunningAppInstancesInfoReturns(0, false, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

				Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("cool-web-app"))

				Expect(appExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(appExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				clock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				appExaminer.RunningAppInstancesInfoReturns(9, false, nil)
				clock.IncrementBySeconds(1)
				Expect(commandFinishChan).ShouldNot(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))

				appExaminer.RunningAppInstancesInfoReturns(10, false, nil)
				clock.IncrementBySeconds(1)

				Eventually(commandFinishChan).Should(BeClosed())
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.Say(colors.Green("cool-web-app is now running.\n")))
				Expect(outputBuffer).To(test_helpers.Say("App is reachable at:\n"))
				Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io\n")))
			})

			Context("when the app does not start before the timeout elapses", func() {
				It("alerts the user the app took too long to start", func() {
					args := []string{
						"cool-web-app",
						"superfun/app",
						"--",
						"/start-me-please",
					}
					dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)
					appExaminer.RunningAppInstancesInfoReturns(0, false, nil)

					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

					Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))
					Expect(outputBuffer).To(test_helpers.SayNewLine())

					clock.IncrementBySeconds(120)

					Eventually(commandFinishChan).Should(BeClosed())

					Expect(outputBuffer).To(test_helpers.Say(colors.Red("Timed out waiting for the container to come up.")))
					Expect(outputBuffer).To(test_helpers.SayNewLine())
					Expect(outputBuffer).To(test_helpers.SayLine("This typically happens because docker layers can take time to download."))
					Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still downloading your application in the background."))
					Expect(outputBuffer).To(test_helpers.SayLine("To view logs:\n\tltc logs cool-web-app"))
					Expect(outputBuffer).To(test_helpers.SayLine("To view status:\n\tltc status cool-web-app"))
					Expect(outputBuffer).To(test_helpers.Say("App will be reachable at:\n"))
					Expect(outputBuffer).To(test_helpers.Say(colors.Green("http://cool-web-app.192.168.11.11.xip.io\n")))
				})
			})

			Context("when there is a placement error when polling for the app to start", func() {
				It("prints an error message and exits", func() {
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
					appExaminer.RunningAppInstancesInfoReturns(0, false, nil)

					commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(createCommand, args)

					Eventually(outputBuffer).Should(test_helpers.Say("Monitoring the app on port 3000..."))
					Eventually(outputBuffer).Should(test_helpers.Say("Creating App: cool-web-app"))

					Expect(appExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
					Expect(appExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

					clock.IncrementBySeconds(1)
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
					Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

					appExaminer.RunningAppInstancesInfoReturns(9, true, nil)
					clock.IncrementBySeconds(1)
					Eventually(commandFinishChan).Should(BeClosed())
					Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))
					Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(1))

					Expect(outputBuffer).To(test_helpers.SayNewLine())
					Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
					Expect(outputBuffer).ToNot(test_helpers.Say("Timed out waiting for the container"))
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
				Expect(appRunner.CreateDockerAppCallCount()).To(Equal(0))
			})

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

				Expect(outputBuffer).To(test_helpers.Say("Error creating app: Major Fault"))
			})
		})
	})

	Describe("CreateLrpCommand", func() {
		var (
			createLrpCommand cli.Command
			tmpDir           string
			tmpFile          *os.File
			err              error
		)

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
				TailedLogsOutputter:   fakeTailedLogsOutputter,
				ExitHandler:           fakeExitHandler,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			createLrpCommand = commandFactory.MakeCreateLrpCommand()
		})

		Context("when the json file exists", func() {
			BeforeEach(func() {
				tmpDir = os.TempDir()
				tmpFile, err = ioutil.TempFile(tmpDir, "tmp_json")

				Expect(err).ToNot(HaveOccurred())
			})

			It("creates an app from json", func() {
				ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value"}`), 0700)
				args := []string{tmpFile.Name()}

				appRunner.CreateLrpReturns("my-json-app", nil)

				test_helpers.ExecuteCommandWithArgs(createLrpCommand, args)

				Expect(appRunner.CreateLrpCallCount()).To(Equal(1))
				Expect(appRunner.CreateLrpArgsForCall(0)).To(Equal([]byte(`{"Value":"test value"}`)))
				Expect(outputBuffer).To(test_helpers.Say(colors.Green("Successfully submitted my-json-app.")))
				Expect(outputBuffer).To(test_helpers.Say("To view the status of your application: ltc status my-json-app"))

			})

			It("prints an error returned by the app_runner", func() {
				args := []string{
					tmpFile.Name(),
				}
				appRunner.CreateLrpReturns("app-that-broke", errors.New("some error"))

				test_helpers.ExecuteCommandWithArgs(createLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error creating app-that-broke: some error"))
				Expect(appRunner.CreateLrpCallCount()).To(Equal(1))
			})
		})

		It("is an error when no path is passed in", func() {
			test_helpers.ExecuteCommandWithArgs(createLrpCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Path to JSON is required"))
			Expect(appRunner.CreateLrpCallCount()).To(BeZero())
		})

		Context("when the file cannot be read", func() {
			It("prints an error", func() {
				args := []string{filepath.Join(tmpDir, "file-no-existy")}

				test_helpers.ExecuteCommandWithArgs(createLrpCommand, args)

				Expect(outputBuffer).To(test_helpers.Say(fmt.Sprintf("Error reading file: open %s: no such file or directory", filepath.Join(tmpDir, "file-no-existy"))))
				Expect(appRunner.CreateLrpCallCount()).To(Equal(0))
			})
		})
	})

	Describe("ScaleAppCommand", func() {

		var scaleCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
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

			appExaminer.RunningAppInstancesInfoReturns(22, false, nil)

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

			appExaminer.RunningAppInstancesInfoReturns(1, false, nil)

			commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))

			Expect(appExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
			Expect(appExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))
			clock.IncrementBySeconds(1)
			Eventually(outputBuffer).Should(test_helpers.Say("."))

			appExaminer.RunningAppInstancesInfoReturns(22, false, nil)
			clock.IncrementBySeconds(1)

			Eventually(commandFinishChan).Should(BeClosed())

			Expect(outputBuffer).To(test_helpers.SayNewLine())
			Expect(outputBuffer).To(test_helpers.Say(colors.Green("App Scaled Successfully")))
		})

		Context("when the app does not scale before the timeout elapses", func() {
			It("alerts the user the app took too long to scale", func() {
				appExaminer.RunningAppInstancesInfoReturns(1, false, nil)
				args := []string{
					"cool-web-app",
					"22",
				}
				dockerMetadataFetcher.FetchMetadataReturns(&docker_metadata_fetcher.ImageMetadata{}, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 22 instances"))
				Expect(outputBuffer).To(test_helpers.SayNewLine())

				clock.IncrementBySeconds(120)

				Eventually(commandFinishChan).Should(BeClosed())

				Expect(outputBuffer).To(test_helpers.Say(colors.Red("Timed out waiting for the container to scale.")))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.SayLine("Lattice is still scaling your application in the background."))
				Expect(outputBuffer).To(test_helpers.SayLine("To view logs:\n\tltc logs cool-web-app"))
				Expect(outputBuffer).To(test_helpers.SayLine("To view status:\n\tltc status cool-web-app"))
				Expect(outputBuffer).To(test_helpers.SayNewLine())
			})
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
				appExaminer.RunningAppInstancesInfoReturns(0, false, nil)

				commandFinishChan := test_helpers.AsyncExecuteCommandWithArgs(scaleCommand, args)

				Eventually(outputBuffer).Should(test_helpers.Say("Scaling cool-web-app to 3 instances"))

				Expect(appExaminer.RunningAppInstancesInfoCallCount()).To(Equal(1))
				Expect(appExaminer.RunningAppInstancesInfoArgsForCall(0)).To(Equal("cool-web-app"))

				clock.IncrementBySeconds(1)
				Expect(fakeTailedLogsOutputter.StopOutputtingCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(BeEmpty())

				appExaminer.RunningAppInstancesInfoReturns(2, true, nil)
				clock.IncrementBySeconds(1)
				Eventually(commandFinishChan).Should(BeClosed())

				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.PlacementError}))

				Expect(outputBuffer).To(test_helpers.SayNewLine())
				Expect(outputBuffer).To(test_helpers.Say(colors.Red("Error, could not place all instances: insufficient resources. Try requesting fewer instances or reducing the requested memory or disk capacity.")))
				Expect(outputBuffer).ToNot(test_helpers.Say("Timed out waiting for the container"))
			})
		})
	})

	Describe("UpdateRoutesCommand", func() {
		var updateRoutesCommand cli.Command

		BeforeEach(func() {
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
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
			appRunnerCommandFactoryConfig = command_factory.AppRunnerCommandFactoryConfig{
				AppRunner:   appRunner,
				AppExaminer: appExaminer,
				UI:          terminalUI,
				DockerMetadataFetcher: dockerMetadataFetcher,
				Domain:                domain,
				Env:                   []string{},
				Clock:                 clock,
				Logger:                logger,
			}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunnerCommandFactoryConfig)
			removeCommand = commandFactory.MakeRemoveAppCommand()
		})

		It("removes an app", func() {
			args := []string{
				"cool",
			}

			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.Say("Removing cool"))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("cool"))
		})


		It("removes multiple apps", func(){

			args := []string{
				"app1",
				"app2",
				"app3",
			}
			
			test_helpers.ExecuteCommandWithArgs(removeCommand, args)

			Eventually(outputBuffer).Should(test_helpers.SayLine("Removing app1..."))
			Eventually(outputBuffer).Should(test_helpers.SayLine("Removing app2..."))
			Eventually(outputBuffer).Should(test_helpers.SayLine("Removing app3..."))

			Expect(appRunner.RemoveAppCallCount()).To(Equal(3))
			Expect(appRunner.RemoveAppArgsForCall(0)).To(Equal("app1"))
			Expect(appRunner.RemoveAppArgsForCall(1)).To(Equal("app2"))
			Expect(appRunner.RemoveAppArgsForCall(2)).To(Equal("app3"))
		})

		Context("invalid syntax", func() {
			It("validates that the name is passed in", func() {
				args := []string{}

				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: App Name required"))
				Expect(appRunner.RemoveAppCallCount()).To(Equal(0))
			})
		})

		Context("when the receptor returns an error", func() {
			It("outputs error messages when trying to remove the app", func() {
				args := []string{
					"cool-web-app",
				}
				appRunner.RemoveAppReturns(errors.New("Major Fault"))

				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.Say("Error stopping cool-web-app: Major Fault"))
				Expect(appRunner.RemoveAppCallCount()).To(Equal(1))
			})

			It("outputs error messages when trying to remove the app", func() {
				args := []string{
					"app1",
					"app2",
					"app3",
				}

				appRunner.RemoveAppStub = func(name string) error {
					if name == "app2" {
						return errors.New("Major Fault")
					}
					return nil
				}

				test_helpers.ExecuteCommandWithArgs(removeCommand, args)

				Expect(outputBuffer).To(test_helpers.SayLine("Removing app1..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Removing app2..."))
				Expect(outputBuffer).To(test_helpers.SayLine("Error stopping app2: Major Fault"))
				Expect(outputBuffer).To(test_helpers.SayLine("Removing app3..."))

				Expect(appRunner.RemoveAppCallCount()).To(Equal(3))
			})
		})

	})
})
