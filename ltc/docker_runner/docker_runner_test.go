package docker_runner_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/docker_runner"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("DockerRunner", func() {

	var (
		fakeReceptorClient *fake_receptor.FakeClient
		dockerAppRunner    docker_runner.DockerRunner
		fakeAppRunner      *fake_app_runner.FakeAppRunner
	)

	BeforeEach(func() {
		fakeReceptorClient = &fake_receptor.FakeClient{}
		fakeAppRunner = &fake_app_runner.FakeAppRunner{}
		dockerAppRunner = docker_runner.New(fakeAppRunner)
	})

	Describe("CreateDockerApp", func() {
		It("provides the correct app param callbacks", func() {
			err := dockerAppRunner.CreateDockerApp(app_runner.CreateAppParams{
				RootFS: "runtest/runner",
			})

			Expect(err).ToNot(HaveOccurred())

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams.GetRootFS()).To(Equal("docker:///runtest/runner#latest"))
			Expect(createAppParams.GetSetupAction()).To(Equal(&models.DownloadAction{
				From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
				To:   "/tmp",
			}))
		})

		Context("when the docker repo url is malformed", func() {
			It("returns an error", func() {
				err := dockerAppRunner.CreateDockerApp(app_runner.CreateAppParams{
					Name:         "nescafe-app",
					StartCommand: "/app",
					RootFS:       "¥¥¥Bad-Docker¥¥¥",
				})

				Expect(err).ToNot(HaveOccurred())

				Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
				createAppParams := fakeAppRunner.CreateAppArgsForCall(0)

				_, err = createAppParams.GetRootFS()
				Expect(err).To(MatchError("Invalid repository name (¥¥¥Bad-Docker¥¥¥), only [a-z0-9-_.] are allowed"))
			})
		})

		Context("when the real app runner returns an error", func() {
			It("returns the error", func() {
				fakeAppRunner.CreateAppReturns(errors.New("boom"))

				err := dockerAppRunner.CreateDockerApp(app_runner.CreateAppParams{
					RootFS: "runtest/runner",
				})

				Expect(err).To(MatchError("boom"))
			})
		})
	})

})
