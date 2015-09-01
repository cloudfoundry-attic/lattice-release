package zipper_test

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore/fake_cf_ignore"
	zipper_package "github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/zipper"
)

var _ = Describe("Zipper", func() {
	var zipper zipper_package.Zipper

	BeforeEach(func() {
		zipper = &zipper_package.DropletArtifactZipper{}
	})

	Describe("#Zip", func() {
		var (
			prevDir, tmpDir string
			err             error

			fakeCFIgnore *fake_cf_ignore.FakeCFIgnore
		)

		BeforeEach(func() {
			fakeCFIgnore = &fake_cf_ignore.FakeCFIgnore{}

			tmpDir, err = ioutil.TempDir(os.TempDir(), "zip_contents")
			Expect(err).NotTo(HaveOccurred())

			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "aaa"), []byte("aaa contents"), 0700)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "bbb"), []byte("bbb contents"), 0750)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "ccc"), []byte("ccc contents"), 0644)).To(Succeed())
			Expect(os.Symlink("ccc", filepath.Join(tmpDir, "ddd"))).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "some-ignored-file"), []byte("ignored contents"), 0644)).To(Succeed())

			Expect(os.Mkdir(filepath.Join(tmpDir, "subfolder"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "subfolder", "sub"), []byte("sub contents"), 0644)).To(Succeed())

			prevDir, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chdir(tmpDir)).To(Succeed())

			fakeCFIgnore.ShouldIgnoreStub = func(path string) bool {
				return path == "some-ignored-file"
			}
		})

		AfterEach(func() {
			Expect(os.Chdir(prevDir)).To(Succeed())
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("zips successfully", func() {
			zipPath, err := zipper.Zip(tmpDir, fakeCFIgnore)
			Expect(err).NotTo(HaveOccurred())

			zipReader, err := zip.OpenReader(zipPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(zipReader.File).To(HaveLen(6))

			buffer := make([]byte, 12)
			h := zipReader.File[0].FileHeader
			f, err := zipReader.File[0].Open()
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			Expect(h.Name).To(Equal("aaa"))
			Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0700)))
			Expect(f.Read(buffer)).To(Equal(12))
			Expect(string(buffer)).To(Equal("aaa contents"))

			buffer = make([]byte, 12)
			h = zipReader.File[1].FileHeader
			f, err = zipReader.File[1].Open()
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			Expect(h.Name).To(Equal("bbb"))
			Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0750)))
			Expect(f.Read(buffer)).To(Equal(12))
			Expect(string(buffer)).To(Equal("bbb contents"))

			buffer = make([]byte, 12)
			h = zipReader.File[2].FileHeader
			f, err = zipReader.File[2].Open()
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			Expect(h.Name).To(Equal("ccc"))
			Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
			Expect(f.Read(buffer)).To(Equal(12))
			Expect(string(buffer)).To(Equal("ccc contents"))

			buffer = make([]byte, 3)
			h = zipReader.File[3].FileHeader
			f, err = zipReader.File[3].Open()
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			Expect(h.Name).To(Equal("ddd"))
			Expect(h.FileInfo().Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
			Expect(f.Read(buffer)).To(Equal(3))
			Expect(string(buffer)).To(Equal("ccc"))

			buffer = make([]byte, 1)
			h = zipReader.File[4].FileHeader
			f, err = zipReader.File[4].Open()
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			Expect(h.Name).To(Equal("subfolder/"))
			Expect(h.FileInfo().IsDir()).To(BeTrue())
			Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(os.ModeDir | 0755)))
			_, err = f.Read(buffer)
			Expect(err).To(MatchError("EOF"))

			buffer = make([]byte, 12)
			h = zipReader.File[5].FileHeader
			f, err = zipReader.File[5].Open()
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			Expect(h.Name).To(Equal("subfolder/sub"))
			Expect(h.FileInfo().Mode()).To(Equal(os.FileMode(0644)))
			Expect(f.Read(buffer)).To(Equal(12))
			Expect(string(buffer)).To(Equal("sub contents"))
		})

		Context("failure", func() {
			It("returns an error if passed a non-directory", func() {
				_, err := zipper.Zip(filepath.Join(tmpDir, "aaa"), fakeCFIgnore)
				Expect(err).To(MatchError(fmt.Sprintf("%s must be a directory", filepath.Join(tmpDir, "aaa"))))
			})

			It("returns an error if .cfignore can't be parsed", func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, ".cfignore"), []byte{}, 0600)).To(Succeed())
				fakeCFIgnore.ParseReturns(errors.New("no"))
				_, err := zipper.Zip(tmpDir, fakeCFIgnore)
				Expect(err).To(MatchError("no"))
			})
		})
	})

	Describe("#Unzip", func() {
		var (
			prevDir, tmpDir string
			err             error
			tmpFile         *os.File
			prevUmask       int
		)

		BeforeEach(func() {
			prevUmask = syscall.Umask(0)

			tmpDir, err = ioutil.TempDir(os.TempDir(), "unzip_contents")
			Expect(err).NotTo(HaveOccurred())

			prevDir, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chdir(tmpDir)).To(Succeed())

			tmpFile, err = ioutil.TempFile("", "zipfile")
			Expect(err).NotTo(HaveOccurred())
			defer tmpFile.Close()

			zipWriter := zip.NewWriter(tmpFile)
			defer zipWriter.Close()

			var (
				w      io.Writer
				length int
				header *zip.FileHeader
			)

			header = &zip.FileHeader{Name: "aaa"}
			header.SetMode(os.FileMode(0644))
			w, err = zipWriter.CreateHeader(header)
			Expect(err).NotTo(HaveOccurred())
			length, err = w.Write([]byte("aaaaa"))
			Expect(length).To(Equal(5))
			Expect(err).NotTo(HaveOccurred())

			header = &zip.FileHeader{Name: "bbb/1.txt"}
			header.SetMode(os.FileMode(0640))
			w, err = zipWriter.CreateHeader(header)
			Expect(err).NotTo(HaveOccurred())
			length, err = w.Write([]byte("one"))
			Expect(length).To(Equal(3))
			Expect(err).NotTo(HaveOccurred())

			header = &zip.FileHeader{Name: "bbb/2.txt"}
			header.SetMode(os.FileMode(0600))
			w, err = zipWriter.CreateHeader(header)
			Expect(err).NotTo(HaveOccurred())
			length, err = w.Write([]byte("twoo"))
			Expect(length).To(Equal(4))
			Expect(err).NotTo(HaveOccurred())

			header = &zip.FileHeader{Name: "ddd/3.txt"}
			header.SetMode(os.FileMode(0600))
			w, err = zipWriter.CreateHeader(header)
			Expect(err).NotTo(HaveOccurred())
			length, err = w.Write([]byte("three"))
			Expect(length).To(Equal(5))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(tmpFile.Name())
			syscall.Umask(prevUmask)
		})

		It("unzips", func() {
			Expect(zipper.Unzip(tmpFile.Name(), tmpDir)).To(Succeed())

			var (
				contents []byte
				err      error
				fileInfo os.FileInfo
			)

			contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "aaa"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("aaaaa"))

			fileInfo, err = os.Stat(filepath.Join(tmpDir, "aaa"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))

			fileInfo, err = os.Stat(filepath.Join(tmpDir, "bbb"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode().Perm()).To(Equal(os.FileMode(0777)))

			contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "bbb/1.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("one"))

			fileInfo, err = os.Stat(filepath.Join(tmpDir, "bbb/1.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0640)))

			contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "bbb/2.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("twoo"))

			fileInfo, err = os.Stat(filepath.Join(tmpDir, "bbb/2.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0600)))

			_, err = os.Stat(filepath.Join(tmpDir, "ccc"))
			Expect(err).To(HaveOccurred())

			contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "ddd/3.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("three"))

			fileInfo, err = os.Stat(filepath.Join(tmpDir, "ddd"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode().Perm()).To(Equal(os.FileMode(0777)))
		})
	})

	Describe("#IsZipFile", func() {
		It("accepts zip files", func() {
			minimalZipBytes := []byte{'P', 'K', 0x05, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

			tmpFile, err := ioutil.TempFile(os.TempDir(), "emptyzip")
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.WriteFile(tmpFile.Name(), minimalZipBytes, 0700)).To(Succeed())
			defer os.Remove(tmpFile.Name())

			Expect(zipper.IsZipFile(tmpFile.Name())).To(BeTrue())
		})

		It("rejects non-zip files", func() {
			tmpFile, err := ioutil.TempFile(os.TempDir(), "badzip")
			Expect(err).NotTo(HaveOccurred())
			Expect(ioutil.WriteFile(tmpFile.Name(), []byte("I promise I'm a zip file"), 0700)).To(Succeed())
			defer os.Remove(tmpFile.Name())

			Expect(zipper.IsZipFile(tmpFile.Name())).To(BeFalse())
		})
	})
})
