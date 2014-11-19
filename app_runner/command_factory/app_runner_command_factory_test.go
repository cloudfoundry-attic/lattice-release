package command_factory_test

import (
	"errors"

	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/diego-edge-cli/test_helpers"

	"github.com/pivotal-cf-experimental/diego-edge-cli/app_runner/command_factory"
)

type startedDockerApps struct {
	name            string
	startCommand    string
	dockerImagePath string
}

type fakeAppRunner struct {
	err               error
	startedDockerApps []startedDockerApps
}

func (f *fakeAppRunner) StartDockerApp(name, startCommand, dockerImagePath string) error {
	if f.err != nil {
		return f.err
	}
	f.startedDockerApps = append(f.startedDockerApps, startedDockerApps{name, startCommand, dockerImagePath})
	return nil
}

func (f *fakeAppRunner) SetError(err error) {
	f.err = err
}

var _ = Describe("CommandFactory", func() {

	var (
		appRunner fakeAppRunner
		buffer    *gbytes.Buffer
		command   cli.Command
	)

	BeforeEach(func() {
		appRunner = fakeAppRunner{startedDockerApps: []startedDockerApps{}}
		buffer = gbytes.NewBuffer()
		commandFactory := command_factory.NewAppRunnerCommandFactory(&appRunner, buffer)
		command = commandFactory.MakeCommand()
	})

	Describe("startDiegoApp", func() {
		It("starts a Docker based Diego app as specified in the command via the AppRunner", func() {

			args := []string{
				"--docker-image=docker://fun/app",
				"--start-command=/start-me-please",
				"cool-web-app",
			}

			context := test_helpers.ContextFromArgsAndCommand(args, command)

			command.Action(context)

			Expect(len(appRunner.startedDockerApps)).To(Equal(1))
			Expect(appRunner.startedDockerApps[0].name).To(Equal("cool-web-app"))
			Expect(appRunner.startedDockerApps[0].startCommand).To(Equal("/start-me-please"))
			Expect(appRunner.startedDockerApps[0].dockerImagePath).To(Equal("docker://fun/app"))

			Expect(buffer).To(gbytes.Say("App Staged Successfully"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"--docker-image=docker://fun/app",
				"--start-command=/start-me-please",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, command)

			command.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("validates that the dockerImage is passed in", func() {
			args := []string{
				"--start-command=/start-me-please",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, command)

			command.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("validates that the startCommand is passed in", func() {
			args := []string{
				"--docker-image=docker://fun/app",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, command)

			command.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("outputs error messages", func() {
			args := []string{
				"--docker-image=docker://fun/app",
				"--start-command=/start-me-please",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, command)

			appRunner.SetError(errors.New("Major Fault"))

			command.Action(context)

			Expect(buffer).To(gbytes.Say("Error Starting App: Major Fault"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})
	})
})
