package app_runner_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-cf-experimental/diego-edge-cli/app_runner"
)

var _ = Describe("AppRunner", func() {
	fakeBbs := &fake_bbs.FakeNsyncBBS{}
	appRunner := app_runner.NewDiegoAppRunner(fakeBbs)

	Describe("StartDockerApp", func() {
		It("Starts a Docker App", func() {
			err := appRunner.StartDockerApp("americano-app", "/app-run-statement", "docker://runtest/runner")
			Expect(err).To(BeNil())

			Expect(fakeBbs.DesireLRPCallCount()).To(Equal(1))
			Expect(fakeBbs.DesireLRPArgsForCall(0)).To(Equal(models.DesiredLRP{
				Domain:      "diego-edge",
				ProcessGuid: "americano-app",
				Instances:   1,
				Stack:       "lucid64",
				RootFSPath:  "docker://runtest/runner",
				Routes:      []string{"americano-app.192.168.11.11.xip.io"},
				MemoryMB:    128,
				DiskMB:      1024,
				Ports: []models.PortMapping{
					{ContainerPort: 8080},
				},
				Log: models.LogConfig{
					Guid:       "americano-app",
					SourceName: "APP",
				},
				Actions: []models.ExecutorAction{
					models.Parallel(
						models.ExecutorAction{
							models.RunAction{
								Path: "/app-run-statement",
							},
						},
						models.ExecutorAction{
							models.MonitorAction{
								Action: models.ExecutorAction{
									models.RunAction{
										Path: "echo",
										Args: []string{"I'm a healthy little spy"},
									},
								},
								HealthyThreshold:   1,
								UnhealthyThreshold: 1,
								HealthyHook: models.HealthRequest{ //Teel the rep where to call back to on exit 0 of spy
									Method: "PUT",
									URL:    "http://127.0.0.1:20515/lrp_running/americano-app/PLACEHOLDER_INSTANCE_INDEX/PLACEHOLDER_INSTANCE_GUID",
								},
							},
						},
					),
				},
			}))
		})

		It("returns errors from the bbs", func() {
			bbsError := errors.New("Something went really wrong")
			fakeBbs.DesireLRPReturns(bbsError)

			err := appRunner.StartDockerApp("nescafe-app", "/app-bork-statement", "docker://faily/boom")
			Expect(err).To(Equal(bbsError))
		})
	})
})
