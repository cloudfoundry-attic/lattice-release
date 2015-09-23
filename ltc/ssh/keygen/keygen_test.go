package keygen_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	keygen_package "github.com/cloudfoundry-incubator/lattice/ltc/ssh/keygen"
)

var _ = Describe("KeyGenerator", func() {
	var fakeReader io.Reader

	BeforeEach(func() {
		fakeReader = bytes.NewReader([]byte{0xc3, 0x4b, 0x57, 0xf5, 0xf1, 0xb8, 0x10, 0x5e, 0x05, 0x2e, 0xbd, 0x86, 0xc8, 0x66, 0x72, 0xaf})
	})

	Describe("#GenerateRSAPRivateKey", func() {
		It("generates a pem-encoded RSA key", func() {
			keygen := &keygen_package.KeyGenerator{fakeReader}
			private, err := keygen.GenerateRSAPrivateKey(40)
			Expect(err).NotTo(HaveOccurred())
			Expect(private).To(Equal("-----BEGIN RSA PRIVATE KEY-----\nMDACAQACBgCsdAECKwIDAQABAgUOJiA6aQIDDfG9AgMMXgcCAwvHrQIDAr8jAgML\nej0=\n-----END RSA PRIVATE KEY-----\n"))
		})
	})

	Describe("#GenerateRSAKeyPair", func() {
		It("generates a pem-encoded RSA key and an ssh authorized key", func() {
			keygen := &keygen_package.KeyGenerator{fakeReader}
			private, public, err := keygen.GenerateRSAKeyPair(40)
			Expect(err).NotTo(HaveOccurred())
			Expect(public).To(Equal("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAABgCsdAECKw=="))
			Expect(private).To(Equal("-----BEGIN RSA PRIVATE KEY-----\nMDACAQACBgCsdAECKwIDAQABAgUOJiA6aQIDDfG9AgMMXgcCAwvHrQIDAr8jAgML\nej0=\n-----END RSA PRIVATE KEY-----\n"))
		})
	})
	// TODO: test errors
})
