package buildpackrunner_test

import (
	"archive/zip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle/buildpackrunner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ZipBuildpack", func() {
	var destination string

	BeforeEach(func() {
		var err error
		destination, err = ioutil.TempDir("", "unzipdir")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(destination)
	})

	Describe("IsZipFile", func() {
		It("returns true with .zip extension", func() {
			Expect(buildpackrunner.IsZipFile("abc.zip")).To(BeTrue())
		})

		It("returns false without .zip extension", func() {
			Expect(buildpackrunner.IsZipFile("abc.tar")).To(BeFalse())
		})
	})

	Describe("DownloadZipAndExtract", func() {
		var fileserver *httptest.Server
		var zipDownloader *buildpackrunner.ZipDownloader

		BeforeEach(func() {
			zipDownloader = buildpackrunner.NewZipDownloader(false)
			fileserver = httptest.NewServer(http.FileServer(http.Dir(os.TempDir())))
		})

		AfterEach(func() {
			fileserver.Close()
		})

		Context("with a valid zip file", func() {
			var zipfile string
			var zipSize uint64

			BeforeEach(func() {
				var err error
				z, err := ioutil.TempFile("", "zipfile")
				Expect(err).NotTo(HaveOccurred())
				zipfile = z.Name()

				w := zip.NewWriter(z)
				f, err := w.Create("contents")
				Expect(err).NotTo(HaveOccurred())
				f.Write([]byte("stuff"))
				err = w.Close()
				Expect(err).NotTo(HaveOccurred())
				fi, err := z.Stat()
				Expect(err).NotTo(HaveOccurred())
				zipSize = uint64(fi.Size())
			})

			AfterEach(func() {
				os.Remove(zipfile)
			})

			It("downloads and extracts", func() {
				u, _ := url.Parse(fileserver.URL)
				u.Path = filepath.Base(zipfile)
				size, err := zipDownloader.DownloadAndExtract(u, destination)
				Expect(err).NotTo(HaveOccurred())
				Expect(size).To(Equal(zipSize))
				file, err := os.Open(filepath.Join(destination, "contents"))
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				bytes, err := ioutil.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())
				Expect(bytes).To(Equal([]byte("stuff")))
			})
		})

		It("fails when the zip file does not exist", func() {
			u, _ := url.Parse("file:///foobar_not_there")
			size, err := zipDownloader.DownloadAndExtract(u, destination)
			Expect(err).To(HaveOccurred())
			Expect(size).To(Equal(uint64(0)))
		})

		It("fails when the file is not a zip file", func() {
			u, _ := url.Parse(fileserver.URL)
			u.Path = filepath.Base(destination)
			size, err := zipDownloader.DownloadAndExtract(u, destination)
			Expect(err).To(HaveOccurred())
			Expect(size).To(Equal(uint64(0)))
		})
	})
})
