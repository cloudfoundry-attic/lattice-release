package persister_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
)

var _ = Describe("filePersister", func() {
	type data struct {
		Value string
	}

	var (
		tmpDir  string
		tmpFile *os.File
		err     error
	)

	BeforeEach(func() {
		tmpDir = os.TempDir()

		tmpFile, err = ioutil.TempFile(tmpDir, "tmp_file")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpFile.Name())).To(Succeed())
	})

	Describe("Load", func() {

		var (
			filePersister persister.Persister
			dataToRead    *data
		)

		BeforeEach(func() {
			dataToRead = &data{}
			filePersister = persister.NewFilePersister(tmpFile.Name())
		})

		JustBeforeEach(func() {
			err = filePersister.Load(dataToRead)
		})

		It("Loads empty data from an empty file", func() {
			Expect(dataToRead.Value).To(BeEmpty())
		})

		Context("when the file already exists", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value"}`), 0700)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Loads JSON from the file", func() {
				Expect(dataToRead.Value).To(Equal("test value"))
			})
		})

		Context("when the file has invalid json", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value`), 0700)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns errors from invalid JSON", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when reading the file returns an error", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(tmpFile.Name(), []byte(""), 0000)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns errors from reading the file", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when reading nonexistant files", func() {
			BeforeEach(func() {
				nonExistantFile := fmt.Sprintf("%s/nonexistant/tmp_file", tmpDir)
				Expect(os.RemoveAll(nonExistantFile)).To(Succeed())

				filePersister = persister.NewFilePersister(nonExistantFile)
			})

			It("handles nonexistant files silently", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Save", func() {

		var (
			filePersister persister.Persister
			dataToSave    *data
		)

		BeforeEach(func() {
			dataToSave = &data{Value: "Some Value to be written in json"}
			filePersister = persister.NewFilePersister(tmpFile.Name())
		})

		JustBeforeEach(func() {
			err = filePersister.Save(dataToSave)
		})

		It("Saves valid JSON to the filepath", func() {
			jsonBytes, err := ioutil.ReadFile(tmpFile.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonBytes).To(MatchJSON(`{"Value":"Some Value to be written in json"}`))
		})

		It("Returns an error rather than save invalid JSON", func() {
			err := filePersister.Save(func() {})
			Expect(err).To(MatchError(ContainSubstring("unsupported type")))
		})

		Context("when reading nonexistant files", func() {

			var nonExistantFile string

			BeforeEach(func() {
				nonExistantFile = filepath.Join(tmpDir, "nonexistant", "tmp_file")
				Expect(os.RemoveAll(filepath.Dir(nonExistantFile))).To(Succeed())

				filePersister = persister.NewFilePersister(nonExistantFile)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(filepath.Dir(nonExistantFile))).To(Succeed())
			})

			It("writes to nonexistant directories", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when making the directory", func() {
			BeforeEach(func() {
				filePath := filepath.Join(tmpFile.Name(), "no_privs", "tmp_file")
				if _, err := os.Stat(tmpFile.Name()); err != nil {
					Expect(os.IsNotExist(err)).To(BeFalse())
				}

				filePersister = persister.NewFilePersister(filePath)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tmpFile.Name())).To(Succeed())
			})

			It("returns errors from making the directory", func() {
				Expect(err).To(MatchError(ContainSubstring("not a directory")))
			})
		})

		Context("when writing the file returns errors", func() {
			BeforeEach(func() {
				filePersister = persister.NewFilePersister(tmpDir)
			})

			It("returns errors from writing the file", func() {
				Expect(err).To(MatchError(ContainSubstring("is a directory")))
			})
		})

	})
})
