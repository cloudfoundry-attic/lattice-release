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
)

type BlobStore struct {
	url url.URL
}

type Blob struct {
	Path    string
	Created time.Time
	Size    int64
}

type Config struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func New(config Config) *BlobStore {
	return &BlobStore{
		url: url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%s", config.Host, config.Port),
			User:   url.UserPassword(config.Username, config.Password),
		},
	}
}

type xmlTime struct {
	time.Time
}

func (t *xmlTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(time.RFC1123, v)
	if err != nil {
		return err
	}
	*t = xmlTime{parse}
	return nil
}

type listResponse struct {
	Responses []struct {
		Href          string  `xml:"href"`
		LastModified  xmlTime `xml:"propstat>prop>getlastmodified"`
		ContentLength int64   `xml:"propstat>prop>getcontentlength"`
	} `xml:"response"`
}

func doListRequest(baseURL *url.URL) (listResponse, error) {
	req, err := http.NewRequest("PROPFIND", baseURL.String(), nil)
	if err != nil {
		return listResponse{}, err
	}

	req.Header.Add("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
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

func listBlobFiles(baseURL *url.URL) ([]Blob, error) {
	listResp, err := doListRequest(baseURL)
	if err != nil {
		return nil, err
	}

	var blobFiles []Blob

	for _, resp := range listResp.Responses {
		u, err := url.Parse(resp.Href)
		if err != nil {
			return nil, err
		}

		if path.Clean(u.Path) == path.Clean(baseURL.Path) {
			continue
		}

		blobFiles = append(blobFiles, Blob{
			Path:    strings.Replace(path.Clean(u.Path), "/blobs/", "", 1),
			Created: resp.LastModified.Time,
			Size:    resp.ContentLength,
		})
	}

	return blobFiles, nil
}

func (b *BlobStore) List() ([]Blob, error) {
	baseURL := &url.URL{
		Scheme: b.url.Scheme,
		Host:   b.url.Host,
		User:   b.url.User,
		Path:   "/blobs",
	}

	listResp, err := doListRequest(baseURL)
	if err != nil {
		return nil, err
	}

	var blobs []Blob

	for _, resp := range listResp.Responses {
		u, err := url.Parse(resp.Href)
		if err != nil {
			return nil, err
		}

		u.User = b.url.User

		if path.Clean(baseURL.Path) == path.Clean(u.Path) {
			continue
		}

		blobFiles, err := listBlobFiles(u)
		if err != nil {
			return nil, err
		}

		blobs = append(blobs, blobFiles...)
	}

	return blobs, nil
}

func ensureParentCollectionExists(baseURL *url.URL) error {
	parentURL := &url.URL{
		Scheme: baseURL.Scheme,
		Host:   baseURL.Host,
		User:   baseURL.User,
		Path:   path.Dir(baseURL.Path),
	}
	_, err := listBlobFiles(parentURL)
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

	resp, err := http.DefaultClient.Do(req)
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
		Scheme: b.url.Scheme,
		Host:   b.url.Host,
		User:   b.url.User,
		Path:   "/blobs/" + path,
	}

	if err := ensureParentCollectionExists(baseURL); err != nil {
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

	resp, err := http.DefaultClient.Do(req)
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
		Scheme: b.url.Scheme,
		Host:   b.url.Host,
		User:   b.url.User,
		Path:   "/blobs/" + path,
	}

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
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
		Scheme: b.url.Scheme,
		Host:   b.url.Host,
		User:   b.url.User,
		Path:   "/blobs/" + path,
	}

	req, err := http.NewRequest("DELETE", baseURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
