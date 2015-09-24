package sshapi_test

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/sshapi"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/sshapi/mocks"
	"golang.org/x/crypto/ssh"
)

var _ = Describe(".New", func() {
	It("should create a client using the package DialFunc", func() {
		origDial := sshapi.DialFunc
		defer func() { sshapi.DialFunc = origDial }()

		dialCalled := false
		sshClient := &ssh.Client{}

		sshapi.DialFunc = func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
			Expect(network).To(Equal("tcp"))
			Expect(addr).To(Equal("some-host"))
			Expect(config.User).To(Equal("some-ssh-user"))

			Expect(config.Auth).To(HaveLen(1))

			actualSecret := reflect.ValueOf(config.Auth[0]).Call([]reflect.Value{})[0].Interface()
			Expect(actualSecret).To(Equal("some-user:some-password"))

			dialCalled = true

			return sshClient, nil
		}

		client, err := sshapi.New("some-ssh-user", "some-user", "some-password", "some-host")
		Expect(err).NotTo(HaveOccurred())
		Expect(client.Dialer == sshClient).To(BeTrue())
		Expect(client.SSHSessionFactory.(*sshapi.CryptoSSHSessionFactory).Client == sshClient).To(BeTrue())
		Expect(client.Stdin).To(Equal(os.Stdin))
		Expect(client.Stdout).To(Equal(os.Stdout))
		Expect(client.Stderr).To(Equal(os.Stderr))

		Expect(dialCalled).To(BeTrue())
	})

	Context("when dialing fails", func() {
		It("should return an error", func() {
			origDial := sshapi.DialFunc
			defer func() { sshapi.DialFunc = origDial }()

			sshapi.DialFunc = func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
				return &ssh.Client{}, errors.New("some error")
			}

			_, err := sshapi.New("some-ssh-user", "some-user", "some-password", "some-host")
			Expect(err).To(MatchError("some error"))
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

var _ = Describe("Client", func() {
	Describe("#Forward", func() {
		var (
			client     *sshapi.Client
			fakeDialer *mocks.FakeDialer
		)

		BeforeEach(func() {
			fakeDialer = &mocks.FakeDialer{}
			client = &sshapi.Client{Dialer: fakeDialer}
		})

		It("should dial a remote connection", func() {
			localConn := &mockConn{Reader: &bytes.Buffer{}, Writer: &bytes.Buffer{}}
			remoteConn := &mockConn{Reader: &bytes.Buffer{}, Writer: &bytes.Buffer{}}
			fakeDialer.DialReturns(remoteConn, nil)

			Expect(client.Forward(localConn, "some remote address")).To(Succeed())

			Expect(fakeDialer.DialCallCount()).To(Equal(1))
			protocol, address := fakeDialer.DialArgsForCall(0)
			Expect(protocol).To(Equal("tcp"))
			Expect(address).To(Equal("some remote address"))
		})

		It("should copy data in both directions", func() {
			localConnBuffer := &bytes.Buffer{}
			remoteConnBuffer := &bytes.Buffer{}
			localConn := &mockConn{Reader: bytes.NewBufferString("some local data"), Writer: localConnBuffer}
			remoteConn := &mockConn{Reader: bytes.NewBufferString("some remote data"), Writer: remoteConnBuffer}
			fakeDialer.DialReturns(remoteConn, nil)

			Expect(client.Forward(localConn, "some remote address")).To(Succeed())

			Expect(localConn.closed).To(BeTrue())
			Expect(remoteConn.closed).To(BeTrue())
			Expect(localConnBuffer.String()).To(Equal("some remote data"))
			Expect(remoteConnBuffer.String()).To(Equal("some local data"))
		})

		Context("when dialing a remote connection fails", func() {
			It("should return an error", func() {
				fakeDialer.DialReturns(nil, errors.New("some error"))
				err := client.Forward(nil, "some remote address")
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("#Open", func() {
		var (
			client             *sshapi.Client
			mockSession        *mocks.FakeSSHSession
			mockSessionFactory *mocks.FakeSSHSessionFactory
			originalTerm       string
		)

		BeforeEach(func() {
			originalTerm = os.Getenv("TERM")
			mockSessionFactory = &mocks.FakeSSHSessionFactory{}
			client = &sshapi.Client{
				SSHSessionFactory: mockSessionFactory,
			}
			mockSession = &mocks.FakeSSHSession{}
			mockSessionFactory.NewReturns(mockSession, nil)
		})

		AfterEach(func() {
			os.Setenv("TERM", originalTerm)
		})

		It("should open a new session", func() {
			os.Setenv("TERM", "some term")

			client.Stdin = bytes.NewBufferString("some client in data")
			client.Stdout = &bytes.Buffer{}
			client.Stderr = &bytes.Buffer{}
			mockSessionStdinBuffer := &bytes.Buffer{}
			mockSessionStdin := &mockConn{Writer: mockSessionStdinBuffer}
			mockSessionStdout := bytes.NewBufferString("some session out data")
			mockSessionStderr := bytes.NewBufferString("some session err data")
			mockSession.StdinPipeReturns(mockSessionStdin, nil)
			mockSession.StdoutPipeReturns(mockSessionStdout, nil)
			mockSession.StderrPipeReturns(mockSessionStderr, nil)

			_, err := client.Open(100, 200, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockSession.RequestPtyCallCount()).To(Equal(1))
			termType, height, width, modes := mockSession.RequestPtyArgsForCall(0)
			Expect(termType).To(Equal("some term"))
			Expect(height).To(Equal(200))
			Expect(width).To(Equal(100))
			Expect(modes[ssh.ECHO]).To(Equal(uint32(1)))
			Expect(modes[ssh.TTY_OP_ISPEED]).To(Equal(uint32(115200)))
			Expect(modes[ssh.TTY_OP_OSPEED]).To(Equal(uint32(115200)))

			Eventually(mockSessionStdinBuffer.String).Should(Equal("some client in data"))
			Eventually(client.Stdout.(*bytes.Buffer).String).Should(Equal("some session out data"))
			Eventually(client.Stderr.(*bytes.Buffer).String).Should(Equal("some session err data"))
			Eventually(func() bool { return mockSessionStdin.closed }).Should(BeTrue())
		})

		It("should not request a pty when desirePTY is false", func() {
			client.Stdin = bytes.NewBufferString("some client in data")
			client.Stdout = &bytes.Buffer{}
			client.Stderr = &bytes.Buffer{}
			mockSessionStdinBuffer := &bytes.Buffer{}
			mockSessionStdin := &mockConn{Writer: mockSessionStdinBuffer}
			mockSessionStdout := bytes.NewBufferString("some session out data")
			mockSessionStderr := bytes.NewBufferString("some session err data")
			mockSession.StdinPipeReturns(mockSessionStdin, nil)
			mockSession.StdoutPipeReturns(mockSessionStdout, nil)
			mockSession.StderrPipeReturns(mockSessionStderr, nil)

			_, err := client.Open(100, 200, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockSession.RequestPtyCallCount()).To(Equal(0))
		})

		It("should request a pty when desirePTY is true", func() {
			client.Stdin = bytes.NewBufferString("some client in data")
			client.Stdout = &bytes.Buffer{}
			client.Stderr = &bytes.Buffer{}
			mockSessionStdinBuffer := &bytes.Buffer{}
			mockSessionStdin := &mockConn{Writer: mockSessionStdinBuffer}
			mockSessionStdout := bytes.NewBufferString("some session out data")
			mockSessionStderr := bytes.NewBufferString("some session err data")
			mockSession.StdinPipeReturns(mockSessionStdin, nil)
			mockSession.StdoutPipeReturns(mockSessionStdout, nil)
			mockSession.StderrPipeReturns(mockSessionStderr, nil)

			_, err := client.Open(100, 200, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockSession.RequestPtyCallCount()).To(Equal(1))
		})

		Context("when we fail to open a new session", func() {
			It("should return an error", func() {
				mockSessionFactory.NewReturns(nil, errors.New("some error"))
				_, err := client.Open(100, 200, true)
				Expect(err).To(MatchError("some error"))
			})
		})

		Context("when we fail to open any of the session pipes", func() {
			It("should return an error", func() {
				mockSession.StderrPipeReturns(nil, errors.New("some stderr error"))
				_, err := client.Open(100, 200, true)
				Expect(err).To(MatchError("some stderr error"))

				mockSession.StdoutPipeReturns(nil, errors.New("some stdout error"))
				_, err = client.Open(100, 200, true)
				Expect(err).To(MatchError("some stdout error"))

				mockSession.StdinPipeReturns(nil, errors.New("some stdin error"))
				_, err = client.Open(100, 200, true)
				Expect(err).To(MatchError("some stdin error"))
			})
		})

		Context("when requesting a PTY fails", func() {
			It("should return an error", func() {
				mockSession.RequestPtyReturns(errors.New("some error"))
				_, err := client.Open(100, 200, true)
				Expect(err).To(MatchError("some error"))

			})
		})
	})
})
