package test_helpers_test

import (
	"os"

	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JsonWriter", func() {
	type TestStruct struct {
		Num int `json:"number"`
	}

	Describe("TempJsonFile", func() {
		var (
			test   TestStruct
			called bool
			buffer []byte
		)

		BeforeEach(func() {
			test = TestStruct{
				Num: 9,
			}

			buffer = make([]byte, 50)
		})

		It("writes the passed in object as json to the file", func() {
			jsonErr := test_helpers.TempJsonFile(test, func(filepath string) {
				called = true
				file, err := os.Open(filepath)
				Expect(err).NotTo(HaveOccurred())

				n, err := file.Read(buffer)
				Expect(err).NotTo(HaveOccurred())
				buffer = buffer[:n]
				Expect(string(buffer)).To(Equal(`{"number":9}`))
			})

			Expect(called).To(BeTrue())
			Expect(jsonErr).NotTo(HaveOccurred())
		})

		It("removes the file when done", func() {
			var file string
			jsonErr := test_helpers.TempJsonFile(test, func(filepath string) {
				called = true
				file = filepath
			})

			Expect(called).To(BeTrue())
			Expect(jsonErr).NotTo(HaveOccurred())
			_, err := os.Open(file)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})
})
