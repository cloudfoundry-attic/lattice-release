package main_test

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Building", func() {
	buildpackFixtures := "fixtures/buildpacks"
	appFixtures := "fixtures/apps"

	var (
		builderCmd *exec.Cmd

		tmpDir                    string
		buildDir                  string
		buildpacksDir             string
		outputDroplet             string
		buildpackOrder            string
		buildArtifactsCacheDir    string
		outputMetadata            string
		outputBuildArtifactsCache string
		skipDetect                bool
	)

	builder := func() *gexec.Session {
		session, err := gexec.Start(
			builderCmd,
			GinkgoWriter,
			GinkgoWriter,
		)
		Expect(err).NotTo(HaveOccurred())

		return session
	}

	cpBuildpack := func(buildpack string) {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(buildpack)))
		cp(path.Join(buildpackFixtures, buildpack), path.Join(buildpacksDir, hash))
	}

	BeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "building-tmp")
		buildDir, err = ioutil.TempDir(tmpDir, "building-app")
		Expect(err).NotTo(HaveOccurred())

		buildpacksDir, err = ioutil.TempDir(tmpDir, "building-buildpacks")
		Expect(err).NotTo(HaveOccurred())

		outputDropletFile, err := ioutil.TempFile(tmpDir, "building-droplet")
		Expect(err).NotTo(HaveOccurred())
		outputDroplet = outputDropletFile.Name()

		outputBuildArtifactsCacheDir, err := ioutil.TempDir(tmpDir, "building-cache-output")
		Expect(err).NotTo(HaveOccurred())
		outputBuildArtifactsCache = filepath.Join(outputBuildArtifactsCacheDir, "cache.tgz")

		buildArtifactsCacheDir, err = ioutil.TempDir(tmpDir, "building-cache")
		Expect(err).NotTo(HaveOccurred())

		outputMetadataFile, err := ioutil.TempFile(tmpDir, "building-result")
		Expect(err).NotTo(HaveOccurred())
		outputMetadata = outputMetadataFile.Name()

		buildpackOrder = ""

		skipDetect = false
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	JustBeforeEach(func() {
		builderCmd = exec.Command(builderPath,
			"-buildDir", buildDir,
			"-buildpacksDir", buildpacksDir,
			"-outputDroplet", outputDroplet,
			"-outputBuildArtifactsCache", outputBuildArtifactsCache,
			"-buildArtifactsCacheDir", buildArtifactsCacheDir,
			"-buildpackOrder", buildpackOrder,
			"-outputMetadata", outputMetadata,
			"-skipDetect="+strconv.FormatBool(skipDetect),
		)

		env := os.Environ()
		builderCmd.Env = append(env, "TMPDIR="+tmpDir)
	})

	resultJSON := func() []byte {
		resultInfo, err := ioutil.ReadFile(outputMetadata)
		Expect(err).NotTo(HaveOccurred())

		return resultInfo
	}

	Context("with a normal buildpack", func() {
		BeforeEach(func() {
			buildpackOrder = "always-detects,also-always-detects"

			cpBuildpack("always-detects")
			cpBuildpack("also-always-detects")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
		})

		JustBeforeEach(func() {
			Eventually(builder(), 5*time.Second).Should(gexec.Exit(0))
		})

		Describe("the contents of the output tgz", func() {
			var files []string

			JustBeforeEach(func() {
				result, err := exec.Command("tar", "-tzf", outputDroplet).Output()
				Expect(err).NotTo(HaveOccurred())

				files = strings.Split(string(result), "\n")
			})

			It("should contain an /app dir with the contents of the compilation", func() {
				Expect(files).To(ContainElement("./app/"))
				Expect(files).To(ContainElement("./app/app.sh"))
				Expect(files).To(ContainElement("./app/compiled"))
			})

			It("should contain an empty /tmp directory", func() {
				Expect(files).To(ContainElement("./tmp/"))
				Expect(files).NotTo(ContainElement(MatchRegexp("\\./tmp/.+")))
			})

			It("should contain an empty /logs directory", func() {
				Expect(files).To(ContainElement("./logs/"))
				Expect(files).NotTo(ContainElement(MatchRegexp("\\./logs/.+")))
			})

			It("should contain a staging_info.yml with the detected buildpack", func() {
				stagingInfo, err := exec.Command("tar", "-xzf", outputDroplet, "-O", "./staging_info.yml").Output()
				Expect(err).NotTo(HaveOccurred())

				expectedYAML := `detected_buildpack: Always Matching
start_command: the start command
`
				Expect(string(stagingInfo)).To(Equal(expectedYAML))
			})
		})

		Describe("the build artifacts cache output tgz", func() {
			BeforeEach(func() {
				buildpackOrder = "always-detects-creates-build-artifacts"

				cpBuildpack("always-detects-creates-build-artifacts")
			})

			It("gets created", func() {
				result, err := exec.Command("tar", "-tzf", outputBuildArtifactsCache).Output()
				Expect(err).NotTo(HaveOccurred())

				Expect(strings.Split(string(result), "\n")).To(ContainElement("./build-artifact"))
			})
		})

		Describe("the result.json, which is used to communicate back to the stager", func() {
			It("exists, and contains the detected buildpack", func() {
				Expect(resultJSON()).To(MatchJSON(`{
					"detected_buildpack": "Always Matching",
					"execution_metadata": "{\"start_command\":\"the start command\"}",
					"buildpack_key": "always-detects",
					"detected_start_command":{"web":"the start command"}
				}`))

			})

			Context("when the app has a Procfile", func() {
				Context("with web defined", func() {
					BeforeEach(func() {
						cp(path.Join(appFixtures, "with-procfile-with-web", "Procfile"), buildDir)
					})

					It("chooses the Procfile-provided command", func() {
						Expect(resultJSON()).To(MatchJSON(`{
					"detected_buildpack": "Always Matching",
					"execution_metadata": "{\"start_command\":\"procfile-provided start-command\"}",
					"buildpack_key": "always-detects",
					"detected_start_command":{"web":"procfile-provided start-command"}
				}`))

					})
				})

				Context("without web", func() {
					BeforeEach(func() {
						cp(path.Join(appFixtures, "with-procfile", "Procfile"), buildDir)
					})

					It("chooses the buildpack-provided command", func() {
						Expect(resultJSON()).To(MatchJSON(`{
					"detected_buildpack": "Always Matching",
					"execution_metadata": "{\"start_command\":\"the start command\"}",
					"buildpack_key": "always-detects",
					"detected_start_command":{"web":"the start command"}
				}`))

					})
				})
			})
		})
	})

	Context("with a buildpack that does not determine a start command", func() {
		BeforeEach(func() {
			buildpackOrder = "release-without-command"
			cpBuildpack("release-without-command")
		})

		Context("when the app has a Procfile", func() {
			Context("with web defined", func() {
				JustBeforeEach(func() {
					Eventually(builder(), 5*time.Second).Should(gexec.Exit(0))
				})

				BeforeEach(func() {
					cp(path.Join(appFixtures, "with-procfile-with-web", "Procfile"), buildDir)
				})

				It("uses the command defined by web in the Procfile", func() {
					Expect(resultJSON()).To(MatchJSON(`{
						"detected_buildpack": "Release Without Command",
						"execution_metadata": "{\"start_command\":\"procfile-provided start-command\"}",
						"buildpack_key": "release-without-command",
						"detected_start_command":{"web":"procfile-provided start-command"}
					}`))

				})
			})

			Context("without web", func() {
				BeforeEach(func() {
					cp(path.Join(appFixtures, "with-procfile", "Procfile"), buildDir)
				})

				It("fails", func() {
					session := builder()
					Eventually(session.Err).Should(gbytes.Say("No start command detected"))
					Eventually(session).Should(gexec.Exit(0))
				})
			})
		})

		Context("and the app has no Procfile", func() {
			BeforeEach(func() {
				cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
			})

			It("fails", func() {
				session := builder()
				Eventually(session.Err).Should(gbytes.Say("No start command detected"))
				Eventually(session).Should(gexec.Exit(0))
			})
		})
	})

	Context("with an app with an invalid Procfile", func() {
		BeforeEach(func() {
			buildpackOrder = "always-detects,also-always-detects"

			cpBuildpack("always-detects")
			cpBuildpack("also-always-detects")

			cp(path.Join(appFixtures, "bogus-procfile", "Procfile"), buildDir)
		})

		It("fails", func() {
			session := builder()
			Eventually(session.Err).Should(gbytes.Say("Failed to read command from Procfile: invalid YAML"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when no buildpacks match", func() {
		BeforeEach(func() {
			buildpackOrder = "always-fails"

			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
			cpBuildpack("always-fails")
		})

		It("should exit with an error", func() {
			session := builder()
			Eventually(session.Err).Should(gbytes.Say("None of the buildpacks detected a compatible application"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the buildpack fails in compile", func() {
		BeforeEach(func() {
			buildpackOrder = "fails-to-compile"

			cpBuildpack("fails-to-compile")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
		})

		It("should exit with an error", func() {
			session := builder()
			Eventually(session.Err).Should(gbytes.Say("Failed to compile droplet: exit status 1"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the buildpack release generates invalid yaml", func() {
		BeforeEach(func() {
			buildpackOrder = "release-generates-bad-yaml"

			cpBuildpack("release-generates-bad-yaml")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
		})

		It("should exit with an error", func() {
			session := builder()
			Eventually(session.Err).Should(gbytes.Say("buildpack's release output invalid"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the buildpack fails to release", func() {
		BeforeEach(func() {
			buildpackOrder = "fails-to-release"

			cpBuildpack("fails-to-release")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
		})

		It("should exit with an error", func() {
			session := builder()
			Eventually(session.Err).Should(gbytes.Say("Failed to build droplet release: exit status 1"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("with a nested buildpack", func() {
		BeforeEach(func() {
			nestedBuildpack := "nested-buildpack"
			buildpackOrder = nestedBuildpack

			nestedBuildpackHash := "70d137ae4ee01fbe39058ccdebf48460"

			nestedBuildpackDir := path.Join(buildpacksDir, nestedBuildpackHash)
			err := os.MkdirAll(nestedBuildpackDir, 0777)
			Expect(err).NotTo(HaveOccurred())

			cp(path.Join(buildpackFixtures, "always-detects"), nestedBuildpackDir)
			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
		})

		It("should detect the nested buildpack", func() {
			Eventually(builder()).Should(gexec.Exit(0))
		})
	})

	Context("skipping detect", func() {
		BeforeEach(func() {
			buildpackOrder = "always-fails"
			skipDetect = true

			cp(path.Join(appFixtures, "bash-app", "app.sh"), buildDir)
			cpBuildpack("always-fails")
		})

		It("should exit with an error", func() {
			session := builder()
			Eventually(session).Should(gexec.Exit(0))
		})
	})
})

func cp(src string, dst string) {
	session, err := gexec.Start(
		exec.Command("cp", "-a", src, dst),
		GinkgoWriter,
		GinkgoWriter,
	)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}
