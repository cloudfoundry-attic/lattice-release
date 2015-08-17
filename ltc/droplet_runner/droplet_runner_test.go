package droplet_runner_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/fake_app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/fake_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner/fake_task_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("DropletRunner", func() {
	var (
		fakeAppRunner      *fake_app_runner.FakeAppRunner
		fakeTaskRunner     *fake_task_runner.FakeTaskRunner
		config             *config_package.Config
		fakeBlobStore      *fake_blob_store.FakeBlobStore
		fakeTargetVerifier *fake_target_verifier.FakeTargetVerifier
		fakeAppExaminer    *fake_app_examiner.FakeAppExaminer
		dropletRunner      droplet_runner.DropletRunner
	)

	BeforeEach(func() {
		fakeAppRunner = &fake_app_runner.FakeAppRunner{}
		fakeTaskRunner = &fake_task_runner.FakeTaskRunner{}
		config = config_package.New(persister.NewMemPersister())
		fakeBlobStore = &fake_blob_store.FakeBlobStore{}
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		fakeAppExaminer = &fake_app_examiner.FakeAppExaminer{}
		dropletRunner = droplet_runner.New(fakeAppRunner, fakeTaskRunner, config, fakeBlobStore, fakeTargetVerifier, fakeAppExaminer)
	})

	Describe("ListDroplets", func() {
		It("returns a list of droplets in the blob store", func() {
			fakeBlobStore.ListReturns([]dav_blob_store.Blob{
				{Path: "X/bits.zip", Created: time.Unix(1000, 0), Size: 100},
				{Path: "X/droplet.tgz", Created: time.Unix(2000, 0), Size: 200},
				{Path: "X/result.json", Created: time.Unix(3000, 0), Size: 300},
				{Path: "Y/bits.zip"},
				{Path: "X/Y/droplet.tgz"},
				{Path: "droplet.tgz"},
			}, nil)

			Expect(dropletRunner.ListDroplets()).To(Equal([]droplet_runner.Droplet{
				{Name: "X", Created: time.Unix(2000, 0), Size: 200},
			}))
		})

		It("returns an error when querying the blob store fails", func() {
			fakeBlobStore.ListReturns(nil, errors.New("some error"))

			_, err := dropletRunner.ListDroplets()
			Expect(err).To(MatchError("some error"))
		})
	})

	Describe("UploadBits", func() {
		Context("when the archive path is a file and exists", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpDir := os.TempDir()
				var err error
				tmpFile, err = ioutil.TempFile(tmpDir, "tmp_file")
				Expect(err).NotTo(HaveOccurred())

				Expect(ioutil.WriteFile(tmpFile.Name(), []byte("some contents"), 0600)).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.Remove(tmpFile.Name())).To(Succeed())
			})

			It("uploads the file to the bucket", func() {
				Expect(dropletRunner.UploadBits("droplet-name", tmpFile.Name())).To(Succeed())

				Expect(fakeBlobStore.UploadCallCount()).To(Equal(1))
				path, contents := fakeBlobStore.UploadArgsForCall(0)
				Expect(path).To(Equal("droplet-name/bits.zip"))
				Expect(ioutil.ReadAll(contents)).To(Equal([]byte("some contents")))
			})

			It("returns an error when we fail to open the droplet bits", func() {
				err := dropletRunner.UploadBits("droplet-name", "some non-existent file")
				Expect(err).To(MatchError("open some non-existent file: no such file or directory"))
			})

			It("returns an error when the upload fails", func() {
				fakeBlobStore.UploadReturns(errors.New("some error"))

				err := dropletRunner.UploadBits("droplet-name", tmpFile.Name())
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("BuildDroplet", func() {
		It("does the build droplet task", func() {
			config.SetBlobStore("blob-host", "7474", "dav-user", "dav-pass")
			Expect(config.Save()).To(Succeed())

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", map[string]string{}, 128, 100, 800)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
			createTaskParams := fakeTaskRunner.CreateTaskArgsForCall(0)
			Expect(createTaskParams).ToNot(BeNil())
			receptorRequest := createTaskParams.GetReceptorRequest()

			blobURL := fmt.Sprintf("http://%s:%s@%s:%s%s",
				config.BlobStore().Username,
				config.BlobStore().Password,
				config.BlobStore().Host,
				config.BlobStore().Port,
				"/blobs/droplet-name")

			expectedActions := &models.SerialAction{
				Actions: []models.Action{
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
						To:   "/tmp",
						User: "vcap",
					},
					&models.DownloadAction{
						From: "http://dav-user:dav-pass@blob-host:7474/blobs/droplet-name/bits.zip",
						To:   "/tmp/app",
						User: "vcap",
					},
					&models.RunAction{
						Path: "/tmp/davtool",
						Dir:  "/",
						Args: []string{"delete", blobURL + "/bits.zip"},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/bin/chmod",
						Dir:  "/tmp/app",
						Args: []string{"-R", "a+X", "."},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/tmp/builder",
						Dir:  "/",
						Args: []string{
							"-buildArtifactsCacheDir=/tmp/cache",
							"-buildDir=/tmp/app",
							"-buildpackOrder=buildpack",
							"-buildpacksDir=/tmp/buildpacks",
							"-outputBuildArtifactsCache=/tmp/output-cache",
							"-outputDroplet=/tmp/droplet",
							"-outputMetadata=/tmp/result.json",
							"-skipCertVerify=false",
							"-skipDetect=true",
						},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/tmp/davtool",
						Dir:  "/",
						Args: []string{"put", blobURL + "/droplet.tgz", "/tmp/droplet"},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/tmp/davtool",
						Dir:  "/",
						Args: []string{"put", blobURL + "/result.json", "/tmp/result.json"},
						User: "vcap",
					},
				},
			}
			Expect(receptorRequest.Action).To(Equal(expectedActions))
			Expect(receptorRequest.TaskGuid).To(Equal("task-name"))
			Expect(receptorRequest.LogGuid).To(Equal("task-name"))
			Expect(receptorRequest.MetricsGuid).To(Equal("task-name"))
			Expect(receptorRequest.RootFS).To(Equal("preloaded:cflinuxfs2"))
			Expect(receptorRequest.EnvironmentVariables).To(matchers.ContainExactly([]receptor.EnvironmentVariable{
				{Name: "CF_STACK", Value: "cflinuxfs2"},
				{Name: "MEMORY_LIMIT", Value: "128M"},
			}))
			Expect(receptorRequest.LogSource).To(Equal("BUILD"))
			Expect(receptorRequest.Domain).To(Equal("lattice"))
			Expect(receptorRequest.Privileged).To(BeTrue())
			Expect(receptorRequest.EgressRules).ToNot(BeNil())
			Expect(receptorRequest.EgressRules).To(BeEmpty())
			Expect(receptorRequest.MemoryMB).To(Equal(128))
		})

		It("passes through user environment variables", func() {
			config.SetBlobStore("blob-host", "7474", "dav-user", "dav-pass")
			Expect(config.Save()).To(Succeed())

			env := map[string]string{
				"ENV_VAR":   "stuff",
				"OTHER_VAR": "same",
			}

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", env, 128, 100, 800)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
			createTaskParams := fakeTaskRunner.CreateTaskArgsForCall(0)
			Expect(createTaskParams).ToNot(BeNil())
			receptorRequest := createTaskParams.GetReceptorRequest()

			Expect(receptorRequest.EnvironmentVariables).To(matchers.ContainExactly([]receptor.EnvironmentVariable{
				{Name: "CF_STACK", Value: "cflinuxfs2"},
				{Name: "MEMORY_LIMIT", Value: "128M"},
				{Name: "ENV_VAR", Value: "stuff"},
				{Name: "OTHER_VAR", Value: "same"},
			}))
		})

		It("passes through cpuWeight, memoryMB and diskMB", func() {
			config.SetBlobStore("blob-host", "7474", "dav-user", "dav-pass")
			Expect(config.Save()).To(Succeed())

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", map[string]string{}, 1, 2, 3)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
			createTaskParams := fakeTaskRunner.CreateTaskArgsForCall(0)
			Expect(createTaskParams).ToNot(BeNil())
			receptorRequest := createTaskParams.GetReceptorRequest()

			Expect(receptorRequest.MemoryMB).To(Equal(1))
			Expect(receptorRequest.CPUWeight).To(Equal(uint(2)))
			Expect(receptorRequest.DiskMB).To(Equal(3))
		})

		It("returns an error when create task fails", func() {
			fakeTaskRunner.CreateTaskReturns(errors.New("creating task failed"))

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", map[string]string{}, 0, 0, 0)

			Expect(err).To(MatchError("creating task failed"))
			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
		})
	})

	Describe("LaunchDroplet", func() {
		BeforeEach(func() {
			config.SetBlobStore("blob-host", "7474", "dav-user", "dav-pass")
			Expect(config.Save()).To(Succeed())
		})

		It("launches the droplet lrp task with a start command from buildpack results", func() {
			executionMetadata := `{"execution_metadata": "{\"start_command\": \"start\"}"}`
			fakeBlobStore.DownloadReturns(ioutil.NopCloser(strings.NewReader(executionMetadata)), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams).ToNot(BeNil())

			Expect(createAppParams.Name).To(Equal("app-name"))
			Expect(createAppParams.RootFS).To(Equal(droplet_runner.DropletRootFS))
			Expect(createAppParams.StartCommand).To(Equal("/tmp/launcher"))
			Expect(createAppParams.AppArgs).To(Equal([]string{"/home/vcap/app", "", `{"start_command": "start"}`}))

			Expect(createAppParams.Annotation).To(MatchJSON(`{
				"droplet_source": {
					"host": "blob-host",
					"port": "7474",
					"droplet_name": "droplet-name"
				}
			}`))

			Expect(createAppParams.Setup).To(Equal(&models.SerialAction{
				LogSource: "app-name",
				Actions: []models.Action{
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
						To:   "/tmp",
						User: "vcap",
					},
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
						To:   "/tmp",
						User: "vcap",
					},
					&models.DownloadAction{
						From: "http://dav-user:dav-pass@blob-host:7474/blobs/droplet-name/droplet.tgz",
						To:   "/home/vcap",
						User: "vcap",
					},
				},
			}))
		})

		It("launches the droplet lrp task with a custom start command", func() {
			executionMetadata := `{"execution_metadata": "{\"start_command\": \"start\"}"}`
			fakeBlobStore.DownloadReturns(ioutil.NopCloser(strings.NewReader(executionMetadata)), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "start-r-up", []string{"-yeah!"}, app_runner.AppEnvironmentParams{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams).ToNot(BeNil())

			Expect(createAppParams.Name).To(Equal("app-name"))
			Expect(createAppParams.StartCommand).To(Equal("/tmp/launcher"))
			Expect(createAppParams.AppArgs).To(Equal([]string{"/home/vcap/app", "start-r-up -yeah!", `{"start_command": "start"}`}))
		})

		It("returns an error when it can't retrieve the execution metadata from the blob store", func() {
			fakeBlobStore.DownloadReturns(nil, errors.New("nope"))

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).To(MatchError("nope"))
		})

		It("returns an error when the downloaded execution metadata is invaild JSON", func() {
			fakeBlobStore.DownloadReturns(ioutil.NopCloser(strings.NewReader("invalid JSON")), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).To(MatchError("invalid character 'i' looking for beginning of value"))
		})

		It("returns an error when create app fails", func() {
			fakeBlobStore.DownloadReturns(ioutil.NopCloser(strings.NewReader(`{}`)), nil)
			fakeAppRunner.CreateAppReturns(errors.New("nope"))

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).To(MatchError("nope"))
		})
	})

	Describe("RemoveDroplet", func() {
		It("recursively removes a droplets from the blob store", func() {
			config.SetBlobStore("blob-host", "7474", "dav-user", "dav-pass")
			Expect(config.Save()).To(Succeed())

			fakeBlobStore.ListReturns([]dav_blob_store.Blob{
				{Path: "drippy/bits.zip"},
				{Path: "drippy/droplet.tgz"},
				{Path: "drippy/result.json"},
			}, nil)

			appInfos := []app_examiner.AppInfo{
				{
					Annotation: "",
				},
				{
					Annotation: `{
						"droplet_source": {
							"host": "other-blob-host",
							"port": "7474",
							"droplet_name": "drippy"
						}
					}`,
				},
				{
					Annotation: `{
						"droplet_source": {
							"host": "blob-host",
							"port": 1234,
							"droplet_name": "drippy"
						}
					}`,
				},
				{
					Annotation: "junk",
				},
				{
					Annotation: `{
						"droplet_source": {
							"host": "blob-host",
							"port": "7474",
							"droplet_name": "other-drippy"
						}
					}`,
				},
			}
			fakeAppExaminer.ListAppsReturns(appInfos, nil)

			Expect(dropletRunner.RemoveDroplet("drippy")).To(Succeed())

			Expect(fakeBlobStore.ListCallCount()).To(Equal(1))

			Expect(fakeBlobStore.DeleteCallCount()).To(Equal(3))
			Expect(fakeBlobStore.DeleteArgsForCall(0)).To(Equal("drippy/bits.zip"))
			Expect(fakeBlobStore.DeleteArgsForCall(1)).To(Equal("drippy/droplet.tgz"))
			Expect(fakeBlobStore.DeleteArgsForCall(2)).To(Equal("drippy/result.json"))

		})

		It("returns an error when querying the blob store fails", func() {
			fakeBlobStore.ListReturns(nil, errors.New("some error"))

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).To(MatchError("some error"))
		})

		It("returns an error when the app specifies that the droplet is in use", func() {
			config.SetBlobStore("blob-host", "7474", "dav-user", "dav-pass")
			Expect(config.Save()).To(Succeed())

			appInfos := []app_examiner.AppInfo{{
				ProcessGuid: "dripapp",
				Annotation: `{
					"droplet_source": {
						"host": "blob-host",
						"port": "7474",
						"droplet_name": "drippy"
					}
				}`,
			}}
			fakeAppExaminer.ListAppsReturns(appInfos, nil)

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).To(MatchError("app dripapp was launched from droplet"))
		})

		It("returns an error when listing the running applications fails", func() {
			fakeAppExaminer.ListAppsReturns(nil, errors.New("some error"))

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).To(MatchError("some error"))
		})

		It("returns an error when the droplet doesn't exist", func() {
			fakeBlobStore.ListReturns([]dav_blob_store.Blob{
				{Path: "drippy/bits.zip"},
				{Path: "drippy/droplet.tgz"},
				{Path: "drippy/result.json"},
			}, nil)

			err := dropletRunner.RemoveDroplet("droopy")
			Expect(err).To(MatchError("droplet not found"))
		})
	})

	Describe("ExportDroplet", func() {
		BeforeEach(func() {
			fakeDropletReader := ioutil.NopCloser(strings.NewReader("some droplet reader"))
			fakeMetadataReader := ioutil.NopCloser(strings.NewReader("some metadata reader"))

			fakeBlobStore.DownloadStub = func(path string) (io.ReadCloser, error) {
				switch path {
				case "drippy/droplet.tgz":
					return fakeDropletReader, nil
				case "drippy/result.json":
					return fakeMetadataReader, nil
				case "no-such-droplet/droplet.tgz":
					return nil, errors.New("some missing droplet error")
				case "no-such-metadata/droplet.tgz":
					return fakeDropletReader, nil
				case "no-such-metadata/result.json":
					return nil, errors.New("some missing metadata error")
				default:
					return nil, errors.New("fake GetReader called with invalid arguments")
				}
			}
		})

		It("returns IO readers for the droplet and its metadata", func() {
			dropletReader, metadataReader, err := dropletRunner.ExportDroplet("drippy")
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.ReadAll(dropletReader)).To(BeEquivalentTo("some droplet reader"))
			Expect(ioutil.ReadAll(metadataReader)).To(BeEquivalentTo("some metadata reader"))
		})

		Context("when the droplet name does not have an associated droplet", func() {
			It("returns an error", func() {
				_, _, err := dropletRunner.ExportDroplet("no-such-droplet")
				Expect(err).To(MatchError("droplet not found: some missing droplet error"))
			})
		})

		Context("when the droplet name does not have an associated metadata", func() {
			It("returns an error", func() {
				_, _, err := dropletRunner.ExportDroplet("no-such-metadata")
				Expect(err).To(MatchError("metadata not found: some missing metadata error"))
			})
		})
	})

	Describe("ImportDroplet", func() {
		var tmpDir, dropletPathArg, metadataPathArg string

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir(os.TempDir(), "droplet")
			Expect(err).NotTo(HaveOccurred())

			dropletPathArg = filepath.Join(tmpDir, "totally-drippy.tgz")
			metadataPathArg = filepath.Join(tmpDir, "result.json")
			Expect(ioutil.WriteFile(dropletPathArg, []byte("droplet contents"), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(metadataPathArg, []byte("result metadata"), 0644)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		Context("when the droplet files exist", func() {
			It("uploads the droplet files to the blob store", func() {
				err := dropletRunner.ImportDroplet("drippy", dropletPathArg, metadataPathArg)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBlobStore.UploadCallCount()).To(Equal(2))

				path, contents := fakeBlobStore.UploadArgsForCall(0)
				Expect(path).To(Equal("drippy/droplet.tgz"))
				Expect(ioutil.ReadAll(contents)).To(Equal([]byte("droplet contents")))

				path, contents = fakeBlobStore.UploadArgsForCall(1)
				Expect(path).To(Equal("drippy/result.json"))
				Expect(ioutil.ReadAll(contents)).To(Equal([]byte("result metadata")))
			})

			Context("when the blob bucket returns error(s)", func() {
				It("returns an error uploading the droplet file", func() {
					fakeBlobStore.UploadReturns(errors.New("some error"))

					err := dropletRunner.ImportDroplet("drippy", dropletPathArg, metadataPathArg)
					Expect(err).To(MatchError("some error"))
				})

				It("returns an error uploading the metadata file", func() {
					fakeBlobStore.UploadStub = func(path string, contents io.ReadSeeker) error {
						if strings.HasSuffix(path, "result.json") {
							return errors.New("some error")
						}
						return nil
					}

					err := dropletRunner.ImportDroplet("drippy", dropletPathArg, metadataPathArg)
					Expect(err).To(MatchError("some error"))
				})
			})
		})

		Context("when the droplet files do not exist", func() {
			It("returns an error opening the droplet file", func() {
				err := dropletRunner.ImportDroplet("drippy", "some/missing/droplet/path", metadataPathArg)
				Expect(err).To(MatchError("open some/missing/droplet/path: no such file or directory"))
			})

			It("returns an error opening the metadata file", func() {
				err := dropletRunner.ImportDroplet("drippy", dropletPathArg, "some/missing/metadata/path")
				Expect(err).To(MatchError("open some/missing/metadata/path: no such file or directory"))
			})
		})
	})
})
