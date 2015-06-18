package droplet_runner_test

import (
	"errors"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store/fake_blob_bucket"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store/fake_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner/fake_task_runner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("DropletRunner", func() {

	var (
		fakeTaskRunner *fake_task_runner.FakeTaskRunner
		config         *config_package.Config
		fakeBlobStore  *fake_blob_store.FakeBlobStore
		fakeBlobBucket *fake_blob_bucket.FakeBlobBucket
		dropletRunner  droplet_runner.DropletRunner
	)

	BeforeEach(func() {
		fakeTaskRunner = new(fake_task_runner.FakeTaskRunner)
		config = config_package.New(persister.NewMemPersister())
		fakeBlobStore = new(fake_blob_store.FakeBlobStore)
		fakeBlobBucket = new(fake_blob_bucket.FakeBlobBucket)
		dropletRunner = droplet_runner.New(fakeTaskRunner, config, fakeBlobStore, fakeBlobBucket)
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
			})

			AfterEach(func() {
				Expect(os.Remove(tmpFile.Name())).To(Succeed())
			})

			It("uploads the file to the bucket", func() {
				err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBlobBucket.PutReaderCallCount()).To(Equal(1))
				path, reader, length, contType, perm, options := fakeBlobBucket.PutReaderArgsForCall(0)
				Expect(path).To(Equal("droplet-name/bits.tgz"))
				Expect(reader).ToNot(BeNil())
				Expect(length).ToNot(BeZero())
				Expect(contType).To(Equal(blob_store.DropletContentType))
				Expect(perm).To(Equal(blob_store.DefaultPrivilege))
				Expect(options).To(BeZero())
			})

			It("errors when Bucket.PutReader fails", func() {
				fakeBlobBucket.PutReaderReturns(errors.New("winter is coming yo"))

				err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

				Expect(err).To(MatchError("winter is coming yo"))
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

			err := dropletRunner.BuildDroplet("droplet-name", "buildpack")

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
			Expect(receptorRequest.TaskGuid).To(Equal("droplet-name"))
			Expect(receptorRequest.LogGuid).To(Equal("droplet-name"))
			Expect(receptorRequest.MetricsGuid).To(Equal("droplet-name"))
			Expect(receptorRequest.RootFS).To(Equal("preloaded:cflinuxfs2"))
			Expect(receptorRequest.LogSource).To(Equal("BUILD"))
			Expect(receptorRequest.Domain).To(Equal("lattice"))
			Expect(receptorRequest.EgressRules).ToNot(BeNil())
			Expect(receptorRequest.EgressRules).To(BeEmpty())
		})

		It("returns an error when create task fails", func() {
			fakeTaskRunner.CreateTaskReturns(errors.New("creating task failed"))

			err := dropletRunner.BuildDroplet("droplet-name", "buildpack")

			Expect(err).To(MatchError("creating task failed"))
			Expect(fakeTaskRunner.CreateTaskCallCount()).To(Equal(1))
		})
	})

})
