package dav_blob_store

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/blob"
	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type BlobStore struct {
	URL    *url.URL
	Client *http.Client
}

func New(config config_package.BlobStoreConfig) *BlobStore {
	return &BlobStore{
		URL: &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%s", config.Host, config.Port),
			User:   url.UserPassword(config.Username, config.Password),
		},
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}

type xmlTime time.Time

func (t *xmlTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(time.RFC1123, v)
	if err != nil {
		return err
	}
	*t = xmlTime(parse)
	return nil
}

type listResponse struct {
	Responses []struct {
		HREF          string  `xml:"href"`
		LastModified  xmlTime `xml:"propstat>prop>getlastmodified"`
		ContentLength int64   `xml:"propstat>prop>getcontentlength"`
	} `xml:"response"`
}

func (b *BlobStore) doListRequest(baseURL *url.URL) (listResponse, error) {
	req, err := http.NewRequest("PROPFIND", baseURL.String(), nil)
	if err != nil {
		return listResponse{}, err
	}

	req.Header.Add("Depth", "1")

	resp, err := b.Client.Do(req)
	if err != nil {
		return listResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 207 {
		return listResponse{}, errors.New(resp.Status)
	}

	decoder := xml.NewDecoder(resp.Body)

	var listResp listResponse
	if err := decoder.Decode(&listResp); err != nil {
		return listResponse{}, err
	}

	return listResp, nil
}

func (b *BlobStore) listBlobFiles(baseURL *url.URL) ([]blob.Blob, error) {
	listResp, err := b.doListRequest(baseURL)
	if err != nil {
		return nil, err
	}

	var blobFiles []blob.Blob
	for _, resp := range listResp.Responses {
		u, err := url.Parse(resp.HREF)
		if err != nil {
			return nil, err
		}

		if path.Clean(u.Path) == path.Clean(baseURL.Path) {
			continue
		}

		blobFiles = append(blobFiles, blob.Blob{
			Path:    strings.Replace(path.Clean(u.Path), "/blobs/", "", 1),
			Created: time.Time(resp.LastModified),
			Size:    resp.ContentLength,
		})
	}

	return blobFiles, nil
}

func (b *BlobStore) List() ([]blob.Blob, error) {
	baseURL := &url.URL{
		Scheme: b.URL.Scheme,
		Host:   b.URL.Host,
		User:   b.URL.User,
		Path:   "/blobs",
	}

	listResp, err := b.doListRequest(baseURL)
	if err != nil {
		return nil, err
	}

	var blobs []blob.Blob

	for _, resp := range listResp.Responses {
		u, err := url.Parse(resp.HREF)
		if err != nil {
			return nil, err
		}

		u.User = b.URL.User

		if path.Clean(baseURL.Path) == path.Clean(u.Path) {
			continue
		}

		blobFiles, err := b.listBlobFiles(u)
		if err != nil {
			return nil, err
		}

		blobs = append(blobs, blobFiles...)
	}

	return blobs, nil
}

func (b *BlobStore) ensureParentCollectionExists(baseURL *url.URL) error {
	parentURL := &url.URL{
		Scheme: baseURL.Scheme,
		Host:   baseURL.Host,
		User:   baseURL.User,
		Path:   path.Dir(baseURL.Path),
	}
	_, err := b.listBlobFiles(parentURL)
	if err == nil {
		return nil
	}
	if err.Error() != "404 Not Found" {
		return err
	}

	req, err := http.NewRequest("MKCOL", parentURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New(resp.Status)
	}

	return nil
}

func (b *BlobStore) Upload(path string, contents io.ReadSeeker) error {
	baseURL := &url.URL{
		Scheme: b.URL.Scheme,
		Host:   b.URL.Host,
		User:   b.URL.User,
		Path:   "/blobs/" + path,
	}

	if err := b.ensureParentCollectionExists(baseURL); err != nil {
		return err
	}

	length, err := contents.Seek(0, 2)
	if err != nil {
		return err
	}
	contents.Seek(0, 0)

	req, err := http.NewRequest("PUT", baseURL.String(), contents)
	if err != nil {
		return err
	}

	req.ContentLength = length

	resp, err := b.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (b *BlobStore) Download(path string) (io.ReadCloser, error) {
	baseURL := &url.URL{
		Scheme: b.URL.Scheme,
		Host:   b.URL.Host,
		User:   b.URL.User,
		Path:   "/blobs/" + path,
	}

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return resp.Body, nil
}

func (b *BlobStore) Delete(path string) error {
	baseURL := &url.URL{
		Scheme: b.URL.Scheme,
		Host:   b.URL.Host,
		User:   b.URL.User,
		Path:   "/blobs/" + path,
	}

	req, err := http.NewRequest("DELETE", baseURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func (b *BlobStore) DownloadAppBitsAction(dropletName string) models.Action {
	return &models.DownloadAction{
		From: b.URL.String() + "/blobs/" + dropletName + "/bits.zip",
		To:   "/tmp/app",
		User: "vcap",
	}
}

func (b *BlobStore) DeleteAppBitsAction(dropletName string) models.Action {
	return &models.RunAction{
		Path: "/tmp/davtool",
		Dir:  "/",
		Args: []string{"delete", b.URL.String() + "/blobs/" + dropletName + "/bits.zip"},
		User: "vcap",
	}
}

func (b *BlobStore) UploadDropletAction(dropletName string) models.Action {
	return &models.RunAction{
		Path: "/tmp/davtool",
		Dir:  "/",
		Args: []string{"put", b.URL.String() + "/blobs/" + dropletName + "/droplet.tgz", "/tmp/droplet"},
		User: "vcap",
	}

}

func (b *BlobStore) UploadDropletMetadataAction(dropletName string) models.Action {
	return &models.RunAction{
		Path: "/tmp/davtool",
		Dir:  "/",
		Args: []string{"put", b.URL.String() + "/blobs/" + dropletName + "/result.json", "/tmp/result.json"},
		User: "vcap",
	}
}

func (b *BlobStore) DownloadDropletAction(dropletName string) models.Action {
	return &models.DownloadAction{
		From: b.URL.String() + "/blobs/" + dropletName + "/droplet.tgz",
		To:   "/home/vcap",
		User: "vcap",
	}

}
