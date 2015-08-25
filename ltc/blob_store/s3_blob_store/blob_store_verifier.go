package s3_blob_store

import (
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

type Verifier struct{}

func (Verifier) Verify(config *config_package.Config) (authorized bool, err error) {
	return true, nil
}
