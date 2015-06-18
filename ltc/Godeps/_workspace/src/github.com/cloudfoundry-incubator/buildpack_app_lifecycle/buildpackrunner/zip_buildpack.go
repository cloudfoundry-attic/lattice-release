package buildpackrunner

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/cacheddownloader"
	"github.com/pivotal-golang/archiver/extractor"
)

type ZipDownloader struct {
	downloader *cacheddownloader.Downloader
}

func IsZipFile(filename string) bool {
	return strings.HasSuffix(filename, ".zip")
}

func NewZipDownloader(skipSSLVerification bool) *ZipDownloader {
	return &ZipDownloader{
		downloader: cacheddownloader.NewDownloader(DOWNLOAD_TIMEOUT, 1, skipSSLVerification),
	}
}

func (z *ZipDownloader) DownloadAndExtract(u *url.URL, destination string) (uint64, error) {
	zipFile, err := ioutil.TempFile("", filepath.Base(u.Path))
	if err != nil {
		return 0, fmt.Errorf("Could not create zip file: %s", err.Error())
	}
	defer os.Remove(zipFile.Name())

	_, _, err = z.downloader.Download(u,
		func() (*os.File, error) {
			return os.OpenFile(zipFile.Name(), os.O_WRONLY, 0666)
		},
		cacheddownloader.CachingInfoType{},
		make(chan struct{}),
	)
	if err != nil {
		return 0, fmt.Errorf("Failed to download buildpack '%s': %s", u.String(), err.Error())
	}

	fi, err := zipFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("Failed to obtain the size of the buildpack '%s': %s", u.String(), err.Error())
	}

	err = extractor.NewZip().Extract(zipFile.Name(), destination)
	if err != nil {
		return 0, fmt.Errorf("Failed to extract buildpack '%s': %s", u.String(), err.Error())
	}

	return uint64(fi.Size()), nil
}
