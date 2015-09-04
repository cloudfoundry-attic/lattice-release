package secure_shell_test

import (
	"errors"
	"os"

	"github.com/docker/docker/pkg/term"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_dialer"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_secure_session"
	"github.com/cloudfoundry-incubator/lattice/ltc/secure_shell/fake_secure_term"
)

var _ = Describe("SecureShell", func() {
	var (
		fakeDialer  *fake_dialer.FakeDialer
		fakeSession *fake_secure_session.FakeSecureSession
		fakeTerm    *fake_secure_term.FakeSecureTerm
		fakeStdin   *gbytes.Buffer
		fakeStdout  *gbytes.Buffer
		fakeStderr  *gbytes.Buffer

		config      *config_package.Config
		secureShell *secure_shell.SecureShell
	)

	BeforeEach(func() {
		fakeDialer = &fake_dialer.FakeDialer{}
		fakeSession = &fake_secure_session.FakeSecureSession{}
		fakeTerm = &fake_secure_term.FakeSecureTerm{}
		fakeStdin = gbytes.NewBuffer()
		fakeStdout = gbytes.NewBuffer()
		fakeStderr = gbytes.NewBuffer()

		config = config_package.New(nil)
		config.SetTarget("10.0.12.34")
		config.SetLogin("user", "past")

		secureShell = &secure_shell.SecureShell{Dialer: fakeDialer, Term: fakeTerm}
		fakeDialer.DialReturns(fakeSession, nil)
	})

	Describe("#ConnectToShell", func() {
		It("connects to the correct server given app name, instance and config", func() {
			fakeDialer.DialReturns(fakeSession, nil)
			fakeSession.StdinPipeReturns(fakeStdin, nil)
			fakeSession.StdoutPipeReturns(fakeStdout, nil)
			fakeSession.StderrPipeReturns(fakeStderr, nil)

			termState := &term.State{}
			fakeTerm.SetRawTerminalReturns(termState, nil)

			err := secureShell.ConnectToShell("app-name", 2, config)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDialer.DialCallCount()).To(Equal(1))
			user, authUser, authPass, address := fakeDialer.DialArgsForCall(0)
			Expect(user).To(Equal("diego:app-name/2"))
			Expect(authUser).To(Equal("user"))
			Expect(authPass).To(Equal("past"))
			Expect(address).To(Equal("10.0.12.34:2222"))

			Expect(fakeTerm.SetRawTerminalCallCount()).To(Equal(1))
			Expect(fakeTerm.SetRawTerminalArgsForCall(0)).To(Equal(os.Stdin.Fd()))

			Expect(fakeTerm.RestoreTerminalCallCount()).To(Equal(1))
			fd, state := fakeTerm.RestoreTerminalArgsForCall(0)
			Expect(fd).To(Equal(os.Stdin.Fd()))
			Expect(state).To(Equal(termState))

			Expect(fakeSession.ShellCallCount()).To(Equal(1))
			Expect(fakeSession.WaitCallCount()).To(Equal(1))
		})

		Context("when the SecureDialer#Dial fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(nil, errors.New("cannot dial error"))

				err := secureShell.ConnectToShell("app-name", 2, config)
				Expect(err).To(MatchError("cannot dial error"))
			})
		})

		Context("when the SecureSession#StdinPipe fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(nil, errors.New("put this in your pipe"))

				err := secureShell.ConnectToShell("app-name", 2, config)
				Expect(err).To(MatchError("put this in your pipe"))
			})
		})

		Context("when the SecureSession#StdoutPipe fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(fakeStdin, nil)
				fakeSession.StdoutPipeReturns(nil, errors.New("put this in your pipe"))

				err := secureShell.ConnectToShell("app-name", 2, config)
				Expect(err).To(MatchError("put this in your pipe"))
			})
		})

		Context("when the SecureSession#StderrPipe fails", func() {
			It("returns an error", func() {
				fakeDialer.DialReturns(fakeSession, nil)
				fakeSession.StdinPipeReturns(fakeStdin, nil)
				fakeSession.StdoutPipeReturns(fakeStdout, nil)
				fakeSession.StderrPipeReturns(nil, errors.New("put this in your pipe"))

				err := secureShell.ConnectToShell("app-name", 2, config)
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

				err := secureShell.ConnectToShell("app-name", 2, config)
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

				err := secureShell.ConnectToShell("app-name", 2, config)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTerm.RestoreTerminalCallCount()).To(Equal(0))
			})
		})
	})
})
