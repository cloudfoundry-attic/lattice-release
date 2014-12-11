package command_factory_test

import (
	"errors"
	"time"

	"github.com/dajulia3/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/colors"
	"github.com/pivotal-cf-experimental/lattice-cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/test_helpers"

	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/command_factory"
)

var _ = Describe("CommandFactory", func() {

	var (
		appRunner *fakeAppRunner
		buffer    *gbytes.Buffer
		timeout   time.Duration = 1 * time.Millisecond
		domain    string        = "192.168.11.11.xip.io"
	)

	BeforeEach(func() {
		appRunner = newFakeAppRunner()
		buffer = gbytes.NewBuffer()
	})

	Describe("startApp", func() {

		var startCommand cli.Command

		BeforeEach(func() {
			env := []string{"SHELL=/bin/bash", "COLOR=Blue"}

			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, env)
			startCommand = commandFactory.MakeStartAppCommand()
		})

		It("starts a Docker based  app as specified in the command via the AppRunner", func() {
			args := []string{
				"--memory-mb=12",
				"--disk-mb=12",
				"--port=3000",
				"--docker-image=docker:///fun/app",
				"--run-as-root=true",
				"--env=TIMEZONE=CST",
				"--env=LANG=\"Chicago English\"",
				"--env=COLOR",
				"--env=UNSET",
				"cool-web-app",
				"--",
				"/start-me-please",
				"AppArg0",
				"--appFlavor=\"purple\"",
			}

			appRunner.upDockerApps["cool-web-app"] = true
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())

			Expect(len(appRunner.startedDockerApps)).To(Equal(1))
			Expect(appRunner.startedDockerApps[0].name).To(Equal("cool-web-app"))
			Expect(appRunner.startedDockerApps[0].startCommand).To(Equal("/start-me-please"))
			Expect(appRunner.startedDockerApps[0].dockerImagePath).To(Equal("docker:///fun/app"))
			Expect(appRunner.startedDockerApps[0].appArgs).To(Equal([]string{"AppArg0", "--appFlavor=\"purple\""}))
			Expect(appRunner.startedDockerApps[0].environmentVariables).To(Equal(map[string]string{"TIMEZONE": "CST", "LANG": "\"Chicago English\"", "COLOR": "Blue", "UNSET": ""}))
			Expect(appRunner.startedDockerApps[0].privileged).To(Equal(true))
			Expect(appRunner.startedDockerApps[0].memoryMB).To(Equal(12))
			Expect(appRunner.startedDockerApps[0].diskMB).To(Equal(12))
			Expect(appRunner.startedDockerApps[0].port).To(Equal(3000))

			Expect(buffer).To(gbytes.Say("Starting App: cool-web-app"))
			Expect(string(buffer.Contents())).To(ContainSubstring(colors.Green("cool-web-app is now running.")))
			Expect(string(buffer.Contents())).To(ContainSubstring(colors.Green("http://cool-web-app.192.168.11.11.xip.io")))
		})

		It("starts a Docker based app with sensible defaults", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.upDockerApps["cool-web-app"] = true
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())

			Expect(len(appRunner.startedDockerApps)).To(Equal(1))
			Expect(appRunner.startedDockerApps[0].privileged).To(Equal(false))
			Expect(appRunner.startedDockerApps[0].memoryMB).To(Equal(128))
			Expect(appRunner.startedDockerApps[0].diskMB).To(Equal(1024))
			Expect(appRunner.startedDockerApps[0].port).To(Equal(8080))
		})

		It("alerts the user if the app does not start", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.upDockerApps["cool-web-app"] = false
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(buffer.Contents())).To(ContainSubstring(colors.Red("cool-web-app took too long to start.")))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
			}

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: App Name required"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))
		})

		It("validates that the dockerImage is passed in", func() {
			args := []string{
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: Docker Image required"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))

		})

		It("validates that the startCommand is passed in", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: Start Command required"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))
		})

		It("validates that the terminator -- is passed in", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"not-the-terminator",
				"start-me-up",
			}
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: '--' Required before start command"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))
		})

		It("validates that the full docker path is passed in", func() {
			args := []string{
				"--docker-image=fun/app",
				"cool-web-app",
				"--",
				"start-me-please",
			}
			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: Docker Image should begin with: docker:///"))
			Expect(len(appRunner.startedDockerApps)).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"--docker-image=docker:///fun/app",
				"cool-web-app",
				"--",
				"/start-me-please",
			}

			appRunner.SetError(errors.New("Major Fault"))

			err := test_helpers.ExecuteCommandWithArgs(startCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Error Starting App: Major Fault"))
		})
	})

	Describe("scaleApp", func() {

		var scaleCommand cli.Command
		BeforeEach(func() {
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, []string{})
			scaleCommand = commandFactory.MakeScaleAppCommand()
		})

		It("starts a Docker based  app as specified in the command via the AppRunner", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
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

			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: App Name required"))
			Expect(len(appRunner.scaledDockerApps)).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"--instances=22",
				"cool-web-app",
			}

			appRunner.SetError(errors.New("Major Fault"))
			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Error Scaling App: Major Fault"))
		})

		It("validates that the number instances is nonzero", func() {
			args := []string{
				"--instances=0",
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(scaleCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Error Scaling to 0 instances - Please stop with: lattice-cli stop cool-web-app"))
			Expect(len(appRunner.scaledDockerApps)).To(Equal(0))
		})
	})

	Describe("stopApp", func() {

		var stopCommand cli.Command
		BeforeEach(func() {
			commandFactory := command_factory.NewAppRunnerCommandFactory(appRunner, output.New(buffer), timeout, domain, []string{})
			stopCommand = commandFactory.MakeStopAppCommand()
		})

		It("stops a Docker based  app as specified in the command via the AppRunner", func() {
			args := []string{
				"cool-web-app",
			}

			err := test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(appRunner.stoppedDockerApps)).To(Equal(1))
			Expect(appRunner.stoppedDockerApps[0].name).To(Equal("cool-web-app"))

			Expect(buffer).To(gbytes.Say("App Stopped Successfully"))
		})

		It("validates that the name is passed in", func() {
			args := []string{
				"",
			}

			err := test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Incorrect Usage: App Name required"))
			Expect(len(appRunner.stoppedDockerApps)).To(Equal(0))
		})

		It("outputs error messages", func() {
			args := []string{
				"cool-web-app",
			}

			appRunner.SetError(errors.New("Major Fault"))
			err := test_helpers.ExecuteCommandWithArgs(stopCommand, args)

			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say("Error Stopping App: Major Fault"))
		})
	})
})

func newFakeAppRunner() *fakeAppRunner {
	return &fakeAppRunner{
		startedDockerApps: []startedDockerApps{},
		scaledDockerApps:  []scaledDockerApps{},
		stoppedDockerApps: []stoppedDockerApps{},
		upDockerApps:      map[string]bool{},
	}
}

type startedDockerApps struct {
	name                 string
	dockerImagePath      string
	startCommand         string
	appArgs              []string
	environmentVariables map[string]string
	privileged           bool
	memoryMB             int
	diskMB               int
	port                 int
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
	upDockerApps      map[string]bool
}

func (f *fakeAppRunner) StartDockerApp(name, dockerImagePath, startCommand string, appArgs []string, environmentVariables map[string]string, privileged bool, memoryMB, diskMB, port int) error {
	if f.err != nil {
		return f.err
	}
	f.startedDockerApps = append(f.startedDockerApps,
		startedDockerApps{
			name:                 name,
			dockerImagePath:      dockerImagePath,
			startCommand:         startCommand,
			appArgs:              appArgs,
			environmentVariables: environmentVariables,
			privileged:           privileged,
			memoryMB:             memoryMB,
			diskMB:               diskMB,
			port:                 port,
		})
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

func (f *fakeAppRunner) IsDockerAppUp(name string) (bool, error) {
	return f.upDockerApps[name], nil
}

func (f *fakeAppRunner) SetError(err error) {
	f.err = err
}
