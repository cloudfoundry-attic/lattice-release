package s3_blob_store

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

type BlobStore struct {
	Bucket string
	S3     *s3.S3
}

func New(blobTarget config_package.S3BlobStoreConfig) *BlobStore {
	awsRegion := blobTarget.Region
	client := s3.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(blobTarget.AccessKey, blobTarget.SecretKey, ""),
		Region:      &awsRegion,
	})

	return &BlobStore{
		Bucket: blobTarget.BucketName,
		S3:     client,
	}
}

func (b *BlobStore) List() ([]blob_store.Blob, error) {
	objects, err := b.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b.Bucket),
	})
	if err != nil {
		return nil, err
	}

	blobs := []blob_store.Blob{}
	for _, obj := range objects.Contents {
		blobs = append(blobs, blob_store.Blob{
			Path:    *obj.Key,
			Size:    *obj.Size,
			Created: *obj.LastModified,
		})
	}

	return blobs, nil
}

func (b *BlobStore) Upload(path string, contents io.ReadSeeker) error {
	_, err := b.S3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(b.Bucket),
		ACL:    aws.String("private"),
		Key:    aws.String(path),
		Body:   contents,
	})
	return err
}

func (b *BlobStore) Download(path string) (io.ReadCloser, error) {
	output, err := b.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(path),
	})
	return output.Body, err
}

func (b *BlobStore) Delete(path string) error {
	_, err := b.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(path),
	})
	return err
}
