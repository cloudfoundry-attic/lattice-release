package droplet_runner

import (
	"os"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/goamz/goamz/s3"
)

//go:generate counterfeiter -o fake_droplet_runner/fake_droplet_runner.go . DropletRunner
type DropletRunner interface {
	UploadBits(dropletName, uploadPath string) error
}

type dropletRunner struct {
	blobStore  blob_store.BlobStore
	blobBucket blob_store.BlobBucket
}

func New(blobStore blob_store.BlobStore, blobBucket blob_store.BlobBucket) *dropletRunner {
	return &dropletRunner{
		blobStore:  blobStore,
		blobBucket: blobBucket,
	}
}

func (dr *dropletRunner) UploadBits(dropletName, uploadPath string) error {
	fileInfo, err := os.Stat(uploadPath)
	if err != nil {
		return err
	}

	uploadFile, err := os.Open(uploadPath)
	if err != nil {
		return err
	}

	return dr.blobBucket.PutReader(dropletName, uploadFile, fileInfo.Size(), blob_store.DropletContentType, blob_store.DefaultPrivilege, s3.Options{})
}
