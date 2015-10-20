package command_factory_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/lattice/ltc/autoupdate/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/autoupdate/command_factory/mocks"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Sync CommandFactory", func() {
	var (
		config          *config_package.Config
		outputBuffer    *gbytes.Buffer
		terminalUI      terminal.UI
		fakeExitHandler *fake_exit_handler.FakeExitHandler
		fakeSync        *mocks.FakeSync
	)

	BeforeEach(func() {
		config = config_package.New(nil)
		config.SetTarget("lattice.xip.io")

		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeExitHandler = &fake_exit_handler.FakeExitHandler{}
		fakeSync = &mocks.FakeSync{}
	})

	Describe("SyncCommand", func() {
		var syncCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewSyncCommandFactory(config, terminalUI, fakeExitHandler, "darwin", "/fake/ltc", fakeSync)
			syncCommand = commandFactory.MakeSyncCommand()
		})

		It("should sync ltc", func() {
			test_helpers.ExecuteCommandWithArgs(syncCommand, []string{})

			Expect(outputBuffer).To(test_helpers.SayLine("Updated ltc to the latest version."))
			Expect(fakeSync.SyncLTCCallCount()).To(Equal(1))
			actualLTCPath, actualArch, actualConfig := fakeSync.SyncLTCArgsForCall(0)
			Expect(actualLTCPath).To(Equal("/fake/ltc"))
			Expect(actualArch).To(Equal("osx"))
			Expect(actualConfig).To(Equal(config))
		})

		Context("when not targeted", func() {
			It("should print an error", func() {
				config.SetTarget("")

				test_helpers.ExecuteCommandWithArgs(syncCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error: Must be targeted to sync."))
				Expect(fakeSync.SyncLTCCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when the architecture is unknown", func() {
			It("should print an error", func() {
				commandFactory := command_factory.NewSyncCommandFactory(config, terminalUI, fakeExitHandler, "unknown-arch", "fakeltc", fakeSync)
				syncCommand = commandFactory.MakeSyncCommand()

				test_helpers.ExecuteCommandWithArgs(syncCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error: Unknown architecture unknown-arch. Sync not supported."))
				Expect(fakeSync.SyncLTCCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when the ltc binary can't be found", func() {
			It("should print an error", func() {
				commandFactory := command_factory.NewSyncCommandFactory(config, terminalUI, fakeExitHandler, "darwin", "", fakeSync)
				syncCommand = commandFactory.MakeSyncCommand()

				test_helpers.ExecuteCommandWithArgs(syncCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error: Unable to locate the ltc binary. Sync not supported."))
				Expect(fakeSync.SyncLTCCallCount()).To(Equal(0))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})

		Context("when SyncLTC fails", func() {
			It("should print an error", func() {
				fakeSync.SyncLTCReturns(errors.New("failed"))

				test_helpers.ExecuteCommandWithArgs(syncCommand, []string{})

				Expect(outputBuffer).To(test_helpers.SayLine("Error: failed"))
				Expect(fakeSync.SyncLTCCallCount()).To(Equal(1))
				Expect(fakeExitHandler.ExitCalledWith).To(Equal([]int{exit_codes.CommandFailed}))
			})
		})
	})
})
