package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
)

type fakePersister struct {
	api string
	err error
}

func (f *fakePersister) Load(dataInterface interface{}) error {
	data, ok := dataInterface.(*config.Data)
	Expect(ok).To(BeTrue())

	data.Api = f.api
	return f.err
}

func (f *fakePersister) Save(dataInterface interface{}) error {
	data, ok := dataInterface.(*config.Data)
	Expect(ok).To(BeTrue())

	f.api = data.Api
	return f.err
}

var _ = Describe("config", func() {
	Describe("Api", func() {
		It("Loads the API from the persistor", func() {
			testConfig := config.New(&fakePersister{api: "receptor.mytestapi.com"})

			testConfig.Load()

			Expect(testConfig.Api()).To(Equal("receptor.mytestapi.com"))
		})

		It("returns errors from loading the config", func() {
			testConfig := config.New(&fakePersister{api: "receptor.mytestapi.com", err: errors.New("Error")})

			err := testConfig.Load()

			Expect(err).To(Equal(errors.New("Error")))
		})
	})

	Describe("SetApi", func() {
		It("saves api to the persistor", func() {
			fakePersister := &fakePersister{}
			testConfig := config.New(fakePersister)

			testConfig.SetApi("receptor.mynewapi.com")
			Expect(testConfig.Api()).To(Equal("receptor.mynewapi.com"))

			Expect(fakePersister.api).To(Equal("receptor.mynewapi.com"))

		})

		It("returns errors from the persistor", func() {
			testConfig := config.New(&fakePersister{api: "receptor.mytestapi.com", err: errors.New("Error")})

			err := testConfig.SetApi("receptor.mynewapi.com")

			Expect(err).To(Equal(errors.New("Error")))
		})
	})
})
