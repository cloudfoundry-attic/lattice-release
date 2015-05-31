package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("config", func() {
	Describe("Target", func() {
		It("sets the target", func() {
			testConfig := config.New(&fakePersister{})
			testConfig.SetTarget("mynewapi.com")

			Expect(testConfig.Target()).To(Equal("mynewapi.com"))
		})
	})

	Describe("Username", func() {
		It("sets the target", func() {
			testConfig := config.New(&fakePersister{})
			testConfig.SetLogin("ausername", "apassword")

			Expect(testConfig.Username()).To(Equal("ausername"))
		})
	})

	Describe("Receptor", func() {
		It("returns the Receptor with a username and password", func() {
			testConfig := config.New(&fakePersister{})
			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("testusername", "testpassword")

			Expect(testConfig.Receptor()).To(Equal("http://testusername:testpassword@receptor.mynewapi.com"))
		})

		It("returns a Receptor without a username and password", func() {
			testConfig := config.New(&fakePersister{})
			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("", "")

			Expect(testConfig.Receptor()).To(Equal("http://receptor.mynewapi.com"))
		})
	})

	Describe("Loggregator", func() {
		It("provides the loggregator doppler path", func() {
			testConfig := config.New(&fakePersister{})
			testConfig.SetTarget("mytestapi.com")

			Expect(testConfig.Loggregator()).To(Equal("doppler.mytestapi.com"))
		})
	})

	Describe("Save", func() {
		It("Saves the target with the persistor", func() {
			fakePersister := &fakePersister{}
			testConfig := config.New(fakePersister)

			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("testusername", "testpassword")

			testConfig.Save()

			Expect(fakePersister.target).To(Equal("mynewapi.com"))
			Expect(fakePersister.username).To(Equal("testusername"))
			Expect(fakePersister.password).To(Equal("testpassword"))
		})

		It("returns errors from the persistor", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.Save()

			Expect(err).To(MatchError("Error"))
		})
	})

	Describe("Load", func() {
		It("loads the target, username, and password from the persister", func() {
			fakePersister := &fakePersister{target: "mysavedapi.com", username: "saveduser", password: "password"}
			testConfig := config.New(fakePersister)

			testConfig.Load()

			Expect(fakePersister.target).To(Equal("mysavedapi.com"))
			Expect(testConfig.Receptor()).To(Equal("http://saveduser:password@receptor.mysavedapi.com"))
		})

		It("returns errors from loading the config", func() {
			testConfig := config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.Load()

			Expect(err).To(MatchError("Error"))
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
