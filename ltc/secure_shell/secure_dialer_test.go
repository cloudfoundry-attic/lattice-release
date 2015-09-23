package secure_shell_test

import (
	"bytes"
	"errors"
	"io"
	"net"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_ssh_client"
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

type mockConn struct {
	io.Reader
	io.Writer
	nilNetConn
	closed bool
}

type nilNetConn struct {
	net.Conn
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

var _ = Describe("SecureClient", func() {
	var (
		secureClient  *secure_shell.SecureClient
		mockSSHClient *fake_ssh_client.FakeSSHClient
	)

	BeforeEach(func() {
		mockSSHClient = &fake_ssh_client.FakeSSHClient{}
		secureClient = &secure_shell.SecureClient{mockSSHClient}
	})

	Describe("#Accept", func() {
		It("should dial a remote connection", func() {
			localConn := &mockConn{Reader: &bytes.Buffer{}, Writer: &bytes.Buffer{}}
			remoteConn := &mockConn{Reader: &bytes.Buffer{}, Writer: &bytes.Buffer{}}
			mockSSHClient.DialReturns(remoteConn, nil)

			Expect(secureClient.Accept(localConn, "some remote address")).To(Succeed())

			Expect(mockSSHClient.DialCallCount()).To(Equal(1))
			protocol, address := mockSSHClient.DialArgsForCall(0)
			Expect(protocol).To(Equal("tcp"))
			Expect(address).To(Equal("some remote address"))
		})

		It("should copy data in both directions", func() {
			localConnBuffer := &bytes.Buffer{}
			remoteConnBuffer := &bytes.Buffer{}
			localConn := &mockConn{Reader: bytes.NewBufferString("some local data"), Writer: localConnBuffer}
			remoteConn := &mockConn{Reader: bytes.NewBufferString("some remote data"), Writer: remoteConnBuffer}
			mockSSHClient.DialReturns(remoteConn, nil)

			Expect(secureClient.Accept(localConn, "some remote address")).To(Succeed())

			Expect(localConn.closed).To(BeTrue())
			Expect(remoteConn.closed).To(BeTrue())
			Expect(localConnBuffer.String()).To(Equal("some remote data"))
			Expect(remoteConnBuffer.String()).To(Equal("some local data"))
		})

		Context("when dialing a remote connection fails", func() {
			It("should return an error", func() {
				mockSSHClient.DialReturns(nil, errors.New("some error"))
				err := secureClient.Accept(nil, "some remote address")
				Expect(err).To(MatchError("some error"))
			})
		})
	})
})
