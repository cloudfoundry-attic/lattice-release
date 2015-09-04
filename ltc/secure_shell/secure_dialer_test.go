package secure_shell_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("SecureDialer", func() {
	var secureDialer *secure_shell.SecureDialer

	BeforeEach(func() {
		secureDialer = &secure_shell.SecureDialer{}
	})

	Describe("#Dial", func() {
		It("passes correct args to the Dial impl", func() {
			dialCalled := false
			secureDialer.DialFunc = func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
				Expect(network).To(Equal("tcp"))
				Expect(addr).To(Equal("10.0.12.34:2222"))
				Expect(config.User).To(Equal("diego:app-name/2"))

				Expect(config.Auth).To(HaveLen(1))

				// (╯°□°）╯︵ ┻━┻
				actualSecret := reflect.ValueOf(config.Auth[0]).Call([]reflect.Value{})[0].Interface()
				Expect(actualSecret).To(Equal("user:past"))

				dialCalled = true

				return nil, errors.New("")
			}

			secureDialer.Dial("diego:app-name/2", "user", "past", "10.0.12.34:2222")

			Expect(dialCalled).To(BeTrue())
		})
	})
})
