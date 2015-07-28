package command_factory_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/cluster_test/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/cluster_test/fake_cluster_test_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("ClusterTestCommandFactory", func() {
	var fakeClusterTestRunner *fake_cluster_test_runner.FakeClusterTestRunner

	BeforeEach(func() {
		fakeClusterTestRunner = &fake_cluster_test_runner.FakeClusterTestRunner{}
	})

	Describe("MakeClusterTestCommand", func() {
		var clusterTestCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewClusterTestCommandFactory(fakeClusterTestRunner)
			clusterTestCommand = commandFactory.MakeClusterTestCommand()
		})

		It("prints the integration test run output and args", func() {
			test_helpers.ExecuteCommandWithArgs(clusterTestCommand, []string{"--timeout=50s", "--verbose=true"})

			Expect(fakeClusterTestRunner.RunCallCount()).To(Equal(1))
			timeoutArg, verboseArg := fakeClusterTestRunner.GetArgsForRun()
			Expect(timeoutArg).To(Equal(time.Second * 50))
			Expect(verboseArg).To(BeTrue())
		})

		It("has sane defaults", func() {
			test_helpers.ExecuteCommandWithArgs(clusterTestCommand, []string{})

			Expect(fakeClusterTestRunner.RunCallCount()).To(Equal(1))
			timeoutArg, verboseArg := fakeClusterTestRunner.GetArgsForRun()
			Expect(timeoutArg).To(Equal(time.Minute * 2))
			Expect(verboseArg).To(BeFalse())
		})
	})
})
