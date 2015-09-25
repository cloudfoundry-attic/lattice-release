package models_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var update models.DesiredLRPUpdate
	var aJson models.DesiredLRPUpdate
	var aProto models.DesiredLRPUpdate

	itSerializes := func(routes *models.Routes) {
		BeforeEach(func() {
			update = models.DesiredLRPUpdate{
				Routes: routes,
			}

			b, err := json.Marshal(update)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(b, &aJson)
			Expect(err).NotTo(HaveOccurred())

			b, err = proto.Marshal(&update)
			Expect(err).NotTo(HaveOccurred())
			err = proto.Unmarshal(b, &aProto)
			Expect(err).NotTo(HaveOccurred())
		})

		It("marshals JSON properly", func() {
			Expect(update.Equal(&aJson)).To(BeTrue())
			Expect(update).To(Equal(aJson))
		})

		It("marshals Proto properly", func() {
			Expect(update.Equal(&aProto)).To(BeTrue())
			Expect(update).To(Equal(aProto))
		})
	}

	itSerializes(nil)
	itSerializes(&models.Routes{
		"abc": &(json.RawMessage{'"', 'd', '"'}),
		"def": &(json.RawMessage{'"', 'g', '"'}),
	})
})
