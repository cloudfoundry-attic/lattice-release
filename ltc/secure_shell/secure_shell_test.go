package secure_shell_test

import (
	"errors"
	"os"
	"syscall"

	"github.com/docker/docker/pkg/term"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_dialer"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_secure_session"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_term"
)

var _ = Describe("SecureShell", func() {
	var (
		fakeDialer  *fake_dialer.FakeDialer
		fakeSession *fake_secure_session.FakeSecureSession
		fakeTerm    *fake_term.FakeTerm
		fakeStdin   *gbytes.Buffer
		fakeStdout  *gbytes.Buffer
		fakeStderr  *gbytes.Buffer

		config      *config_package.Config
		secureShell *secure_shell.SecureShell

		oldTerm string
	)

	BeforeEach(func() {
		fakeDialer = &fake_dialer.FakeDialer{}
		fakeSession = &fake_secure_session.FakeSecureSession{}
		fakeTerm = &fake_term.FakeTerm{}
		fakeStdin = gbytes.NewBuffer()
		fakeStdout = gbytes.NewBuffer()
		fakeStderr = gbytes.NewBuffer()

		config = config_package.New(nil)
		config.SetTarget("10.0.12.34")
		config.SetLogin("user", "past")

		secureShell = &secure_shell.SecureShell{Dialer: fakeDialer, Term: fakeTerm}
		fakeDialer.DialReturns(fakeSession, nil)

		oldTerm = os.Getenv("TERM")
		os.Setenv("TERM", "defaultterm")
	})

	AfterEach(func() {
		os.Setenv("TERM", oldTerm)
	})

	Describe("#ConnectToShell", func() {
		It("connects to the correct server given app name, instance and config", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)
			fakeTerm.GetWinsizeReturns(1000, 2000)

			termState := &term.State{}
			fakeTerm.SetRawTerminalReturns(termState, nil)

			err := secureShell.ConnectToShell("app-name", 2, "", config)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDialer.DialCallCount()).To(Equal(1))
			user, authUser, authPass, address := fakeDialer.DialArgsForCall(0)
			Expect(user).To(Equal("diego:app-name/2"))
			Expect(authUser).To(Equal("user"))
			Expect(authPass).To(Equal("past"))
			Expect(address).To(Equal("10.0.12.34:2222"))

			Expect(fakeTerm.SetRawTerminalCallCount()).To(Equal(1))
			Expect(fakeTerm.SetRawTerminalArgsForCall(0)).To(Equal(os.Stdin.Fd()))

			Expect(fakeTerm.GetWinsizeCallCount()).To(Equal(1))
			Expect(fakeTerm.GetWinsizeArgsForCall(0)).To(Equal(os.Stdout.Fd()))

			Expect(fakeSession.RequestPtyCallCount()).To(Equal(1))
			termType, height, width, _ := fakeSession.RequestPtyArgsForCall(0)
			Expect(termType).To(Equal("defaultterm"))
			Expect(width).To(Equal(1000))
			Expect(height).To(Equal(2000))

			Expect(fakeTerm.RestoreTerminalCallCount()).To(Equal(1))
			fd, state := fakeTerm.RestoreTerminalArgsForCall(0)
			Expect(fd).To(Equal(os.Stdin.Fd()))
			Expect(state).To(Equal(termState))

			Expect(fakeSession.ShellCallCount()).To(Equal(1))
			Expect(fakeSession.WaitCallCount()).To(Equal(1))
		})

		It("runs a remote command", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)

			err := secureShell.ConnectToShell("app-name", 2, "/bin/ls", config)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSession.ShellCallCount()).To(Equal(0))
			Expect(fakeSession.WaitCallCount()).To(Equal(0))

			Expect(fakeSession.RunCallCount()).To(Equal(1))
			Expect(fakeSession.RunArgsForCall(0)).To(Equal("/bin/ls"))
		})

		It("respects the user's TERM environment variable", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)

			os.Setenv("TERM", "term2000")

			err := secureShell.ConnectToShell("app-name", 2, "", config)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSession.RequestPtyCallCount()).To(Equal(1))
			termType, _, _, _ := fakeSession.RequestPtyArgsForCall(0)
			Expect(termType).To(Equal("term2000"))
		})

		It("defaults to xterm ifno TERM environment variable is set", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)

			os.Setenv("TERM", "")

			err := secureShell.ConnectToShell("app-name", 2, "", config)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSession.RequestPtyCallCount()).To(Equal(1))
			termType, _, _, _ := fakeSession.RequestPtyArgsForCall(0)
			Expect(termType).To(Equal("xterm"))
		})

		It("resizes the remote terminal if the local terminal is resized", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)

			fakeTerm.GetWinsizeReturns(10, 20)

			waitChan := make(chan struct{})
			fakeSession.ShellStub = func() error {
				Expect(fakeSession.SendRequestCallCount()).To(Equal(0))
				Expect(fakeTerm.GetWinsizeCallCount()).To(Equal(1))
				fakeTerm.GetWinsizeReturns(30, 40)
				Eventually(waitChan, 3).Should(Receive())
				return nil
			}

			go func() {
				defer GinkgoRecover()

				Eventually(waitChan).Should(Receive())
				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).NotTo(HaveOccurred())
			}()

			waitChan <- struct{}{}

			err := syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
			Expect(err).NotTo(HaveOccurred())

			Eventually(fakeTerm.GetWinsizeCallCount, 3).Should(Equal(2))
			Expect(fakeSession.SendRequestCallCount()).To(Equal(1))
			name, wantReply, payload := fakeSession.SendRequestArgsForCall(0)
			Expect(name).To(Equal("window-change"))
			Expect(wantReply).To(BeFalse())
			Expect(payload).To(Equal([]byte{0, 0, 0, 30, 0, 0, 0, 40, 0, 0, 0, 0, 0, 0, 0, 0}))

			waitChan <- struct{}{}
		})

		It("does not resize the remote terminal if SIGWINCH is received but the window is the same size", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)

			fakeTerm.GetWinsizeReturns(10, 20)

			waitChan := make(chan struct{})
			fakeSession.ShellStub = func() error {
				Expect(fakeSession.SendRequestCallCount()).To(Equal(0))
				Expect(fakeTerm.GetWinsizeCallCount()).To(Equal(1))
				Eventually(waitChan, 3).Should(Receive())
				return nil
			}

			go func() {
				defer GinkgoRecover()

				Eventually(waitChan).Should(Receive())
				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).NotTo(HaveOccurred())
			}()

			waitChan <- struct{}{}

			Expect(syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)).To(Succeed())

			Eventually(fakeTerm.GetWinsizeCallCount).Should(Equal(2))
			Expect(fakeSession.SendRequestCallCount()).To(Equal(0))

			waitChan <- struct{}{}
		})

		Context("when the SecureDialer#Dial fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(nil, errors.New("cannot dial error"))

				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).To(MatchError("cannot dial error"))
			})
		})

		Context("when the SecureSession#StdinPipe fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(nil, errors.New("put this in your pipe"))

				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).To(MatchError("put this in your pipe"))
			})
		})

		Context("when the SecureSession#StdoutPipe fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(fakeStdin, nil)
				fakeSession.StdoutPipeReturns(nil, errors.New("put this in your pipe"))

				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).To(MatchError("put this in your pipe"))
			})
		})

		Context("when the SecureSession#StderrPipe fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(fakeStdin, nil)
				fakeSession.StdoutPipeReturns(fakeStdout, nil)
				fakeSession.StderrPipeReturns(nil, errors.New("put this in your pipe"))

				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).To(MatchError("put this in your pipe"))
			})
		})

		Context("when the SecureSession#RequestPty fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(fakeStdin, nil)
				fakeSession.StdoutPipeReturns(fakeStdout, nil)
				fakeSession.StderrPipeReturns(fakeStderr, nil)
				fakeSession.RequestPtyReturns(errors.New("no pty"))

				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).To(MatchError("no pty"))
			})
		})

		Context("when the SecureTerm#SetRawTerminal fails", func() {
			It("does not call RestoreTerminal", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(fakeStdin, nil)
				fakeSession.StdoutPipeReturns(fakeStdout, nil)
				fakeSession.StderrPipeReturns(fakeStderr, nil)
				fakeTerm.SetRawTerminalReturns(nil, errors.New("can't set raw"))

				err := secureShell.ConnectToShell("app-name", 2, "", config)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTerm.RestoreTerminalCallCount()).To(Equal(0))
			})
		})
	})
})
