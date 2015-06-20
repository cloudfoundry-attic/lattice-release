package target_verifier

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

var (
	awsRegion = aws.Region{Name: "riak-region-1", S3Endpoint: "http://s3.amazonaws.com"}
)

func (t *targetVerifier) VerifyBlobTarget(host string, port uint16, accessKey, secretKey, bucketName string) (bool, error) {
	s3Auth := aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	config := config.New(persister.NewMemPersister())
	config.SetBlobTarget(host, port, accessKey, secretKey, bucketName)

	s3S3 := s3.New(s3Auth, awsRegion, &http.Client{
		Transport: &http.Transport{
			Proxy: config.BlobTarget().Proxy(),
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	})
	s3S3.AttemptStrategy = aws.AttemptStrategy{}

	blobStore := blob_store.NewBlobStore(config, s3S3)
	blobBucket := blobStore.Bucket(config.BlobTarget().BucketName)

	if _, err := blobBucket.List("", "/", "", 1); err != nil {
		httpError, ok := err.(*s3.Error)
		if ok {
			switch httpError.StatusCode {
			case 200:
				return true, nil
			case 403:
				return false, fmt.Errorf("unauthorized")
			default:
				return false, fmt.Errorf("%s", httpError)
			}
		}

		return false, fmt.Errorf("blob target is down: %s", err)
	}

	return true, nil
}
