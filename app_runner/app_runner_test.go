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

	Describe("StartDockerApp", func() {
		var (
			fakeReceptorClient *fake_receptor.FakeClient
			appRunner          *app_runner.DiegoAppRunner
		)

		BeforeEach(func() {
			fakeReceptorClient = &fake_receptor.FakeClient{}
			appRunner = app_runner.NewDiegoAppRunner(fakeReceptorClient)

		})

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
				Ports:       []uint32{8080},
				LogGuid:     "americano-app",
				LogSource:   "APP",
				Setup: &models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz",
					To:   "/tmp",
				},
				Action: &models.RunAction{
					Path: "/app-run-statement",
				},
				Monitor: &models.RunAction{
					Path: "/tmp/spy",
					Args: []string{"-addr", ":8080"},
				},
			}))
		})

		It("returns errors if the app is already started", func() {
			existingLRPResponse := receptor.ActualLRPResponse{ProcessGuid: "app-all-ready-running"}
			fakeReceptorClient.ActualLRPsByProcessGuidReturns([]receptor.ActualLRPResponse{existingLRPResponse}, nil)

			err := appRunner.StartDockerApp("app-all-ready-running", "/app-bork-statement", "docker://faily/boom")

			Expect(err.Error()).To(Equal("App app-all-ready-running, is already running"))
			Expect(fakeReceptorClient.ActualLRPsByProcessGuidCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.ActualLRPsByProcessGuidArgsForCall(0)).To(Equal("app-all-ready-running"))
		})

		Describe("returning errors from the receptor", func() {
			It("returns desiring lrp errors", func() {
				receptorError := errors.New("error - Desiring an LRP")
				fakeReceptorClient.CreateDesiredLRPReturns(receptorError)

				err := appRunner.StartDockerApp("nescafe-app", "/app-bork-statement", "docker://faily/boom")
				Expect(err).To(Equal(receptorError))
			})

			It("returns existing count errors", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.ActualLRPsByProcessGuidReturns([]receptor.ActualLRPResponse{}, receptorError)

				err := appRunner.StartDockerApp("nescafe-app", "/app-bork-statement", "docker://faily/boom")
				Expect(err).To(Equal(receptorError))
			})
		})

	})
})
