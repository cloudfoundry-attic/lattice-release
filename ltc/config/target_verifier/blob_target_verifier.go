package target_verifier

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/s3_blob_store"
)

// TODO: should be separate object
func (t *targetVerifier) VerifyBlobTarget(targetInfo config.BlobTargetInfo) error {
	blobStore := s3_blob_store.New(targetInfo)
	// TODO: seems like it would be better to retry
	blobStore.S3.ShouldRetry = func(_ *aws.Request) bool { return false }

	if _, err := blobStore.List(); err != nil {
		if httpError, ok := err.(awserr.RequestFailure); ok {
			if httpError.StatusCode() == 403 {
				return fmt.Errorf("unauthorized")
			}
			return httpError
		}

		return fmt.Errorf("blob target is down: %s", err)
	}

	return nil
}
