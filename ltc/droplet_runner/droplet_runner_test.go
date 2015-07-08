package droplet_runner_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
		dropletRunner      droplet_runner.DropletRunner
	)

	BeforeEach(func() {
		fakeAppRunner = new(fake_app_runner.FakeAppRunner)
		fakeTaskRunner = new(fake_task_runner.FakeTaskRunner)
		config = config_package.New(persister.NewMemPersister())
		fakeBlobStore = new(fake_blob_store.FakeBlobStore)
		fakeBlobBucket = new(fake_blob_bucket.FakeBlobBucket)
		fakeTargetVerifier = new(fake_target_verifier.FakeTargetVerifier)
		dropletRunner = droplet_runner.New(fakeAppRunner, fakeTaskRunner, config, fakeBlobStore, fakeBlobBucket, fakeTargetVerifier)
	})

	Describe("ListDroplets", func() {
		It("returns a list of droplets in the blob store", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			fakeBlobBucket.ListStub = func(prefix, delim, marker string, max int) (result *s3.ListResp, err error) {
				if prefix == "" {
					return &s3.ListResp{
						Name:           "bucket-name",
						Prefix:         "",
						Delimiter:      "/",
						CommonPrefixes: []string{"X/", "Y/", "Z/"},
					}, nil
				} else if prefix == "X/" {
					return &s3.ListResp{
						Name:      "bucket-name",
						Prefix:    "X/",
						Delimiter: "/",
						Contents: []s3.Key{
							s3.Key{Key: "X/bits.tgz", LastModified: "2006-01-02T15:04:05.999Z"},
							s3.Key{Key: "X/droplet.tgz", LastModified: "2006-01-02T15:04:05.999Z"},
							s3.Key{Key: "X/result.json", LastModified: "2006-01-02T15:04:05.999Z"},
						},
					}, nil
				} else if prefix == "Y/" {
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
				} else if prefix == "Z/" {
					return &s3.ListResp{
						Name:      "bucket-name",
						Prefix:    "Z/",
						Delimiter: "/",
						Contents: []s3.Key{
							s3.Key{Key: "Z/bits.tgz"},
						},
					}, nil
				} else {
					panic("no stub for arguments: " + prefix + "," + delim + "," + marker + "," + string(max))
				}
			}

			droplets, err := dropletRunner.ListDroplets()

			Expect(err).NotTo(HaveOccurred())
			Expect(len(droplets)).To(Equal(2))
			Expect(droplets[0].Name).To(Equal("X"))
			Expect((*droplets[0].Created).Unix()).To(Equal(time.Date(2006, 1, 2, 15, 4, 5, 999, time.UTC).Unix()))
			Expect(droplets[1].Name).To(Equal("Y"))
			Expect(droplets[1].Created).To(BeNil())
		})

		It("returns an error when querying the blob store fails", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

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
			// name doesn't match file descriptor
			osFile := os.NewFile(os.Stdout.Fd(), "new-file-yo")

			err := dropletRunner.UploadBits("droplet-name", osFile.Name())

			Expect(err).To(HaveOccurred())
			Expect(fakeBlobBucket.PutReaderCallCount()).To(BeZero())
		})

		// FIXME: This strategy doesn't work when run as root on CI.
		//
		// It("errors when file can be Stat'ed but not Opened", func() {
		// 	tmpFile, err := ioutil.TempFile(os.TempDir(), "stat")
		// 	Expect(err).ToNot(HaveOccurred())

		// 	Expect(os.Chmod(tmpFile.Name(), 0)).To(Succeed())

		// 	err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

		// 	Expect(err).To(HaveOccurred())
		// 	Expect(fakeBlobBucket.PutReaderCallCount()).To(BeZero())
		// })
	})

	Describe("BuildDroplet", func() {
		It("does the build droplet task", func() {
			config.SetBlobTarget("blob-host", 7474, "access-key", "secret-key", "bucket-name")
			config.Save()

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
			createTaskParams := fakeTaskRunner.CreateTaskArgsForCall(0)
			Expect(createTaskParams).ToNot(BeNil())
			receptorRequest := createTaskParams.GetReceptorRequest()

			expectedActions := &models.SerialAction{
				Actions: []models.Action{
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-support.tgz",
						To:   "/tmp",
					},
					&models.RunAction{
						Path: "/tmp/s3downloader",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/bits.tgz", "/tmp/bits.tgz"},
					},
					&models.RunAction{
						Path: "/bin/mkdir",
						Dir:  "/",
						Args: []string{"/tmp/app"},
					},
					&models.RunAction{
						Path: "/bin/tar",
						Dir:  "/",
						Args: []string{"-C", "/tmp/app", "-xf", "/tmp/bits.tgz"},
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
							"-skipDetect=false",
						},
					},
					&models.RunAction{
						Path: "/tmp/s3uploader",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/droplet.tgz", "/tmp/droplet"},
					},
					&models.RunAction{
						Path: "/tmp/s3uploader",
						Dir:  "/",
						Args: []string{"access-key", "secret-key", "http://blob-host:7474/", "bucket-name", "droplet-name/result.json", "/tmp/result.json"},
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

		It("returns an error when create task fails", func() {
			fakeTaskRunner.CreateTaskReturns(errors.New("creating task failed"))

			err := dropletRunner.BuildDroplet("task-name", "droplet-name", "buildpack")

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
			js := []byte(`{"detected_start_command":{"web":"start"}}`)
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(bytes.NewReader(js)), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams).ToNot(BeNil())

			Expect(createAppParams.Name).To(Equal("app-name"))
			Expect(createAppParams.RootFS).To(Equal(droplet_runner.DropletRootFS))
			Expect(createAppParams.StartCommand).To(Equal("/tmp/lrp-launcher"))
			Expect(createAppParams.AppArgs).To(Equal([]string{"start"}))

			Expect(createAppParams.Setup).To(Equal(&models.SerialAction{
				LogSource: "app-name",
				Actions: []models.Action{
					&models.DownloadAction{
						From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-support.tgz",
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
					},
					&models.RunAction{
						Path: "/bin/tar",
						Dir:  "/home/vcap",
						Args: []string{"-zxf", "/tmp/droplet.tgz"},
					},
				},
			}))
		})

		It("launches the droplet lrp task with the droplet name as the start command", func() {
			js := []byte(`{}`)
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(bytes.NewReader(js)), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams).ToNot(BeNil())

			Expect(createAppParams.Name).To(Equal("app-name"))
			Expect(createAppParams.StartCommand).To(Equal("/tmp/lrp-launcher"))
			Expect(createAppParams.AppArgs).To(Equal([]string{"droplet-name"}))
		})

		It("launches the droplet lrp task with a custom start command", func() {
			js := []byte(`{}`)
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(bytes.NewReader(js)), nil)

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "start-r-up", []string{"-yeah!"}, app_runner.AppEnvironmentParams{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAppRunner.CreateAppCallCount()).To(Equal(1))
			createAppParams := fakeAppRunner.CreateAppArgsForCall(0)
			Expect(createAppParams).ToNot(BeNil())

			Expect(createAppParams.Name).To(Equal("app-name"))
			Expect(createAppParams.StartCommand).To(Equal("start-r-up"))
			Expect(createAppParams.AppArgs).To(Equal([]string{"-yeah!"}))
		})

		It("returns an error when it can't download result.json from the blob store", func() {
			fakeBlobBucket.GetReaderReturns(nil, errors.New("nope"))

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})

			Expect(err).To(HaveOccurred())
		})

		It("returns an error when create app fails", func() {
			fakeBlobBucket.GetReaderReturns(ioutil.NopCloser(bytes.NewReader([]byte(`{}`))), nil)
			fakeAppRunner.CreateAppReturns(errors.New("nope"))

			err := dropletRunner.LaunchDroplet("app-name", "droplet-name", "", []string{}, app_runner.AppEnvironmentParams{})

			Expect(err).To(HaveOccurred())
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
	})
})
