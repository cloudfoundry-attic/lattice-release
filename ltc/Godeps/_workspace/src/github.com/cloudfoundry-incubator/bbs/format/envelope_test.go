package format_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/format/fakes"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Envelope", func() {
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
	})

	Describe("Marshal", func() {
		It("can successfully marshal a model object envelope", func() {
			task := model_helpers.NewValidTask("some-guid")
			encoded, err := format.MarshalEnvelope(format.PROTO, task)
			Expect(err).NotTo(HaveOccurred())

			Expect(format.EnvelopeFormat(encoded[0])).To(Equal(format.PROTO))
			Expect(format.Version(encoded[1])).To(Equal(format.V0))

			var newTask models.Task
			modelErr := proto.Unmarshal(encoded[2:], &newTask)
			Expect(modelErr).To(BeNil())

			Expect(*task).To(Equal(newTask))
		})

		Context("when model validation fails", func() {
			It("returns an error ", func() {
				model := &fakes.FakeVersioner{}
				model.ValidateReturns(errors.New("go away"))

				_, err := format.MarshalEnvelope(format.PROTO, model)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Unmarshal", func() {
		It("can marshal and unmarshal a task without losing data", func() {
			task := model_helpers.NewValidTask("some-guid")
			payload, err := format.MarshalEnvelope(format.PROTO, task)
			Expect(err).NotTo(HaveOccurred())

			resultingTask := new(models.Task)
			err = format.UnmarshalEnvelope(logger, payload, resultingTask)
			Expect(err).NotTo(HaveOccurred())

			Expect(*resultingTask).To(BeEquivalentTo(*task))
		})

		It("calls MigrateFromVersion on on the model object with the envelope version", func() {
			model := &fakes.FakeVersioner{}
			payload := []byte{byte(format.JSON), byte(format.V0), '{', '}'}

			err := format.UnmarshalEnvelope(logger, payload, model)
			Expect(err).NotTo(HaveOccurred())
			Expect(model.MigrateFromVersionCallCount()).To(Equal(1))
			Expect(model.MigrateFromVersionArgsForCall(0)).To(Equal(format.V0))
		})

		It("returns an error when the serialization format is unknown", func() {
			model := &fakes.FakeVersioner{}
			payload := []byte{byte(format.EnvelopeFormat(99)), byte(format.V0), '{', '}'}
			err := format.UnmarshalEnvelope(logger, payload, model)
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the json payload is invalid", func() {
			model := &fakes.FakeVersioner{}
			payload := []byte{byte(format.JSON), byte(format.V0), 'f', 'o', 'o'}
			err := format.UnmarshalEnvelope(logger, payload, model)
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the protobuf payload is invalid", func() {
			model := model_helpers.NewValidTask("foo")
			payload := []byte{byte(format.PROTO), byte(format.V0), 'f', 'o', 'o'}
			err := format.UnmarshalEnvelope(logger, payload, model)
			Expect(err).To(HaveOccurred())
		})
	})
})

func bytesForEnvelope(f format.EnvelopeFormat, v format.Version, payloads ...string) []byte {
	env := []byte{byte(f), byte(v)}
	for i := range payloads {
		env = append(env, []byte(payloads[i])...)
	}
	return env
}
