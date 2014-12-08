package command_factory_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config/persister"
	"github.com/pivotal-cf-experimental/diego-edge-cli/test_helpers"

	"github.com/pivotal-cf-experimental/diego-edge-cli/config/command_factory"
)

var _ = Describe("CommandFactory", func() {
	Describe("setApiEndpoint", func() {
		var (
			output *gbytes.Buffer
		)

		BeforeEach(func() {
			output = gbytes.NewBuffer()
		})

		Describe("targetCommand", func() {
			It("sets the api from the target specified", func() {
				config := config.New(persister.NewFakePersister())
				commandFactory := command_factory.NewConfigCommandFactory(config, output)

				command := commandFactory.MakeSetTargetCommand()

				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(err).NotTo(HaveOccurred())
				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://receptor.myapi.com"))
				Expect(output).To(gbytes.Say("Api Location Set\n"))
			})

			It("returns an error if the target is blank", func() {
				config := config.New(persister.NewFakePersister())
				commandFactory := command_factory.NewConfigCommandFactory(config, output)

				command := commandFactory.MakeSetTargetCommand()

				err := test_helpers.ExecuteCommandWithArgs(command, []string{""})

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(gbytes.Say("Incorrect Usage\n"))
			})

			It("outputs errors from setting the target", func() {
				fakePersister := persister.NewFakePersisterWithError(errors.New("FAILURE setting api"))

				config := config.New(fakePersister)
				commandFactory := command_factory.NewConfigCommandFactory(config, output)

				command := commandFactory.MakeSetTargetCommand()

				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(gbytes.Say("FAILURE setting api"))
			})
		})

	})
})
