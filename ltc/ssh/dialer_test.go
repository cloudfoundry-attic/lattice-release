package ssh_test

import (
	"errors"
	"reflect"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	sshp "github.com/cloudfoundry-incubator/lattice/ltc/ssh"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/sshapi"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppDialer", func() {
	Describe("#Dial", func() {
		var (
			origDial func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
			config   *config_package.Config
		)

		BeforeEach(func() {
			config = config_package.New(nil)
			config.SetTarget("some-host")
			config.SetLogin("some-user", "some-password")
			origDial = sshapi.DialFunc
		})

		AfterEach(func() {
			sshapi.DialFunc = origDial
		})

		It("should create a client", func() {
			dialCalled := false
			sshClient := &ssh.Client{}

			sshapi.DialFunc = func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
				Expect(network).To(Equal("tcp"))
				Expect(addr).To(Equal("some-host:2222"))
				Expect(config.User).To(Equal("diego:some-app-name/100"))

				Expect(config.Auth).To(HaveLen(1))

				actualSecret := reflect.ValueOf(config.Auth[0]).Call([]reflect.Value{})[0].Interface()
				Expect(actualSecret).To(Equal("some-user:some-password"))

				dialCalled = true

				return sshClient, nil
			}

			client, err := (&sshp.AppDialer{}).Dial("some-app-name", 100, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(client.(*sshapi.Client).Dialer == sshClient).To(BeTrue())

			Expect(dialCalled).To(BeTrue())
		})

		Context("when dialing fails", func() {
			It("should return an error", func() {
				origDial := sshapi.DialFunc
				defer func() { sshapi.DialFunc = origDial }()

				sshapi.DialFunc = func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
					return &ssh.Client{}, errors.New("some error")
				}

				_, err := (&sshp.AppDialer{}).Dial("some-app-name", 100, config)
				Expect(err).To(MatchError("some error"))
			})
		})
	})
})
