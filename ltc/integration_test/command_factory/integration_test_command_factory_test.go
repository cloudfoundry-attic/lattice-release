package command_factory_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/integration_test/fake_integration_test_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("IntegrationTestCommandFactory", func() {
	var (
		fakeIntegrationTestRunner *fake_integration_test_runner.FakeIntegrationTestRunner
		integrationTestCommand    cli.Command
		outputBuffer              *gbytes.Buffer
	)

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
		fakeIntegrationTestRunner = fake_integration_test_runner.NewFakeIntegrationTestRunner(output.New(outputBuffer))
		integrationTestCommand = command_factory.NewIntegrationTestCommandFactory(fakeIntegrationTestRunner, output.New(outputBuffer)).MakeIntegrationTestCommand()
	})

	Describe("MakeIntegrationTestCommand", func() {

		It("prints the integration test run output and args", func() {
			test_helpers.ExecuteCommandWithArgs(integrationTestCommand, []string{"--timeout=50s", "--verbose=true"})

			timeoutArg, verboseArg := fakeIntegrationTestRunner.GetArgsForRun()

			Expect(timeoutArg).To(Equal(time.Second * 50))
			Expect(verboseArg).To(Equal(true))

			Expect(outputBuffer).To(test_helpers.Say("Running fake integration tests!!!\n"))
		})

		It("has sane defaults", func() {
			test_helpers.ExecuteCommandWithArgs(integrationTestCommand, []string{})

			timeoutArg, verboseArg := fakeIntegrationTestRunner.GetArgsForRun()

			Expect(timeoutArg).To(Equal(time.Second * 30))
			Expect(verboseArg).To(Equal(false))

			Expect(outputBuffer).To(test_helpers.Say("Running fake integration tests!!!\n"))
		})

	})
})
