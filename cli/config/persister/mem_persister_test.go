package persister_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
)

var _ = Describe("memPersister", func() {
	type data struct {
		Value string
	}

	type dataTwo struct {
		OtherValue string
	}

	It("Loads and Saves data in memory", func() {
		memPersister := persister.NewMemPersister()

		dataToSave := &data{Value: "Save?"}
		err := memPersister.Save(dataToSave)
		Expect(err).ToNot(HaveOccurred())

		dataToLoad := &data{}
		err = memPersister.Load(dataToLoad)
		Expect(err).ToNot(HaveOccurred())

		Expect(dataToLoad.Value).To(Equal("Save?"))
	})

	It("returns errors when using complex types", func() {
		memPersister := persister.NewMemPersister()

		err := memPersister.Save(func() {})
		Expect(err).To(HaveOccurred())

		err = memPersister.Load(func() {})
		Expect(err).To(HaveOccurred())
	})
})
