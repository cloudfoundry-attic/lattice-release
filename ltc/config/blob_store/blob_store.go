package blob_store

import (
	"io"
	"net/http"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/goamz/goamz/s3"
)

const (
	BucketName     = "condenser-bucket"
	TarContentType = "application/x-tar"
)

var (
	DefaultPrivilege s3.ACL = s3.PublicReadWrite
)

//go:generate counterfeiter -o fake_blob_store/fake_blob_store.go . BlobStore
type BlobStore interface {
	Bucket(name string) BlobBucket
	S3Endpoint() *s3.S3 // TODO: development only
}

//go:generate counterfeiter -o fake_blob_bucket/fake_blob_bucket.go . BlobBucket
type BlobBucket interface {
	Head(path string, headers map[string][]string) (*http.Response, error)
	Put(path string, data []byte, contType string, perm s3.ACL, options s3.Options) error
	PutReader(path string, r io.Reader, length int64, contType string, perm s3.ACL, options s3.Options) error
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

func (bs *blobStore) S3Endpoint() *s3.S3 {
	return bs.s3Endpoint
}

func (bs *blobStore) Bucket(name string) BlobBucket {
	return bs.s3Endpoint.Bucket(name)
}
