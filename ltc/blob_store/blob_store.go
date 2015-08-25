package blob_store

import (
	"io"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/blob"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/s3_blob_store"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type BlobStore interface {
	List() ([]blob.Blob, error)
	Delete(path string) error
	Upload(path string, contents io.ReadSeeker) error
	Download(path string) (io.ReadCloser, error)

	DropletStore
}

type DropletStore interface {
	DownloadAppBitsAction(dropletName string) models.Action
	DeleteAppBitsAction(dropletName string) models.Action
	UploadDropletAction(dropletName string) models.Action
	UploadDropletMetadataAction(dropletName string) models.Action
	DownloadDropletAction(dropletName string) models.Action
}

type Verifier interface {
	Verify(config *config_package.Config) (authorized bool, err error)
}

func New(config *config_package.Config) BlobStore {
	switch config.ActiveBlobStore() {
	case config_package.DAVBlobStore:
		return dav_blob_store.New(config.BlobStore())
	case config_package.S3BlobStore:
		return s3_blob_store.New(config.S3BlobStore())
	}

	return dav_blob_store.New(config.BlobStore())
}

func NewVerifier(config *config_package.Config) Verifier {
	switch config.ActiveBlobStore() {
	case config_package.DAVBlobStore:
		return dav_blob_store.Verifier{}
	case config_package.S3BlobStore:
		return &s3_blob_store.Verifier{}
	}

	return dav_blob_store.Verifier{}
}
