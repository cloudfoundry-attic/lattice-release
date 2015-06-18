package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Launcher", func() {
	var extractDir string
	var appDir string
	var launcherCmd *exec.Cmd
	var session *gexec.Session

	BeforeEach(func() {
		os.Setenv("CALLERENV", "some-value")

		var err error
		extractDir, err = ioutil.TempDir("", "vcap")
		Expect(err).NotTo(HaveOccurred())

		appDir = path.Join(extractDir, "app")
		err = os.MkdirAll(appDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		launcherCmd = &exec.Cmd{
			Path: launcher,
			Dir:  extractDir,
			Env: append(
				os.Environ(),
				"PORT=8080",
				"INSTANCE_GUID=some-instance-guid",
				"INSTANCE_INDEX=123",
				`VCAP_APPLICATION={"foo":1}`,
			),
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(extractDir)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(launcherCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	var ItExecutesTheCommandWithTheRightEnvironment = func() {
		It("executes with the environment of the caller", func() {
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session).Should(gbytes.Say("CALLERENV=some-value"))
		})

		It("executes the start command with $HOME as the given dir", func() {
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session).Should(gbytes.Say("HOME=" + appDir))
		})

		It("changes to the app directory when running", func() {
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session).Should(gbytes.Say("PWD=" + appDir))
		})

		It("executes the start command with $TMPDIR as the extract directory + /tmp", func() {
			absDir, err := filepath.Abs(filepath.Join(appDir, "..", "tmp"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session).Should(gbytes.Say("TMPDIR=" + absDir))
		})

		It("munges VCAP_APPLICATION appropriately", func() {
			Eventually(session).Should(gexec.Exit(0))

			vcapAppPattern := regexp.MustCompile("VCAP_APPLICATION=(.*)")
			vcapApplicationBytes := vcapAppPattern.FindSubmatch(session.Out.Contents())[1]

			vcapApplication := map[string]interface{}{}
			err := json.Unmarshal(vcapApplicationBytes, &vcapApplication)
			Expect(err).NotTo(HaveOccurred())

			Expect(vcapApplication["host"]).To(Equal("0.0.0.0"))
			Expect(vcapApplication["port"]).To(Equal(float64(8080)))
			Expect(vcapApplication["instance_index"]).To(Equal(float64(123)))
			Expect(vcapApplication["instance_id"]).To(Equal("some-instance-guid"))
			Expect(vcapApplication["foo"]).To(Equal(float64(1)))
		})

		Context("when the given dir has .profile.d with scripts in it", func() {
			BeforeEach(func() {
				var err error

				profileDir := path.Join(appDir, ".profile.d")

				err = os.MkdirAll(profileDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(profileDir, "a.sh"), []byte("echo sourcing a\nexport A=1\n"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(profileDir, "b.sh"), []byte("echo sourcing b\nexport B=1\n"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("sources them before executing", func() {
				Eventually(session).Should(gexec.Exit(0))
				Eventually(session).Should(gbytes.Say("sourcing a"))
				Eventually(session).Should(gbytes.Say("sourcing b"))
				Eventually(session).Should(gbytes.Say("A=1"))
				Eventually(session).Should(gbytes.Say("B=1"))
				Eventually(session).Should(gbytes.Say("running app"))
			})
		})
	}

	Context("when a start command is given", func() {
		BeforeEach(func() {
			launcherCmd.Args = []string{
				"launcher",
				appDir,
				"env; echo running app",
				`{ "start_command": "echo should not run this" }`,
			}
		})

		ItExecutesTheCommandWithTheRightEnvironment()
	})

	Context("when no start command is given", func() {
		BeforeEach(func() {
			launcherCmd.Args = []string{
				"launcher",
				appDir,
				"",
				`{ "start_command": "env; echo running app" }`,
			}
		})

		ItExecutesTheCommandWithTheRightEnvironment()
	})

	var ItPrintsUsageInformation = func() {
		It("prints usage information", func() {
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Usage: launcher <app directory> <start command> <metadata>"))
		})
	}

	Context("when the start command and start_command metadata are empty", func() {
		BeforeEach(func() {
			launcherCmd.Args = []string{
				"launcher",
				appDir,
				"",
				"",
			}
		})

		Context("when the app package does not contain staging_info.yml", func() {
			ItPrintsUsageInformation()
		})

		Context("when the app package has a staging_info.yml", func() {

			Context("when it is missing a start command", func() {
				BeforeEach(func() {
					writeStagingInfo(extractDir, "detected_buildpack: Ruby")
				})

				ItPrintsUsageInformation()
			})

			Context("when it contains a start command", func() {
				BeforeEach(func() {
					writeStagingInfo(extractDir, "detected_buildpack: Ruby\nstart_command: env; echo running app")
				})

				ItExecutesTheCommandWithTheRightEnvironment()
			})

			Context("when it references unresolvable types in non-essential fields", func() {
				BeforeEach(func() {
					writeStagingInfo(
						extractDir,
						"---\nbuildpack_path: !ruby/object:Pathname\n  path: /tmp/buildpacks/null-buildpack\ndetected_buildpack: \nstart_command: env; echo running app\n",
					)
				})

				ItExecutesTheCommandWithTheRightEnvironment()
			})

			Context("when it is not valid YAML", func() {
				BeforeEach(func() {
					writeStagingInfo(extractDir, "start_command: &ruby/object:Pathname")
				})

				It("prints an error message", func() {
					Eventually(session).Should(gexec.Exit(1))
					Eventually(session.Err).Should(gbytes.Say("Invalid staging info"))
				})
			})

		})

	})

	Context("when arguments are missing", func() {
		BeforeEach(func() {
			launcherCmd.Args = []string{
				"launcher",
				appDir,
				"env",
			}
		})

		ItPrintsUsageInformation()
	})

	Context("when the given execution metadata is not valid JSON", func() {
		BeforeEach(func() {
			launcherCmd.Args = []string{
				"launcher",
				appDir,
				"",
				"{ not-valid-json }",
			}
		})

		It("prints an error message", func() {
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Invalid metadata"))
		})
	})
})

func writeStagingInfo(extractDir, stagingInfo string) {
	err := ioutil.WriteFile(filepath.Join(extractDir, "staging_info.yml"), []byte(stagingInfo), 0644)
	Expect(err).NotTo(HaveOccurred())
}
