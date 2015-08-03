package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
)

var _ = Describe("Config", func() {
	var (
		testPersister *fakePersister
		testConfig    *config.Config
	)

	BeforeEach(func() {
		testPersister = &fakePersister{}
		testConfig = config.New(testPersister)
	})

	Describe("Target", func() {
		It("sets the target", func() {
			testConfig.SetTarget("mynewapi.com")

			Expect(testConfig.Target()).To(Equal("mynewapi.com"))
		})
	})

	Describe("Username", func() {
		It("sets the target", func() {
			testConfig.SetLogin("ausername", "apassword")

			Expect(testConfig.Username()).To(Equal("ausername"))
		})
	})

	Describe("Receptor", func() {
		It("returns the Receptor with a username and password", func() {
			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("testusername", "testpassword")

			Expect(testConfig.Receptor()).To(Equal("http://testusername:testpassword@receptor.mynewapi.com"))
		})

		It("returns a Receptor without a username and password", func() {
			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("", "")

			Expect(testConfig.Receptor()).To(Equal("http://receptor.mynewapi.com"))
		})
	})

	Describe("Loggregator", func() {
		It("provides the loggregator doppler path", func() {
			testConfig.SetTarget("mytestapi.com")

			Expect(testConfig.Loggregator()).To(Equal("doppler.mytestapi.com"))
		})
	})

	Describe("Save", func() {
		It("Saves the target with the persistor", func() {
			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("testusername", "testpassword")
			Expect(testConfig.Save()).To(Succeed())

			Expect(testPersister.target).To(Equal("mynewapi.com"))
			Expect(testPersister.username).To(Equal("testusername"))
			Expect(testPersister.password).To(Equal("testpassword"))
		})

		It("returns errors from the persistor", func() {
			testPersister.err = errors.New("Error")

			err := testConfig.Save()
			Expect(err).To(MatchError("Error"))
		})
	})

	Describe("Load", func() {
		It("loads the target, username, and password from the persister", func() {
			testPersister.target = "mysavedapi.com"
			testPersister.username = "saveduser"
			testPersister.password = "password"

			Expect(testConfig.Load()).To(Succeed())

			Expect(testPersister.target).To(Equal("mysavedapi.com"))
			Expect(testConfig.Receptor()).To(Equal("http://saveduser:password@receptor.mysavedapi.com"))
		})

		It("returns errors from loading the config", func() {
			testPersister.err = errors.New("Error")

			err := testConfig.Load()
			Expect(err).To(MatchError("Error"))
		})
	})

	Describe("TargetBlob", func() {
		It("sets the blob target", func() {
			testConfig.SetBlobStore("some-host", "7474", "some-username", "some-password")

			Expect(testConfig.BlobStore()).To(Equal(dav_blob_store.Config{
				Host:     "some-host",
				Port:     "7474",
				Username: "some-username",
				Password: "some-password",
			}))
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
