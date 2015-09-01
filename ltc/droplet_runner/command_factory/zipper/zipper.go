package zipper

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore"
)

//go:generate counterfeiter -o fake_zipper/fake_zipper.go . Zipper
type Zipper interface {
	Zip(srcDir string, cfIgnore cf_ignore.CFIgnore) (string, error)
	IsZipFile(path string) bool
	Unzip(srcZip string, destDir string) error
}

type DropletArtifactZipper struct{}

func (zipper *DropletArtifactZipper) Zip(srcDir string, cfIgnore cf_ignore.CFIgnore) (string, error) {
	tmpPath, err := ioutil.TempDir(os.TempDir(), "droplet-artifact-zipper")
	if err != nil {
		return "", err
	}

	fileWriter, err := os.OpenFile(filepath.Join(tmpPath, "artifact.zip"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(fileWriter)
	defer zipWriter.Close()

	contentsFileInfo, err := os.Stat(srcDir)
	if err != nil {
		return "", err
	}

	if !contentsFileInfo.IsDir() {
		return "", fmt.Errorf("%s must be a directory", srcDir)
	}

	if ignoreFile, err := os.Open(filepath.Join(srcDir, ".cfignore")); err == nil {
		if err := cfIgnore.Parse(ignoreFile); err != nil {
			return "", err
		}
	}

	err = filepath.Walk(srcDir, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(srcDir, fullPath)
		if err != nil {
			return err
		}

		if cfIgnore.ShouldIgnore(relativePath) {
			return nil
		}

		if relativePath == fileWriter.Name() || relativePath == "." || relativePath == ".." {
			return nil
		}

		if h, err := zip.FileInfoHeader(info); err == nil {
			h.Name = relativePath

			if info.IsDir() {
				h.Name = h.Name + "/"
			}

			h.SetMode(info.Mode())

			writer, err := zipWriter.CreateHeader(h)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			li, err := os.Lstat(fullPath)
			if err != nil {
				return err
			}

			if li.Mode()&os.ModeSymlink == os.ModeSymlink {
				dest, err := os.Readlink(fullPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(writer, strings.NewReader(dest)); err != nil {
					return err
				}
			} else {
				fr, err := os.Open(fullPath)
				if err != nil {
					return err
				}
				defer fr.Close()
				if _, err := io.Copy(writer, fr); err != nil {
					return err
				}
			}
		}

		return nil
	})

	return fileWriter.Name(), err
}

func (zipper *DropletArtifactZipper) IsZipFile(path string) bool {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return false
	} else {
		reader.Close()
		return true
	}
}

func (zipper *DropletArtifactZipper) Unzip(srcZip string, destDir string) error {
	reader, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		err := func() error {
			fileReader, err := f.Open()
			if err != nil {
				return err
			}
			defer fileReader.Close()

			fileInfo := f.FileHeader.FileInfo()

			if fileInfo.IsDir() {
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(filepath.Join(destDir, f.FileHeader.Name)), os.ModeDir|os.ModePerm); err != nil {
				return err
			}

			fileWriter, err := os.OpenFile(filepath.Join(destDir, f.FileHeader.Name), os.O_CREATE|os.O_WRONLY, f.FileHeader.Mode())
			if err != nil {
				return err
			}
			defer fileWriter.Close()

			_, err = io.Copy(fileWriter, fileReader)
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
