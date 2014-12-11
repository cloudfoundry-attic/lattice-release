package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/lattice-cli/config"
)

var _ = Describe("config", func() {
	Describe("Load", func() {
		It("returns errors from loading the config", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.Load()

			Expect(err).To(Equal(errors.New("Error")))
		})
	})

	Describe("SetTarget", func() {
		It("saves api to the persistor", func() {
			fakePersister := &fakePersister{}
			testConfig := config.New(fakePersister)

			testConfig.SetTarget("mynewapi.com")

			Expect(fakePersister.target).To(Equal("mynewapi.com"))
		})

		It("returns errors from the persistor", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.SetTarget("mynewapi.com")

			Expect(err).To(Equal(errors.New("Error")))
		})
	})

	Describe("SetLogin", func() {
		It("saves api to the persistor", func() {
			fakePersister := &fakePersister{}
			testConfig := config.New(fakePersister)

			testConfig.SetLogin("testusername", "testpassword")

			Expect(fakePersister.username).To(Equal("testusername"))
			Expect(fakePersister.password).To(Equal("testpassword"))

		})

		It("returns errors from the persistor", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.SetLogin("testusername", "testpassword")

			Expect(err).To(Equal(errors.New("Error")))
		})
	})

	Describe("Receptor", func() {
		It("Loads the Receptor from the persistor with a username and password", func() {
			testConfig := config.New(&fakePersister{target: "mytestapi.com", username: "testusername", password: "testpassword"})

			testConfig.Load()

			Expect(testConfig.Receptor()).To(Equal("http://testusername:testpassword@receptor.mytestapi.com"))
		})

		It("Loads the Receptor from the persistor without a username and password", func() {
			testConfig := config.New(&fakePersister{target: "mytestapi.com"})

			testConfig.Load()

			Expect(testConfig.Receptor()).To(Equal("http://receptor.mytestapi.com"))
		})
	})

	Describe("Loggregator", func() {
		It("Loads the Loggregator from the persistor", func() {
			testConfig := config.New(&fakePersister{target: "mytestapi.com"})

			testConfig.Load()

			Expect(testConfig.Loggregator()).To(Equal("doppler.mytestapi.com"))
		})
	})

	Describe("Target", func() {
		It("Loads the target from the persistor", func() {
			testConfig := config.New(&fakePersister{target: "mytestapi.com"})

			testConfig.Load()

			Expect(testConfig.Target()).To(Equal("mytestapi.com"))
		})

	})
})

type fakePersister struct {
	target   string
	username string
	password string
	err      error
}

func (f *fakePersister) Load(dataInterface interface{}) error {
	data, ok := dataInterface.(*config.Data)
	Expect(ok).To(BeTrue())

	data.Target = f.target
	data.Username = f.username
	data.Password = f.password
	return f.err
}

func (f *fakePersister) Save(dataInterface interface{}) error {
	data, ok := dataInterface.(*config.Data)
	Expect(ok).To(BeTrue())

	f.target = data.Target
	f.username = data.Username
	f.password = data.Password
	return f.err
}
