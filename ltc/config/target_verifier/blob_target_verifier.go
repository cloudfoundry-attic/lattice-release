package target_verifier

import (
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
	awsRegion = aws.Region{Name: "faux-region-1", S3Endpoint: "http://s3.amazonaws.com"}
)

func (t *targetVerifier) VerifyBlobTarget(host string, port uint16, accessKey, secretKey string) (bool, bool, error) {
	s3Auth := aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	config := config.New(persister.NewMemPersister())
	config.SetBlobTarget(host, port, accessKey, secretKey)

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

	blobStore := blob_store.NewBlobStore(config, s3S3)
	blobBucket := blobStore.Bucket(blob_store.BucketName)

	if _, err := blobBucket.Head("/verifier/invalid-path", map[string][]string{}); err != nil {
		httpError, ok := err.(*s3.Error)
		if ok {
			if httpError.StatusCode == 403 {
				return true, false, nil
			} else {
				return true, true, nil
			}
		}

		return false, false, err
	}

	return true, true, nil
}
