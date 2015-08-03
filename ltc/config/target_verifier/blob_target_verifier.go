package target_verifier

import (
	"fmt"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
)

// TODO: should be separate object
func (t *targetVerifier) VerifyBlobTarget(targetInfo dav_blob_store.Config) error {
	blobStore := dav_blob_store.New(targetInfo)

	if _, err := blobStore.List(); err != nil {
		if err.Error() == "401 Unauthorized" || err.Error() == "403 Forbidden" {
			return fmt.Errorf("unauthorized")
		} else {
			return fmt.Errorf("blob target is down: %s", err)
		}
	}

	return nil
}
