package command_factory_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test/fake_integration_test_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("IntegrationTestCommandFactory", func() {
	var (
		fakeIntegrationTestRunner *fake_integration_test_runner.FakeIntegrationTestRunner
	)

	BeforeEach(func() {
		fakeIntegrationTestRunner = fake_integration_test_runner.NewFakeIntegrationTestRunner()
	})

	Describe("MakeIntegrationTestCommand", func() {

		var integrationTestCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewIntegrationTestCommandFactory(fakeIntegrationTestRunner)
			integrationTestCommand = commandFactory.MakeIntegrationTestCommand()
		})

		It("prints the integration test run output and args", func() {
			test_helpers.ExecuteCommandWithArgs(integrationTestCommand, []string{"--timeout=50s", "--verbose=true", "--cli-help"})

			Expect(fakeIntegrationTestRunner.RunCallCount()).To(Equal(1))
			timeoutArg, verboseArg, cliHelpArg := fakeIntegrationTestRunner.GetArgsForRun()
			Expect(timeoutArg).To(Equal(time.Second * 50))
			Expect(verboseArg).To(Equal(true))
			Expect(cliHelpArg).To(Equal(true))
		})

		It("has sane defaults", func() {
			test_helpers.ExecuteCommandWithArgs(integrationTestCommand, []string{})

			Expect(fakeIntegrationTestRunner.RunCallCount()).To(Equal(1))
			timeoutArg, verboseArg, cliHelpArg := fakeIntegrationTestRunner.GetArgsForRun()
			Expect(timeoutArg).To(Equal(time.Minute * 2))
			Expect(verboseArg).To(BeFalse())
			Expect(cliHelpArg).To(BeFalse())
		})

	})
})
