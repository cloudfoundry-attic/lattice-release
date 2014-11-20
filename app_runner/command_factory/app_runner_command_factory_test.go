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

var _ = Describe("CommandFactory", func() {

	var (
		appRunner fakeAppRunner
		buffer    *gbytes.Buffer
	)

	BeforeEach(func() {
		appRunner = fakeAppRunner{startedDockerApps: []startedDockerApps{}, scaledDockerApps: []scaledDockerApps{}, stoppedDockerApps: []stoppedDockerApps{}}
		buffer = gbytes.NewBuffer()
	})

	Describe("startDiegoApp", func() {

		var startDiegoCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewAppRunnerCommandFactory(&appRunner, buffer)
			startDiegoCommand = commandFactory.MakeStartDiegoAppCommand()
		})

		It("starts a Docker based Diego app as specified in the command via the AppRunner", func() {

			args := []string{
				"--docker-image=docker://fun/app",
				"--start-command=/start-me-please",
				"cool-web-app",
			}

			context := test_helpers.ContextFromArgsAndCommand(args, startDiegoCommand)

			startDiegoCommand.Action(context)

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
			context := test_helpers.ContextFromArgsAndCommand(args, startDiegoCommand)

			startDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("validates that the dockerImage is passed in", func() {
			args := []string{
				"--start-command=/start-me-please",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, startDiegoCommand)

			startDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("validates that the startCommand is passed in", func() {
			args := []string{
				"--docker-image=docker://fun/app",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, startDiegoCommand)

			startDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("outputs error messages", func() {
			args := []string{
				"--docker-image=docker://fun/app",
				"--start-command=/start-me-please",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, startDiegoCommand)

			appRunner.SetError(errors.New("Major Fault"))

			startDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Error Starting App: Major Fault"))
		})
	})

	Describe("scaleDiegoApp", func() {

		var scaleDiegoCommand cli.Command
		BeforeEach(func() {
			commandFactory := command_factory.NewAppRunnerCommandFactory(&appRunner, buffer)
			scaleDiegoCommand = commandFactory.MakeScaleDiegoAppCommand()
		})

		It("starts a Docker based Diego app as specified in the command via the AppRunner", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			context := test_helpers.ContextFromArgsAndCommand(args, scaleDiegoCommand)

			scaleDiegoCommand.Action(context)

			Expect(len(appRunner.scaledDockerApps)).To(Equal(1))
			Expect(appRunner.scaledDockerApps[0].name).To(Equal("cool-web-app"))
			Expect(appRunner.scaledDockerApps[0].instances).To(Equal(22))

			Expect(buffer).To(gbytes.Say("App Scaled Successfully"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"--instances=22",
				"",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, scaleDiegoCommand)

			scaleDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.scaledDockerApps)).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, scaleDiegoCommand)

			appRunner.SetError(errors.New("Major Fault"))

			scaleDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Error Scaling App: Major Fault"))
		})

		It("validates that the number instances is nonzero", func() {
			args := []string{
				"--instances=0",
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, scaleDiegoCommand)

			scaleDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Error Scaling to 0 instances - Please stop with: diego-edge-cli stop cool-web-app"))
			Expect(len(appRunner.scaledDockerApps)).To(Equal(0))
		})
	})

	Describe("stopDiegoApp", func() {

		var stopDiegoCommand cli.Command
		BeforeEach(func() {
			commandFactory := command_factory.NewAppRunnerCommandFactory(&appRunner, buffer)
			stopDiegoCommand = commandFactory.MakeStopDiegoAppCommand()
		})

		It("stops a Docker based Diego app as specified in the command via the AppRunner", func() {
			args := []string{
				"cool-web-app",
			}

			context := test_helpers.ContextFromArgsAndCommand(args, stopDiegoCommand)

			stopDiegoCommand.Action(context)

			Expect(len(appRunner.stoppedDockerApps)).To(Equal(1))
			Expect(appRunner.stoppedDockerApps[0].name).To(Equal("cool-web-app"))

			Expect(buffer).To(gbytes.Say("App Stopped Successfully"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, stopDiegoCommand)

			stopDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Incorrect Usage\n"))
			Expect(len(appRunner.stoppedDockerApps)).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}
			context := test_helpers.ContextFromArgsAndCommand(args, stopDiegoCommand)

			appRunner.SetError(errors.New("Major Fault"))

			stopDiegoCommand.Action(context)

			Expect(buffer).To(gbytes.Say("Error Stopping App: Major Fault"))
		})
	})
})

type startedDockerApps struct {
	name            string
	startCommand    string
	dockerImagePath string
}

type scaledDockerApps struct {
	name      string
	instances int
}

type stoppedDockerApps struct {
	name string
}

type fakeAppRunner struct {
	err               error
	startedDockerApps []startedDockerApps
	scaledDockerApps  []scaledDockerApps
	stoppedDockerApps []stoppedDockerApps
}

func (f *fakeAppRunner) StartDockerApp(name, startCommand, dockerImagePath string) error {
	if f.err != nil {
		return f.err
	}
	f.startedDockerApps = append(f.startedDockerApps, startedDockerApps{name, startCommand, dockerImagePath})
	return nil
}

func (f *fakeAppRunner) ScaleDockerApp(name string, instances int) error {
	if f.err != nil {
		return f.err
	}
	f.scaledDockerApps = append(f.scaledDockerApps, scaledDockerApps{name, instances})
	return nil
}

func (f *fakeAppRunner) StopDockerApp(name string) error {
	if f.err != nil {
		return f.err
	}
	f.stoppedDockerApps = append(f.stoppedDockerApps, stoppedDockerApps{name})
	return nil
}

func (f *fakeAppRunner) SetError(err error) {
	f.err = err
}
