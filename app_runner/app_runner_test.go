package app_runner_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-cf-experimental/diego-edge-cli/app_runner"
)

var _ = Describe("AppRunner", func() {
	fakeReceptorClient := &fake_receptor.FakeClient{}
	appRunner := app_runner.NewDiegoAppRunner(fakeReceptorClient)

	Describe("StartDockerApp", func() {
		It("Starts a Docker App", func() {
			err := appRunner.StartDockerApp("americano-app", "/app-run-statement", "docker://runtest/runner")
			Expect(err).To(BeNil())

			Expect(fakeReceptorClient.CreateDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.CreateDesiredLRPArgsForCall(0)).To(Equal(receptor.DesiredLRPCreateRequest{
				ProcessGuid: "americano-app",
				Domain:      "diego-edge",
				RootFSPath:  "docker://runtest/runner",
				Instances:   1,
				Stack:       "lucid64",
				Routes:      []string{"americano-app.192.168.11.11.xip.io"},
				MemoryMB:    128,
				DiskMB:      1024,
				Ports: []receptor.PortMapping{
					{ContainerPort: 8080},
				},
				LogGuid:   "americano-app",
				LogSource: "APP",
				Actions: []models.ExecutorAction{
					{
						Action: models.DownloadAction{
							From:     "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz",
							To:       "/tmp",
							CacheKey: "",
						},
					},
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
										Path: "/tmp/spy",
										Args: []string{"-addr", ":8080"},
									},
								},
								HealthyThreshold:   1,
								UnhealthyThreshold: 1,
								HealthyHook: models.HealthRequest{
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
			receptorClientError := errors.New("Something went really wrong")
			fakeReceptorClient.CreateDesiredLRPReturns(receptorClientError)

			err := appRunner.StartDockerApp("nescafe-app", "/app-bork-statement", "docker://faily/boom")
			Expect(err).To(Equal(receptorClientError))
		})
	})
})
