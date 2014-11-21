package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
)

var _ = Describe("config", func() {
	Describe("Load", func() {
		It("returns errors from loading the config", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.Load()

			Expect(err).To(Equal(errors.New("Error")))
		})
	})

	Describe("Api", func() {
		It("Loads the API from the persistor", func() {
			testConfig := config.New(&fakePersister{api: "receptor.mytestapi.com"})

			testConfig.Load()

			Expect(testConfig.Api()).To(Equal("receptor.mytestapi.com"))
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

	Describe("Loggregator", func() {
		It("Loads the Loggregator from the persistor", func() {
			testConfig := config.New(&fakePersister{loggregator: "doppler.mytestapi.com"})

			testConfig.Load()

			Expect(testConfig.Loggregator()).To(Equal("doppler.mytestapi.com"))
		})

	})

	Describe("SetLoggregator", func() {
		It("saves loggregator to the persistor", func() {
			fakePersister := &fakePersister{}
			testConfig := config.New(fakePersister)

			testConfig.SetLoggregator("doppler.mynewapi.com")
			Expect(testConfig.Loggregator()).To(Equal("doppler.mynewapi.com"))

			Expect(fakePersister.loggregator).To(Equal("doppler.mynewapi.com"))

		})

		It("returns errors from the persistor", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.SetLoggregator("receptor.mynewapi.com")

			Expect(err).To(Equal(errors.New("Error")))
		})
	})
})

type fakePersister struct {
	api         string
	loggregator string
	err         error
}

func (f *fakePersister) Load(dataInterface interface{}) error {
	data, ok := dataInterface.(*config.Data)
	Expect(ok).To(BeTrue())

	data.Api = f.api
	data.Loggregator = f.loggregator
	return f.err
}

func (f *fakePersister) Save(dataInterface interface{}) error {
	data, ok := dataInterface.(*config.Data)
	Expect(ok).To(BeTrue())

	f.api = data.Api
	f.loggregator = data.Loggregator
	return f.err
}
