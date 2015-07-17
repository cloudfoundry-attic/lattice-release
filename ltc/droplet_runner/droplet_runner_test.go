package droplet_runner_test

import (
	"errors"
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
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store/fake_blob_bucket"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store/fake_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier/fake_target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner/fake_task_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/matchers"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/goamz/goamz/s3"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("DropletRunner", func() {
	var (
		fakeAppRunner      *fake_app_runner.FakeAppRunner
		fakeTaskRunner     *fake_task_runner.FakeTaskRunner
		config             *config_package.Config
		fakeBlobStore      *fake_blob_store.FakeBlobStore
		fakeBlobBucket     *fake_blob_bucket.FakeBlobBucket
		fakeTargetVerifier *fake_target_verifier.FakeTargetVerifier
		fakeAppExaminer    *fake_app_examiner.FakeAppExaminer
		dropletRunner      droplet_runner.DropletRunner
	)

	BeforeEach(func() {
		fakeAppRunner = &fake_app_runner.FakeAppRunner{}
		fakeTaskRunner = &fake_task_runner.FakeTaskRunner{}
		config = config_package.New(persister.NewMemPersister())
		fakeBlobStore = &fake_blob_store.FakeBlobStore{}
		fakeBlobBucket = &fake_blob_bucket.FakeBlobBucket{}
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		fakeAppExaminer = &fake_app_examiner.FakeAppExaminer{}
		dropletRunner = droplet_runner.New(fakeAppRunner, fakeTaskRunner, config, fakeBlobStore, fakeBlobBucket, fakeTargetVerifier, fakeAppExaminer)
	})

	Describe("ListDroplets", func() {
		It("returns a list of droplets in the blob store", func() {
			fakeBlobBucket.ListStub = func(prefix, delim, marker string, max int) (result *s3.ListResp, err error) {
				switch prefix {
				case "":
					return &s3.ListResp{
						Name:           "bucket-name",
						Prefix:         "",
						Delimiter:      "/",
						CommonPrefixes: []string{"X/", "Y/", "Z/"},
					}, nil
				case "X/":
					return &s3.ListResp{
						Name:      "bucket-name",
						Prefix:    "X/",
						Delimiter: "/",
						Contents: []s3.Key{
							s3.Key{Key: "X/bits.tgz", LastModified: "2006-01-02T15:04:05.999Z", Size: 100},
							s3.Key{Key: "X/droplet.tgz", LastModified: "2006-01-02T15:04:05.999Z", Size: 200},
							s3.Key{Key: "X/result.json", LastModified: "2006-01-02T15:04:05.999Z", Size: 300},
						},
					}, nil
				case "Y/":
					return &s3.ListResp{
						Name:      "bucket-name",
						Prefix:    "Y/",
						Delimiter: "/",
						Contents: []s3.Key{
							s3.Key{Key: "Y/bits.tgz"},
							s3.Key{Key: "Y/droplet.tgz"},
							s3.Key{Key: "Y/result.json"},
						},
					}, nil
				case "Z/":
					return &s3.ListResp{
						Name:      "bucket-name",
						Prefix:    "Z/",
						Delimiter: "/",
						Contents: []s3.Key{
							s3.Key{Key: "Z/bits.tgz"},
						},
					}, nil
				}

				Fail("no stub for arguments: " + prefix + "," + delim + "," + marker + "," + string(max))
				return nil, nil
			}

			droplets, err := dropletRunner.ListDroplets()

			Expect(err).NotTo(HaveOccurred())
			Expect(len(droplets)).To(Equal(2))
			Expect(droplets[0].Name).To(Equal("X"))
			Expect(droplets[0].Created.Unix()).To(Equal(time.Date(2006, 1, 2, 15, 4, 5, 999, time.UTC).Unix()))
			Expect(droplets[0].Size).To(Equal(int64(200)))
			Expect(droplets[1].Name).To(Equal("Y"))
			Expect(droplets[1].Created).To(BeZero())
			Expect(droplets[1].Size).To(Equal(int64(0)))
		})

		It("returns an error when querying the blob store fails", func() {
			fakeBlobBucket.ListReturns(nil, errors.New("boom"))

			_, err := dropletRunner.ListDroplets()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UploadBits", func() {
		Context("when the archive path is a file and exists", func() {
			var (
				tmpFile *os.File
				err     error
			)

			BeforeEach(func() {
				tmpDir := os.TempDir()
				tmpFile, err = ioutil.TempFile(tmpDir, "tmp_file")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value"}`), 0700)
				Expect(err).NotTo(HaveOccurred())

				fakeTargetVerifier.VerifyBlobTargetReturns(true, nil)
			})

			AfterEach(func() {
				Expect(os.Remove(tmpFile.Name())).To(Succeed())
			})

			It("uploads the file to the bucket", func() {
				err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))

				Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(1))
				path, reader, length, contType, perm, options := fakeBlobBucket.PutReaderArgsForCall(0)
				Expect(path).To(Equal("droplet-name/bits.tgz"))
				Expect(reader).ToNot(BeNil())
				Expect(length).ToNot(BeZero())
				Expect(contType).To(Equal(blob_store.DropletContentType))
				Expect(perm).To(Equal(blob_store.DefaultPrivilege))
				Expect(options).To(BeZero())
			})

			It("errors when the blob store verifier fails", func() {
				fakeTargetVerifier.VerifyBlobTargetReturns(false, errors.New("no blobs here"))

				err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

				Expect(err).To(MatchError("no blobs here"))

				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(0))
			})

			It("errors when Bucket.PutReader fails", func() {
				fakeBlobBucket.PutReaderReturns(errors.New("winter is coming yo"))

				err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

				Expect(err).To(MatchError("winter is coming yo"))
				Expect(fakeTargetVerifier.VerifyBlobTargetCallCount()).To(Equal(1))
				Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(1))
			})
		})

		It("errors when file cannot be Stat'ed", func() {
			err := dropletRunner.UploadBits("droplet-name", "new-file-yo")

			Expect(err).To(HaveOccurred())
			Expect(fakeBlobBucket.PutReaderCallCount()).To(BeZero())
		})
	})

	Describe("BuildDroplet", func() {
		It("does the build droplet task", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", map[string]string{})

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
			createTaskParams := fakeTaskRunner.CreateTaskArgsForCall(0)
			Expect(createTaskParams).ToNot(BeNil())
			receptorRequest := createTaskParams.GetReceptorRequest()

			expectedActions := &models.SerialAction{
				Actions: []models.Action{
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
						To:   "/tmp",
					},
					&models.RunAction{
						Path: "/tmp/s3downloader",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/bits.tgz", "/tmp/bits.tgz"},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/bin/mkdir",
						Dir:  "/",
						Args: []string{"/tmp/app"},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/bin/tar",
						Dir:  "/",
						Args: []string{"-C", "/tmp/app", "-xf", "/tmp/bits.tgz"},
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
						Path: "/tmp/s3uploader",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/droplet.tgz", "/tmp/droplet"},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/tmp/s3uploader",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/result.json", "/tmp/result.json"},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/tmp/s3deleter",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/bits.tgz"},
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
				receptor.EnvironmentVariable{
					Name:  "CF_STACK",
					Value: "cflinuxfs2",
				},
			}))
			Expect(receptorRequest.LogSource).To(Equal("BUILD"))
			Expect(receptorRequest.Domain).To(Equal("lattice"))
			Expect(receptorRequest.EgressRules).ToNot(BeNil())
			Expect(receptorRequest.EgressRules).To(BeEmpty())
		})

		It("passes through user environment variables", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			env := map[string]string{
				"ENV_VAR":   "stuff",
				"OTHER_VAR": "same",
			}

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", env)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
			createTaskParams := fakeTaskRunner.CreateTaskArgsForCall(0)
			Expect(createTaskParams).ToNot(BeNil())
			receptorRequest := createTaskParams.GetReceptorRequest()

			Expect(receptorRequest.EnvironmentVariables).To(matchers.ContainExactly([]receptor.EnvironmentVariable{
				receptor.EnvironmentVariable{
					Name:  "CF_STACK",
					Value: "cflinuxfs2",
				},
				receptor.EnvironmentVariable{
					Name:  "ENV_VAR",
					Value: "stuff",
				},
				receptor.EnvironmentVariable{
					Name:  "OTHER_VAR",
					Value: "same",
				},
			}))
		})

		It("returns an error when create task fails", func() {
			fakeTaskRunner.CreateTaskReturns(errors.New("creating task failed"))

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack", map[string]string{})

			Expect(err).To(MatchError("creating task failed"))
			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
		})
	})

	Describe("LaunchDroplet", func() {
		BeforeEach(func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()
		})

		It("launches the droplet lrp task with a start command from buildpack results", func() {
			executionMetadata := `{"execution_metadata": "{\"start_command\": \"start\"}"}`
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(strings.NewReader(executionMetadata)), nil)

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
					"port": 7474,
					"bucket_name": "bucket-name",
					"droplet_name": "droplet-name"
				}
			}`))

			Expect(createAppParams.Setup).To(Equal(&models.SerialAction{
				LogSource: "app-name",
				Actions: []models.Action{
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
						To:   "/tmp",
					},
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
						To:   "/tmp",
					},
					&models.RunAction{
						Path: "/tmp/s3downloader",
						Args: []string{
							"access-key",
							"secret-key",
							"http://blob-host:7474",
							"bucket-name",
							"droplet-name/droplet.tgz",
							"/tmp/droplet.tgz",
						},
						User: "vcap",
					},
					&models.RunAction{
						Path: "/bin/tar",
						Dir:  "/home/vcap",
						Args: []string{"-zxf", "/tmp/droplet.tgz"},
						User: "vcap",
					},
				},
			}))
		})

		It("launches the droplet lrp task with a custom start command", func() {
			executionMetadata := `{"execution_metadata": "{\"start_command\": \"start\"}"}`
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(strings.NewReader(executionMetadata)), nil)

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
			fakeBlobBucket.GetReaderReturns(nil, errors.New("nope"))

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).To(MatchError("nope"))
		})

		It("returns an error when the downloaded execution metadata is invaild JSON", func() {
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(strings.NewReader("invalid JSON")), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).To(MatchError("invalid character 'i' looking for beginning of value"))
		})

		It("returns an error when create app fails", func() {
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(strings.NewReader(`{}`)), nil)
			fakeAppRunner.CreateAppReturns(errors.New("nope"))

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).To(MatchError("nope"))
		})
	})

	Describe("RemoveDroplet", func() {
		It("recursively removes a droplets from the blob store", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			dropletContents := &s3.ListResp{
				Name:      "bucket-name",
				Prefix:    "drippy/",
				Delimiter: "/",
				Contents: []s3.Key{
					s3.Key{Key: "drippy/bits.tgz"},
					s3.Key{Key: "drippy/droplet.tgz"},
					s3.Key{Key: "drippy/result.json"},
				},
			}
			fakeBlobBucket.ListReturns(dropletContents, nil)

			fakeBlobBucket.DelStub = func(path string) error {
				switch path {
				case "drippy/bits.tgz":
					return nil
				case "drippy/droplet.tgz":
					return nil
				case "drippy/result.json":
					return nil
				default:
					return errors.New("bad arg to bucket.Del(): " + path)
				}
			}

			appInfos := []app_examiner.AppInfo{
				{
					Annotation: "",
				},
				{
					Annotation: `{
						"droplet_source": {
							"host": "other-blob-host",
							"port": 7474,
							"bucket_name": "bucket-name",
							"droplet_name": "drippy"
						}
					}`,
				},
				{
					Annotation: `{
						"droplet_source": {
							"host": "blob-host",
							"port": 1234,
							"bucket_name": "bucket-name",
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
							"port": 7474,
							"bucket_name": "other-bucket-name",
							"droplet_name": "drippy"
						}
					}`,
				},
				{
					Annotation: `{
						"droplet_source": {
							"host": "blob-host",
							"port": 7474,
							"bucket_name": "bucket-name",
							"droplet_name": "other-drippy"
						}
					}`,
				},
			}
			fakeAppExaminer.ListAppsReturns(appInfos, nil)

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeBlobBucket.ListCallCount()).To(Equal(1))
			prefix, _, _, _ := fakeBlobBucket.ListArgsForCall(0)
			Expect(prefix).To(Equal("drippy/"))

			Expect(fakeBlobBucket.DelCallCount()).To(Equal(3))
		})

		It("returns an error when querying the blob store fails", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			fakeBlobBucket.ListReturns(nil, errors.New("boom"))

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the app specifies that the droplet is in use", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			dropletContents := &s3.ListResp{
				Name:      "bucket-name",
				Prefix:    "drippy/",
				Delimiter: "/",
				Contents: []s3.Key{
					s3.Key{Key: "drippy/bits.tgz"},
					s3.Key{Key: "drippy/droplet.tgz"},
					s3.Key{Key: "drippy/result.json"},
				},
			}
			fakeBlobBucket.ListReturns(dropletContents, nil)

			appInfos := []app_examiner.AppInfo{{
				ProcessGuid: "dripapp",
				Annotation: `{
					"droplet_source": {
						"host": "blob-host",
						"port": 7474,
						"bucket_name": "bucket-name",
						"droplet_name": "drippy"
					}
				}`,
			}}
			fakeAppExaminer.ListAppsReturns(appInfos, nil)

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).To(MatchError("app dripapp was launched from droplet"))
		})

		It("returns an error when listing the running applications fails", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			fakeAppExaminer.ListAppsReturns(nil, errors.New("some error"))

			err := dropletRunner.RemoveDroplet("drippy")
			Expect(err).To(MatchError("some error"))
		})
	})

	Describe("ExportDroplet", func() {
		BeforeEach(func() {
			fakeDropletReader := ioutil.NopCloser(strings.NewReader("some droplet reader"))
			fakeMetadataReader := ioutil.NopCloser(strings.NewReader("some metadata reader"))

			fakeBlobBucket.GetReaderStub = func(path string) (io.ReadCloser, error) {
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

				Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(2))

				dropletPath, dropletReader, dropletLength, dropletContentType, dropletPerm, dropletOptions := fakeBlobBucket.PutReaderArgsForCall(0)
				Expect(dropletPath).To(Equal("drippy/droplet.tgz"))
				Expect(dropletReader).ToNot(BeNil())
				Expect(dropletLength).ToNot(BeZero())
				Expect(dropletContentType).To(Equal(blob_store.DropletContentType))
				Expect(dropletPerm).To(Equal(blob_store.DefaultPrivilege))
				Expect(dropletOptions).To(BeZero())

				metadataPath, metadataReader, metadataLength, metadataContentType, metadataPerm, metadataOptions := fakeBlobBucket.PutReaderArgsForCall(1)
				Expect(metadataPath).To(Equal("drippy/result.json"))
				Expect(metadataReader).ToNot(BeNil())
				Expect(metadataLength).ToNot(BeZero())
				Expect(metadataContentType).To(Equal(blob_store.DropletContentType))
				Expect(metadataPerm).To(Equal(blob_store.DefaultPrivilege))
				Expect(metadataOptions).To(BeZero())
			})

			Context("when the blob bucket returns error(s)", func() {
				It("returns an error uploading the droplet file", func() {
					fakeBlobBucket.PutReaderReturns(errors.New("unable2upload"))

					err := dropletRunner.ImportDroplet("drippy", dropletPathArg, metadataPathArg)
					Expect(err).To(MatchError("unable2upload"))

					Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(1))
				})

				It("returns an error uploading the metadata file", func() {
					fakeBlobBucket.PutReaderStub = func(path string, r io.Reader, length int64, contType string, perm s3.ACL, options s3.Options) error {
						if strings.HasSuffix(path, "result.json") {
							return errors.New("ugly-duckling")
						}
						return nil
					}

					err := dropletRunner.ImportDroplet("drippy", dropletPathArg, metadataPathArg)
					Expect(err).To(MatchError("ugly-duckling"))

					Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(2))
				})
			})
		})

		Context("when the droplet files do not exist", func() {
			It("returns an error stat'ing the droplet file", func() {
				err := dropletRunner.ImportDroplet("drippy", "some/missing/droplet/path", metadataPathArg)
				Expect(err).To(MatchError("stat some/missing/droplet/path: no such file or directory"))
			})

			It("returns an error stat'ing the metadata file", func() {
				err := dropletRunner.ImportDroplet("drippy", dropletPathArg, "some/missing/metadata/path")
				Expect(err).To(MatchError("stat some/missing/metadata/path: no such file or directory"))
			})
		})
	})
})
