package s3_blob_store

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/blob"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type BlobStore struct {
	Bucket     string
	S3         *s3.S3
	blobTarget config_package.S3BlobStoreConfig
}

func New(blobTarget config_package.S3BlobStoreConfig) *BlobStore {
	client := s3.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(blobTarget.AccessKey, blobTarget.SecretKey, ""),
		Region:           aws.String(blobTarget.Region),
		S3ForcePathStyle: aws.Bool(true),
	})

	return &BlobStore{
		Bucket:     blobTarget.BucketName,
		S3:         client,
		blobTarget: blobTarget,
	}
}

func (b *BlobStore) List() ([]blob.Blob, error) {
	objects, err := b.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b.Bucket),
	})
	if err != nil {
		return nil, err
	}

	blobs := []blob.Blob{}
	for _, obj := range objects.Contents {
		blobs = append(blobs, blob.Blob{
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

func (b *BlobStore) DownloadAppBitsAction(dropletName string) models.Action {
	return &models.SerialAction{
		Actions: []models.Action{
			&models.RunAction{
				Path: "/tmp/s3tool",
				Dir:  "/",
				Args: []string{
					"get",
					b.blobTarget.AccessKey,
					b.blobTarget.SecretKey,
					b.Bucket,
					b.blobTarget.Region,
					"/" + dropletName + "/bits.zip",
					"/tmp/bits.zip",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/bin/mkdir",
				Args: []string{"/tmp/app"},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/usr/bin/unzip",
				Dir:  "/tmp/app",
				Args: []string{"-q", "/tmp/bits.zip"},
				User: "vcap",
			},
		},
	}
}

func (b *BlobStore) DeleteAppBitsAction(dropletName string) models.Action {
	return &models.RunAction{
		Path: "/tmp/s3tool",
		Dir:  "/",
		Args: []string{
			"delete",
			b.blobTarget.AccessKey,
			b.blobTarget.SecretKey,
			b.Bucket,
			b.blobTarget.Region,
			"/" + dropletName + "/bits.zip",
		},
		User: "vcap",
	}
}

func (b *BlobStore) UploadDropletAction(dropletName string) models.Action {
	return &models.RunAction{
		Path: "/tmp/s3tool",
		Dir:  "/",
		Args: []string{
			"put",
			b.blobTarget.AccessKey,
			b.blobTarget.SecretKey,
			b.Bucket,
			b.blobTarget.Region,
			"/" + dropletName + "/droplet.tgz",
			"/tmp/droplet",
		},
		User: "vcap",
	}
}

func (b *BlobStore) UploadDropletMetadataAction(dropletName string) models.Action {
	return &models.RunAction{
		Path: "/tmp/s3tool",
		Dir:  "/",
		Args: []string{
			"put",
			b.blobTarget.AccessKey,
			b.blobTarget.SecretKey,
			b.Bucket,
			b.blobTarget.Region,
			"/" + dropletName + "/result.json",
			"/tmp/result.json",
		},
		User: "vcap",
	}
}

func (b *BlobStore) DownloadDropletAction(dropletName string) models.Action {
	return &models.SerialAction{
		Actions: []models.Action{
			&models.RunAction{
				Path: "/tmp/s3tool",
				Dir:  "/",
				Args: []string{
					"get",
					b.blobTarget.AccessKey,
					b.blobTarget.SecretKey,
					b.Bucket,
					b.blobTarget.Region,
					"/" + dropletName + "/droplet.tgz",
					"/tmp/droplet.tgz",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/bin/tar",
				Args: []string{"zxf", "/tmp/droplet.tgz"},
				Dir:  "/home/vcap",
				User: "vcap",
			},
		},
	}
}
