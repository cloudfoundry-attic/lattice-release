package config_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/config"
)

var _ = Describe("Config", func() {
	var testConfig *config.Config

	BeforeEach(func() {
		testConfig = config.New(&fakePersister{})
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
			fakePersister := &fakePersister{}
			testConfig = config.New(fakePersister)

			testConfig.SetTarget("mynewapi.com")
			testConfig.SetLogin("testusername", "testpassword")

			testConfig.Save()

			Expect(fakePersister.target).To(Equal("mynewapi.com"))
			Expect(fakePersister.username).To(Equal("testusername"))
			Expect(fakePersister.password).To(Equal("testpassword"))
		})

		It("returns errors from the persistor", func() {
			testConfig = config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.Save()

			Expect(err).To(MatchError("Error"))
		})
	})

	Describe("Load", func() {
		It("loads the target, username, and password from the persister", func() {
			fakePersister := &fakePersister{target: "mysavedapi.com", username: "saveduser", password: "password"}
			testConfig = config.New(fakePersister)

			testConfig.Load()

			Expect(fakePersister.target).To(Equal("mysavedapi.com"))
			Expect(testConfig.Receptor()).To(Equal("http://saveduser:password@receptor.mysavedapi.com"))
		})

		It("returns errors from loading the config", func() {
			testConfig = config.New(&fakePersister{err: errors.New("Error")})

			err := testConfig.Load()

			Expect(err).To(MatchError("Error"))
		})
	})

	Describe("TargetBlob", func() {
		It("sets the blob target", func() {
			testConfig.SetBlobTarget("s3-compatible-store", 7474, "NUYP3C_MBM-WDDWYKIUN", "Nb5vjT2V-ZX0O0s00xURSsg2Se0w-bmX40IQNg4==", "the-bucket")

			blobTarget := testConfig.BlobTarget()
			Expect(blobTarget.TargetHost).To(Equal("s3-compatible-store"))
			Expect(blobTarget.TargetPort).To(Equal(uint16(7474)))
			Expect(blobTarget.AccessKey).To(Equal("NUYP3C_MBM-WDDWYKIUN"))
			Expect(blobTarget.SecretKey).To(Equal("Nb5vjT2V-ZX0O0s00xURSsg2Se0w-bmX40IQNg4=="))
			Expect(blobTarget.BucketName).To(Equal("the-bucket"))
		})
	})

	Describe("BlobTargetInfo", func() {
		var blobTargetInfo config.BlobTargetInfo

		Describe("Proxy", func() {
			It("returns the proxy func", func() {
				blobTargetInfo = config.BlobTargetInfo{
					TargetHost: "success",
					TargetPort: 1818,
				}

				proxyURL, err := blobTargetInfo.Proxy()(&http.Request{})
				Expect(err).NotTo(HaveOccurred())
				Expect(proxyURL.Host).To(Equal("success:1818"))
			})
			Context("when the target host is empty", func() {
				It("returns an func that returns an error", func() {
					blobTargetInfo = config.BlobTargetInfo{
						TargetPort: 1818,
					}

					_, err := blobTargetInfo.Proxy()(&http.Request{})
					Expect(err).To(MatchError("missing proxy host"))
				})
			})
			Context("when the target port is zero", func() {
				It("returns an func that returns an error", func() {
					blobTargetInfo = config.BlobTargetInfo{
						TargetHost: "success",
					}

					_, err := blobTargetInfo.Proxy()(&http.Request{})
					Expect(err).To(MatchError("missing proxy port"))
				})
			})
			Context("when the url is malformed", func() {
				It("returns a func that returns an error", func() {
					blobTargetInfo = config.BlobTargetInfo{
						TargetHost: "succ%2Fess",
						TargetPort: 1818,
					}

					_, err := blobTargetInfo.Proxy()(&http.Request{})
					Expect(err).To(MatchError(ContainSubstring("invalid proxy address")))
				})
			})
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
