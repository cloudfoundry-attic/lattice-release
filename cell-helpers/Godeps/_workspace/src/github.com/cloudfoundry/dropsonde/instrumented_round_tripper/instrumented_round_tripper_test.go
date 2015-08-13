package instrumented_round_tripper_test

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/dropsonde/instrumented_round_tripper"
	"github.com/cloudfoundry/sonde-go/events"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstrumentedRoundTripper", func() {
	var fakeRoundTripper *FakeRoundTripper
	var rt http.RoundTripper
	var req *http.Request
	var fakeEmitter *fake.FakeEventEmitter

	var origin = "testRoundtripper/42"

	BeforeEach(func() {
		var err error
		fakeEmitter = fake.NewFakeEventEmitter(origin)

		fakeRoundTripper = &FakeRoundTripper{}
		rt = instrumented_round_tripper.InstrumentedRoundTripper(fakeRoundTripper, fakeEmitter)

		req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
		Expect(err).ToNot(HaveOccurred())
		req.RemoteAddr = "127.0.0.1"
		req.Header.Set("User-Agent", "our-testing-client")
	})

	Context("when the round tripper is a cancelable round tripper", func() {
		var fcrt *fakeCancelableRoundTripper
		BeforeEach(func() {
			fcrt = &fakeCancelableRoundTripper{}
			rt = instrumented_round_tripper.InstrumentedRoundTripper(fcrt, fakeEmitter)
		})

		It("returns an instrumentedCancelableRoundTripper", func() {
			Expect(reflect.TypeOf(rt).Elem().Name()).To(Equal("instrumentedCancelableRoundTripper"))

			_, ok := rt.(canceler)
			Expect(ok).To(BeTrue())

			_, ok = rt.(http.RoundTripper)
			Expect(ok).To(BeTrue())
		})

		It("delegates CancelRequest", func() {
			Expect(fcrt.canceled).To(BeFalse())

			c := rt.(canceler)

			c.CancelRequest(nil)
			Expect(fcrt.canceled).To(BeTrue())
		})
	})

	Context("when the round tripper is not a cancelable round tripper", func() {
		BeforeEach(func() {
			fakeRoundTripper = &FakeRoundTripper{}
			rt = instrumented_round_tripper.InstrumentedRoundTripper(fakeRoundTripper, fakeEmitter)
		})

		It("returns an instrumentedRoundTripper", func() {
			Expect(reflect.TypeOf(rt).Elem().Name()).To(Equal("instrumentedRoundTripper"))

			_, ok := rt.(http.RoundTripper)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("request ID", func() {
		It("should generate a new request ID", func() {
			rt.RoundTrip(req)
			Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
		})

		Context("if request ID can't be generated", func() {
			BeforeEach(func() {
				instrumented_round_tripper.GenerateUuid = func() (u *uuid.UUID, err error) {
					return nil, errors.New("test error")
				}
			})
			AfterEach(func() {
				instrumented_round_tripper.GenerateUuid = uuid.NewV4
			})

			It("defaults to an empty request ID", func() {
				rt.RoundTrip(req)
				Expect(req.Header.Get("X-CF-RequestID")).To(Equal("00000000-0000-0000-0000-000000000000"))
			})
		})
	})

	Context("event emission", func() {
		It("should emit a start event", func() {
			rt.RoundTrip(req)
			Expect(fakeEmitter.GetMessages()[0].Event).To(BeAssignableToTypeOf(new(events.HttpStart)))
			Expect(fakeEmitter.GetMessages()[0].Origin).To(Equal("testRoundtripper/42"))
		})

		Context("if request ID already exists", func() {
			var existingRequestId *uuid.UUID

			BeforeEach(func() {
				existingRequestId, _ = uuid.NewV4()
				req.Header.Set("X-CF-RequestID", existingRequestId.String())
			})

			It("should emit the existing request ID as the parent request ID", func() {
				rt.RoundTrip(req)
				startEvent := fakeEmitter.GetMessages()[0].Event.(*events.HttpStart)
				Expect(startEvent.GetParentRequestId()).To(Equal(factories.NewUUID(existingRequestId)))
			})
		})

		Context("if round tripper returns an error", func() {
			It("should emit a stop event with blank response fields", func() {
				fakeRoundTripper.fakeError = errors.New("fakeEmitter error")
				rt.RoundTrip(req)

				Expect(fakeEmitter.GetMessages()[1].Event).To(BeAssignableToTypeOf(new(events.HttpStop)))

				stopEvent := fakeEmitter.GetMessages()[1].Event.(*events.HttpStop)
				Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 0))
				Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 0))
			})
		})

		Context("if round tripper does not return an error", func() {
			It("should emit a stop event with the round tripper's response", func() {
				rt.RoundTrip(req)

				Expect(fakeEmitter.GetMessages()[1].Event).To(BeAssignableToTypeOf(new(events.HttpStop)))

				stopEvent := fakeEmitter.GetMessages()[1].Event.(*events.HttpStop)
				Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 123))
				Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 1234))
			})
		})
	})

})

type FakeRoundTripper struct {
	fakeError error
}

func (frt *FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 123, ContentLength: 1234}, frt.fakeError
}

type fakeCancelableRoundTripper struct {
	fakeError error
	canceled  bool
}

func (frt *fakeCancelableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 123, ContentLength: 1234}, frt.fakeError
}

func (frt *fakeCancelableRoundTripper) CancelRequest(req *http.Request) {
	frt.canceled = true
}

type canceler interface {
	CancelRequest(*http.Request)
}
