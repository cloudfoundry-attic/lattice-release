package command_factory_test

import (
	"errors"

	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	config_package "github.com/pivotal-cf-experimental/lattice-cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	"github.com/pivotal-cf-experimental/lattice-cli/config/command_factory"
)

var _ = Describe("CommandFactory", func() {
	Describe("setApiEndpoint", func() {
		var (
			buffer  *gbytes.Buffer
			command cli.Command
			config  *config_package.Config
		)

		BeforeEach(func() {
			buffer = gbytes.NewBuffer()
			config = config_package.New(persister.NewFakePersister())

			commandFactory := command_factory.NewConfigCommandFactory(config, output.New(buffer))
			command = commandFactory.MakeSetTargetCommand()
		})

		Describe("targetCommand", func() {
			It("sets the api, username, password from the target specified", func() {
				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com", "--username=testusername", "--password=testpassword"})

				Expect(err).NotTo(HaveOccurred())
				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
				Expect(buffer).To(gbytes.Say("Api Location Set"))
			})

			It("does not set a username or password if none are passed in", func() {
				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(err).NotTo(HaveOccurred())
				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://receptor.myapi.com"))
				Expect(buffer).To(gbytes.Say("Api Location Set"))
			})

			It("returns an error if the target is blank", func() {
				err := test_helpers.ExecuteCommandWithArgs(command, []string{""})

				Expect(err).NotTo(HaveOccurred())
				Expect(buffer).To(gbytes.Say("Incorrect Usage: Target required."))
			})

			It("returns an error if username is set and password is not", func() {
				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com", "--username=testusername"})

				Expect(err).NotTo(HaveOccurred())
				Expect(buffer).To(gbytes.Say("Incorrect Usage: Password required with Username."))
			})

			It("returns an error if password is set and username is not", func() {
				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com", "--password=testpassword"})

				Expect(err).NotTo(HaveOccurred())
				Expect(buffer).To(gbytes.Say("Incorrect Usage: Username required with Password."))
			})

			It("buffers errors from setting the target", func() {
				fakePersister := persister.NewFakePersisterWithError(errors.New("FAILURE setting api"))

				commandFactory := command_factory.NewConfigCommandFactory(config_package.New(fakePersister), output.New(buffer))
				command = commandFactory.MakeSetTargetCommand()

				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(err).NotTo(HaveOccurred())
				Expect(buffer).To(gbytes.Say("FAILURE setting api"))
			})
		})
	})
})
