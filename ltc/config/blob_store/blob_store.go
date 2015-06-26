package blob_store

import (
	"io"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/goamz/goamz/s3"
)

const (
	DropletContentType = "application/octet-stream"
)

var (
	DefaultPrivilege s3.ACL = s3.Private
)

//go:generate counterfeiter -o fake_blob_store/fake_blob_store.go . BlobStore
type BlobStore interface {
	Bucket(name string) BlobBucket
}

//go:generate counterfeiter -o fake_blob_bucket/fake_blob_bucket.go . BlobBucket
type BlobBucket interface {
	List(prefix, delim, marker string, max int) (result *s3.ListResp, err error)
	PutReader(path string, r io.Reader, length int64, contType string, perm s3.ACL, options s3.Options) error
	GetReader(path string) (rc io.ReadCloser, err error)
	Del(path string) error
}

type blobStore struct {
	config     *config.Config
	s3Endpoint *s3.S3
}

func NewBlobStore(config *config.Config, s3S3 *s3.S3) *blobStore {
	return &blobStore{
		config:     config,
		s3Endpoint: s3S3,
	}
}

func (bs *blobStore) Bucket(name string) BlobBucket {
	return bs.s3Endpoint.Bucket(name)
}
