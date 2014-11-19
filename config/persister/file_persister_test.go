package persister_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/diego-edge-cli/config/persister"
)

type data struct {
	Value string
}

var _ = Describe("filePersister", func() {
	var (
		tmpDir  string
		tmpFile *os.File
	)

	BeforeEach(func() {
		var err error
		tmpDir = os.TempDir()

		tmpFile, err = ioutil.TempFile(tmpDir, "tmp_file")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Load", func() {
		It("Loads empty data from an empty file", func() {
			persister := persister.NewFilePersister(tmpFile.Name())
			data := &data{}

			persister.Load(data)

			Expect(data.Value).To(Equal(""))
		})

		It("Loads JSON from the file", func() {
			err := ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value"}`), 0700)
			Expect(err).ToNot(HaveOccurred())

			persister := persister.NewFilePersister(tmpFile.Name())
			data := &data{}

			persister.Load(data)

			Expect(data.Value).To(Equal("test value"))
		})

		It("returns errors from invalid JSON", func() {
			err := ioutil.WriteFile(tmpFile.Name(), []byte(`{"Value":"test value`), 0700)
			Expect(err).ToNot(HaveOccurred())

			persister := persister.NewFilePersister(tmpFile.Name())

			err = persister.Load(&data{})

			Expect(err).ToNot(BeNil())
		})

		It("handles nonexistant files silently", func() {
			nonExistantFile := fmt.Sprintf("%snonexistant/tmp_file", tmpDir)
			persister := persister.NewFilePersister(nonExistantFile)

			err := persister.Load(&data{})

			Expect(err).To(BeNil())
		})
	})

	Describe("Save", func() {
		It("Saves valid JSON to the filepath", func() {
			persister := persister.NewFilePersister(tmpFile.Name())

			persister.Save(&data{Value: "Some Value to be written in json"})
			jsonBytes, err := ioutil.ReadFile(tmpFile.Name())
			Expect(err).ToNot(HaveOccurred())

			Expect(string(jsonBytes)).To(Equal(`{"Value":"Some Value to be written in json"}`))
		})

		It("Returns an error rather than save invalid JSON", func() {
			persister := persister.NewFilePersister(tmpFile.Name())
			err := persister.Save(func() {})

			Expect(err).To(HaveOccurred())

		})

		It("writes to nonexistant directories", func() {
			nonExistantFile := fmt.Sprintf("%snonexistant/tmp_file", tmpDir)
			persister := persister.NewFilePersister(nonExistantFile)

			err := persister.Save(&data{"Some Value"})

			Expect(err).To(BeNil())
		})

		It("Returns errors from writing the file", func() {
			persister := persister.NewFilePersister(tmpDir)
			err := persister.Save(&data{})

			Expect(err).ToNot(BeNil())
		})
	})
})
