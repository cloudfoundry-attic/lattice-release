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

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		appRunner          *app_runner.DiegoAppRunner
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		appRunner = app_runner.NewDiegoAppRunner(fakeReceptorClient)

	})

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

		It("returns errors if the app is already desired", func() {
			desiredLRPResponse := receptor.DesiredLRPResponse{ProcessGuid: "app-already-desired", Instances: 1}
			fakeReceptorClient.GetDesiredLRPReturns(desiredLRPResponse, nil)

			err := appRunner.StartDockerApp("app-already-desired", "/app-bork-statement", "docker://faily/boom")

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("App app-already-desired, is already running"))
			Expect(fakeReceptorClient.GetDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.GetDesiredLRPArgsForCall(0)).To(Equal("app-already-desired"))
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
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.StartDockerApp("nescafe-app", "/app-bork-statement", "docker://faily/boom")
				Expect(err).To(Equal(receptorError))
			})
		})

	})

	Describe("ScaleDockerApp", func() {

		It("Scales a Docker App", func() {
			desiredLRPResponse := receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}
			fakeReceptorClient.GetDesiredLRPReturns(desiredLRPResponse, nil)
			instanceCount := 25

			err := appRunner.ScaleDockerApp("americano-app", instanceCount)
			Expect(err).To(BeNil())

			Expect(fakeReceptorClient.UpdateDesiredLRPCallCount()).To(Equal(1))
			processGuid, updateRequest := fakeReceptorClient.UpdateDesiredLRPArgsForCall(0)
			Expect(processGuid).To(Equal("americano-app"))

			Expect(updateRequest).To(Equal(receptor.DesiredLRPUpdateRequest{Instances: &instanceCount}))
		})

		It("returns errors if the app is NOT already started", func() {
			fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, errors.New("LRP not found"))

			err := appRunner.ScaleDockerApp("app-not-running", 15)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("app-not-running, is not started. Please start an app first"))
			Expect(fakeReceptorClient.GetDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.GetDesiredLRPArgsForCall(0)).To(Equal("app-not-running"))
		})

		Describe("returning errors from the receptor", func() {
			It("returns desiring lrp errors", func() {
				desiredLRPResponse := receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}
				fakeReceptorClient.GetDesiredLRPReturns(desiredLRPResponse, nil)

				receptorError := errors.New("error - Updating an LRP")
				fakeReceptorClient.UpdateDesiredLRPReturns(receptorError)

				err := appRunner.ScaleDockerApp("nescafe-app", 17)
				Expect(err).To(Equal(receptorError))
			})

			It("returns errors fetching the ® count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.ScaleDockerApp("nescafe-app", 2)
				Expect(err).To(Equal(receptorError))
			})
		})

	})

	Describe("StopDockerApp", func() {
		It("Stops a Docker App", func() {
			desiredLRPResponse := receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}
			fakeReceptorClient.GetDesiredLRPReturns(desiredLRPResponse, nil)

			fakeReceptorClient.DeleteDesiredLRPReturns(nil)

			err := appRunner.StopDockerApp("americano-app")
			Expect(err).To(BeNil())

			Expect(fakeReceptorClient.DeleteDesiredLRPCallCount()).To(Equal(1))

			Expect(fakeReceptorClient.DeleteDesiredLRPArgsForCall(0)).To(Equal("americano-app"))

		})

		It("returns errors if the app is NOT already started", func() {
			fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, errors.New("LRP not found"))

			err := appRunner.StopDockerApp("app-not-running")

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("app-not-running, is not started. Please start an app first"))
			Expect(fakeReceptorClient.GetDesiredLRPCallCount()).To(Equal(1))
			Expect(fakeReceptorClient.GetDesiredLRPArgsForCall(0)).To(Equal("app-not-running"))

		})

		Describe("returning errors from the receptor", func() {
			It("returns deleting lrp errors", func() {
				desiredLRPResponse := receptor.DesiredLRPResponse{ProcessGuid: "americano-app", Instances: 1}
				fakeReceptorClient.GetDesiredLRPReturns(desiredLRPResponse, nil)

				deletingError := errors.New("deleting failed")
				fakeReceptorClient.DeleteDesiredLRPReturns(deletingError)

				err := appRunner.StopDockerApp("nescafe-app")

				Expect(err).To(Equal(deletingError))
			})

			It("returns errors fetching the existing count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{}, receptorError)

				err := appRunner.StopDockerApp("nescafe-app")
				Expect(err).To(Equal(receptorError))
			})
		})

	})

	Describe("IsDockerAppUp", func() {
		It("returns true if the docker app is running with a status of ActualLRPStateRunning", func() {
			actualLrpsResponse := []receptor.ActualLRPResponse{receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateRunning}}
			fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLrpsResponse, nil)

			status, err := appRunner.IsDockerAppUp("americano-app")
			Expect(err).To(BeNil())
			Expect(status).To(BeTrue())

			Expect(fakeReceptorClient.ActualLRPsByProcessGuidCallCount()).To(Equal(1))

			Expect(fakeReceptorClient.ActualLRPsByProcessGuidArgsForCall(0)).To(Equal("americano-app"))
		})

		It("returns false if the docker app is running without a status of ActualLRPStateRunning", func() {
			actualLrpsResponse := []receptor.ActualLRPResponse{receptor.ActualLRPResponse{ProcessGuid: "americano-app", State: receptor.ActualLRPStateStarting}}
			fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLrpsResponse, nil)

			status, _ := appRunner.IsDockerAppUp("americano-app")
			Expect(status).To(BeFalse())

		})

		It("returns false if the docker app is not running", func() {
			actualLrpsResponse := []receptor.ActualLRPResponse{}
			fakeReceptorClient.ActualLRPsByProcessGuidReturns(actualLrpsResponse, nil)

			status, _ := appRunner.IsDockerAppUp("americano-app")
			Expect(status).To(BeFalse())

		})

		Describe("returning errors from the receptor", func() {
			It("returns errors fetching the existing count", func() {
				receptorError := errors.New("error - Existing Count")
				fakeReceptorClient.ActualLRPsByProcessGuidReturns([]receptor.ActualLRPResponse{}, receptorError)
				_, err := appRunner.IsDockerAppUp("nescafe-app")
				Expect(err).To(Equal(receptorError))
			})
		})
	})
})
