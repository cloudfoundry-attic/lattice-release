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
			inbuffer *gbytes.Buffer
			buffer   *gbytes.Buffer
			command  cli.Command
			config   *config_package.Config
		)

		BeforeEach(func() {
			inbuffer = gbytes.NewBuffer()
			buffer = gbytes.NewBuffer()
			config = config_package.New(persister.NewFakePersister())

			commandFactory := command_factory.NewConfigCommandFactory(config, inbuffer, output.New(buffer))
			command = commandFactory.MakeSetTargetCommand()
		})

		Describe("targetCommand", func() {
			It("sets the api, username, password from the target specified", func() {

				inbuffer.Write([]byte("testusername\n"))
				inbuffer.Write([]byte("testpassword\n"))

				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(buffer).To(gbytes.Say("Username: "))
				Expect(buffer).To(gbytes.Say("Password: "))

				Expect(err).NotTo(HaveOccurred())
				Expect(config.Target()).To(Equal("myapi.com"))
				Expect(config.Receptor()).To(Equal("http://testusername:testpassword@receptor.myapi.com"))
				Expect(buffer).To(gbytes.Say("Api Location Set"))
			})

			It("does not set a username or password if none are passed in", func() {
				inbuffer.Write([]byte("\n"))
				inbuffer.Write([]byte("\n"))

				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(buffer).To(gbytes.Say("Username: "))
				Expect(buffer).To(gbytes.Say("Password: "))

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

			It("bubbles errors from setting the target", func() {
				fakePersister := persister.NewFakePersisterWithError(errors.New("FAILURE setting api"))

				inbuffer.Write([]byte("\n"))
				inbuffer.Write([]byte("\n"))

				commandFactory := command_factory.NewConfigCommandFactory(config_package.New(fakePersister), inbuffer, output.New(buffer))
				command = commandFactory.MakeSetTargetCommand()

				err := test_helpers.ExecuteCommandWithArgs(command, []string{"myapi.com"})

				Expect(buffer).To(gbytes.Say("Username: "))
				Expect(buffer).To(gbytes.Say("Password: "))

				Expect(err).NotTo(HaveOccurred())
				Expect(buffer).To(gbytes.Say("FAILURE setting api"))
			})
		})
	})
})
