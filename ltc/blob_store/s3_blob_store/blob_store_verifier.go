package s3_blob_store

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

type Verifier struct {
	Endpoint string
}

func (v *Verifier) Verify(config *config_package.Config) (authorized bool, err error) {
	blobStoreConfig := config.S3BlobStore()
	client := s3.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(blobStoreConfig.AccessKey, blobStoreConfig.SecretKey, ""),
		Region:           aws.String(blobStoreConfig.Region),
		S3ForcePathStyle: aws.Bool(true),
	})
	if v.Endpoint != "" {
		client.Endpoint = v.Endpoint
	}
	_, err = client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(blobStoreConfig.BucketName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok && awsErr.StatusCode() == 403 {
			return false, nil
		}

		return false, err
	}
	return true, nil
}
