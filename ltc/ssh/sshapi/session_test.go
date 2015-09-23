package sshapi_test

import (
	"bytes"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/sshapi"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/sshapi/mocks"
)

var _ = Describe("Session", func() {
	var (
		client      *sshapi.Client
		mockSession *mocks.FakeSSHSession
		session     *sshapi.Session
	)

	BeforeEach(func() {
		mockSessionFactory := &mocks.FakeSSHSessionFactory{}
		client = &sshapi.Client{
			SSHSessionFactory: mockSessionFactory,
			Stdin:             &bytes.Buffer{},
			Stdout:            &bytes.Buffer{},
			Stderr:            &bytes.Buffer{},
		}
		mockSession = &mocks.FakeSSHSession{}
		mockSession.StdinPipeReturns(&mockConn{Writer: &bytes.Buffer{}}, nil)
		mockSession.StdoutPipeReturns(&bytes.Buffer{}, nil)
		mockSession.StderrPipeReturns(&bytes.Buffer{}, nil)
		mockSessionFactory.NewReturns(mockSession, nil)
		var err error
		session, err = client.Open(0, 0)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#KeepAlive", func() {
		It("should send a request when the ticker ticks", func() {
			ticker := make(chan time.Time)
			session.KeepAliveTicker.C = ticker

			stopChan := session.KeepAlive()
			Consistently(mockSession.SendRequestCallCount).Should(Equal(0))
			ticker <- time.Time{}
			Eventually(mockSession.SendRequestCallCount).Should(Equal(1))
			ticker <- time.Time{}
			Eventually(mockSession.SendRequestCallCount).Should(Equal(2))
			stopChan <- struct{}{}
			go func() { ticker <- time.Time{} }()
			Consistently(mockSession.SendRequestCallCount).Should(Equal(2))

			name, reply, payload := mockSession.SendRequestArgsForCall(0)
			Expect(name).To(Equal("keepalive@cloudfoundry.org"))
			Expect(reply).To(BeTrue())
			Expect(payload).To(BeNil())
		})
	})

	Describe("#Resize", func() {
		It("should send a window-change request", func() {
			Expect(session.Resize(100, 200)).To(Succeed())
			Expect(mockSession.SendRequestCallCount()).To(Equal(1))
			name, reply, payload := mockSession.SendRequestArgsForCall(0)
			Expect(name).To(Equal("window-change"))
			Expect(reply).To(BeFalse())
			Expect(payload).To(Equal([]byte{0, 0, 0, 100, 0, 0, 0, 200, 0, 0, 0, 0, 0, 0, 0, 0}))
		})

		Context("when sending the request fails", func() {
			It("should return an error", func() {
				mockSession.SendRequestReturns(false, errors.New("some error"))
				err := session.Resize(100, 200)
				Expect(err).To(MatchError("some error"))
			})
		})
	})
})
