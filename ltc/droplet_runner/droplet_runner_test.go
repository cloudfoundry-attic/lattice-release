package droplet_runner_test

import (
	"errors"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store/fake_blob_bucket"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store/fake_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
)

var _ = Describe("DropletRunner", func() {

	var (
		fakeBlobStore  *fake_blob_store.FakeBlobStore
		fakeBlobBucket *fake_blob_bucket.FakeBlobBucket
		dropletRunner  droplet_runner.DropletRunner
	)

	BeforeEach(func() {
		fakeBlobStore = new(fake_blob_store.FakeBlobStore)
		fakeBlobBucket = new(fake_blob_bucket.FakeBlobBucket)
		dropletRunner = droplet_runner.New(fakeBlobStore, fakeBlobBucket)
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
				Expect(path).To(Equal("droplet-name"))
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

		It("errors when file can be Stat'ed but not Opened", func() {
			tmpFile, err := ioutil.TempFile(os.TempDir(), "stat")
			Expect(err).ToNot(HaveOccurred())

			Expect(os.Chmod(tmpFile.Name(), 0)).To(Succeed())

			err = dropletRunner.UploadBits("droplet-name", tmpFile.Name())

			Expect(err).To(HaveOccurred())
			Expect(fakeBlobBucket.PutReaderCallCount()).To(BeZero())
		})
	})

})
