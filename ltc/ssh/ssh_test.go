package ssh_test

import (
	"errors"
	"io"
	"os"
	"syscall"

	"github.com/docker/docker/pkg/term"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/mocks"
)

type dummyConn struct {
	io.ReadWriteCloser
}

var _ = Describe("SSH", func() {
	var (
		fakeClientDialer   *mocks.FakeClientDialer
		fakeClient         *mocks.FakeClient
		fakeTerm           *mocks.FakeTerm
		fakeListener       *mocks.FakeListener
		fakeSessionFactory *mocks.FakeSessionFactory
		fakeSession        *mocks.FakeSession

		config *config_package.Config
		appSSH *ssh.SSH
	)

	BeforeEach(func() {
		fakeClientDialer = &mocks.FakeClientDialer{}
		fakeClient = &mocks.FakeClient{}
		fakeTerm = &mocks.FakeTerm{}
		fakeListener = &mocks.FakeListener{}
		fakeSessionFactory = &mocks.FakeSessionFactory{}
		fakeSession = &mocks.FakeSession{}

		config = config_package.New(nil)
		appSSH = &ssh.SSH{
			Listener:       fakeListener,
			ClientDialer:   fakeClientDialer,
			Term:           fakeTerm,
			SessionFactory: fakeSessionFactory,
		}
		fakeClientDialer.DialReturns(fakeClient, nil)
		fakeSessionFactory.NewReturns(fakeSession, nil)
	})

	Describe("#ConnectAndForward", func() {
		It("should should forward connection data between the local and remote servers", func() {
			acceptChan := make(chan io.ReadWriteCloser)

			fakeListener.ListenReturns(acceptChan, nil)

			shellChan := make(chan error)
			go func() {
				shellChan <- appSSH.ConnectAndForward("some-app-name", 100, "some local address", "some remote address", config)
			}()

			localConn := &dummyConn{}
			acceptChan <- localConn

			Eventually(fakeClient.ForwardCallCount).Should(Equal(1))
			expectedLocalConn, remoteAddress := fakeClient.ForwardArgsForCall(0)
			Expect(localConn == expectedLocalConn).To(BeTrue())
			Expect(remoteAddress).To(Equal("some remote address"))

			close(acceptChan)

			Expect(<-shellChan).To(Succeed())

			Expect(fakeClientDialer.DialCallCount()).To(Equal(1))
			appName, instanceIndex, configArg := fakeClientDialer.DialArgsForCall(0)
			Expect(appName).To(Equal("some-app-name"))
			Expect(instanceIndex).To(Equal(100))
			Expect(configArg == config).To(BeTrue())

			Expect(fakeListener.ListenCallCount()).To(Equal(1))
			listenNetwork, localAddr := fakeListener.ListenArgsForCall(0)
			Expect(listenNetwork).To(Equal("tcp"))
			Expect(localAddr).To(Equal("some local address"))
		})
	})

	Describe("#ConnectToShell", func() {
		var stopKeepAliveChan chan struct{}

		BeforeEach(func() {
			stopKeepAliveChan = make(chan struct{})
			fakeSession.KeepAliveReturns(stopKeepAliveChan)
		})

		It("should open an interactive terminal to the server with keepalive", func() {
			fakeTerm.GetWinsizeReturns(200, 300)
			termState := &term.State{}
			fakeTerm.SetRawTerminalReturns(termState, nil)

			Expect(appSSH.ConnectToShell("some-app-name", 100, "", config)).To(Succeed())

			Expect(fakeClientDialer.DialCallCount()).To(Equal(1))
			appName, instanceIndex, configArg := fakeClientDialer.DialArgsForCall(0)
			Expect(appName).To(Equal("some-app-name"))
			Expect(instanceIndex).To(Equal(100))
			Expect(configArg == config).To(BeTrue())

			Expect(fakeTerm.GetWinsizeCallCount()).To(Equal(1))
			Expect(fakeTerm.GetWinsizeArgsForCall(0)).To(Equal(os.Stdout.Fd()))

			Expect(fakeSessionFactory.NewCallCount()).To(Equal(1))
			client, width, height := fakeSessionFactory.NewArgsForCall(0)
			Expect(client).To(Equal(fakeClient))
			Expect(width).To(Equal(200))
			Expect(height).To(Equal(300))

			Expect(fakeTerm.SetRawTerminalCallCount()).To(Equal(1))
			Expect(fakeTerm.SetRawTerminalArgsForCall(0)).To(Equal(os.Stdin.Fd()))

			Expect(fakeTerm.RestoreTerminalCallCount()).To(Equal(1))
			fd, state := fakeTerm.RestoreTerminalArgsForCall(0)
			Expect(fd).To(Equal(os.Stdin.Fd()))
			Expect(state == termState).To(BeTrue())

			Expect(fakeSession.KeepAliveCallCount()).To(Equal(1))
			Expect(stopKeepAliveChan).To(BeClosed())

			Expect(fakeSession.ShellCallCount()).To(Equal(1))
			Expect(fakeSession.WaitCallCount()).To(Equal(1))
			Expect(fakeSession.CloseCallCount()).To(Equal(1))
		})

		It("should run a remote command if provided", func() {
			Expect(appSSH.ConnectToShell("some-app-name", 100, "some-command", config)).To(Succeed())

			Expect(fakeSession.ShellCallCount()).To(Equal(0))
			Expect(fakeSession.WaitCallCount()).To(Equal(0))

			Expect(fakeSession.RunCallCount()).To(Equal(1))
			Expect(fakeSession.RunArgsForCall(0)).To(Equal("some-command"))
		})

		It("resizes the remote terminal if the local terminal is resized", func() {
			fakeTerm.GetWinsizeReturns(10, 20)
			waitChan := make(chan struct{})
			shellChan := make(chan error)
			fakeSession.ShellStub = func() error {
				defer GinkgoRecover()
				Expect(fakeSession.ResizeCallCount()).To(Equal(0))
				Expect(fakeTerm.GetWinsizeCallCount()).To(Equal(1))
				fakeTerm.GetWinsizeReturns(30, 40)
				waitChan <- struct{}{}
				waitChan <- struct{}{}
				return nil
			}

			go func() {
				shellChan <- appSSH.ConnectToShell("some-app-name", 100, "", config)
			}()

			<-waitChan

			Expect(syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)).To(Succeed())
			Eventually(fakeTerm.GetWinsizeCallCount, 5).Should(Equal(2))
			Expect(fakeSession.ResizeCallCount()).To(Equal(1))
			width, height := fakeSession.ResizeArgsForCall(0)
			Expect(width).To(Equal(30))
			Expect(height).To(Equal(40))

			<-waitChan

			Expect(<-shellChan).To(Succeed())
		})

		It("does not resize the remote terminal if SIGWINCH is received but the window is the same size", func() {
			fakeTerm.GetWinsizeReturns(10, 20)
			waitChan := make(chan struct{})
			shellChan := make(chan error)
			fakeSession.ShellStub = func() error {
				defer GinkgoRecover()
				Expect(fakeSession.ResizeCallCount()).To(Equal(0))
				Expect(fakeTerm.GetWinsizeCallCount()).To(Equal(1))
				waitChan <- struct{}{}
				waitChan <- struct{}{}
				return nil
			}

			go func() {
				shellChan <- appSSH.ConnectToShell("some-app-name", 100, "", config)
			}()

			<-waitChan

			Expect(syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)).To(Succeed())
			Eventually(fakeTerm.GetWinsizeCallCount, 5).Should(Equal(2))
			Expect(fakeSession.ResizeCallCount()).To(Equal(0))

			<-waitChan

			Expect(<-shellChan).To(Succeed())
		})

		Context("when we fail to dail the SSH server", func() {
			It("returns an error", func() {
				fakeClientDialer.DialReturns(nil, errors.New("some error"))

				err := appSSH.ConnectToShell("", 0, "", config)
				Expect(err).To(MatchError("some error"))
			})
		})

		Context("when we fail to set the raw terminal", func() {
			It("does not restore the terminal at the end", func() {
				fakeTerm.SetRawTerminalReturns(nil, errors.New("some error"))

				Expect(appSSH.ConnectToShell("", 0, "", config)).To(Succeed())
				Expect(fakeTerm.RestoreTerminalCallCount()).To(Equal(0))
			})
		})
	})
})
