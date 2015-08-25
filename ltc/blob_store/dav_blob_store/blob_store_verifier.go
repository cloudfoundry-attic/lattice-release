package dav_blob_store

import (
	"fmt"
	"net/http"
	"net/url"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

type Verifier struct{}

func (Verifier) Verify(config *config_package.Config) (authorized bool, err error) {
	blobStoreURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", config.BlobStore().Host, config.BlobStore().Port),
		User:   url.UserPassword(config.BlobStore().Username, config.BlobStore().Password),
	}

	baseURL := &url.URL{
		Scheme: blobStoreURL.Scheme,
		Host:   blobStoreURL.Host,
		User:   blobStoreURL.User,
		Path:   "/blobs/",
	}

	req, err := http.NewRequest("PROPFIND", baseURL.String(), nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Depth", "1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	return resp.StatusCode == 207, err
}
