package models_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP Requests", func() {
	Describe("ActualLRPGroupsRequest", func() {
		Describe("Validate", func() {
			var request models.ActualLRPGroupsRequest

			BeforeEach(func() {
				request = models.ActualLRPGroupsRequest{}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})
		})
	})

	Describe("ActualLRPGroupsByProcessGuidRequest", func() {
		Describe("Validate", func() {
			var request models.ActualLRPGroupsByProcessGuidRequest

			BeforeEach(func() {
				request = models.ActualLRPGroupsByProcessGuidRequest{
					ProcessGuid: "something",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					request.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})
		})
	})

	Describe("ActualLRPGroupByProcessGuidAndIndexRequest", func() {
		Describe("Validate", func() {
			var request models.ActualLRPGroupByProcessGuidAndIndexRequest

			BeforeEach(func() {
				request = models.ActualLRPGroupByProcessGuidAndIndexRequest{
					ProcessGuid: "something",
					Index:       5,
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					request.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the Index is negative", func() {
				BeforeEach(func() {
					request.Index = -1
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"index"}))
				})
			})
		})
	})

	Describe("RemoveActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.RemoveActualLRPRequest

			BeforeEach(func() {
				request = models.RemoveActualLRPRequest{
					ProcessGuid: "something",
					Index:       5,
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					request.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the Index is negative", func() {
				BeforeEach(func() {
					request.Index = -1
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"index"}))
				})
			})
		})
	})

	Describe("ClaimActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.ClaimActualLRPRequest

			BeforeEach(func() {
				request = models.ClaimActualLRPRequest{
					ProcessGuid:          "p-guid",
					Index:                2,
					ActualLrpInstanceKey: &models.ActualLRPInstanceKey{InstanceGuid: "i-guid", CellId: "c-id"},
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					request.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the ActualLrpInstanceKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_instance_key"}))
				})
			})

			Context("when the ActualLrpInstanceKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey.InstanceGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"instance_guid"}))
				})
			})
		})
	})

	Describe("StartActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.StartActualLRPRequest

			BeforeEach(func() {
				request = models.StartActualLRPRequest{
					ActualLrpKey:         &models.ActualLRPKey{ProcessGuid: "p-guid", Index: 2, Domain: "domain"},
					ActualLrpInstanceKey: &models.ActualLRPInstanceKey{InstanceGuid: "i-guid", CellId: "c-id"},
					ActualLrpNetInfo:     &models.ActualLRPNetInfo{Address: "addr"},
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ActualLrpKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_key"}))
				})
			})

			Context("when the ActualLrpKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the ActualLrpInstanceKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_instance_key"}))
				})
			})

			Context("when the ActualLrpInstanceKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey.InstanceGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"instance_guid"}))
				})
			})

			Context("when the ActualLrpNetInfo is blank", func() {
				BeforeEach(func() {
					request.ActualLrpNetInfo = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_net_info"}))
				})
			})

			Context("when the ActualLrpNetInfo is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpNetInfo.Address = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"address"}))
				})
			})
		})
	})

	Describe("CrashActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.CrashActualLRPRequest

			BeforeEach(func() {
				request = models.CrashActualLRPRequest{
					ActualLrpKey:         &models.ActualLRPKey{ProcessGuid: "p-guid", Index: 2, Domain: "domain"},
					ActualLrpInstanceKey: &models.ActualLRPInstanceKey{InstanceGuid: "i-guid", CellId: "c-id"},
					ErrorMessage:         "string",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ActualLrpKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_key"}))
				})
			})

			Context("when the ActualLrpKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the ActualLrpInstanceKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_instance_key"}))
				})
			})

			Context("when the ErrorMessage is blank", func() {
				BeforeEach(func() {
					request.ErrorMessage = ""
				})

				It("is still valid", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})
		})
	})

	Describe("RetireActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.RetireActualLRPRequest

			BeforeEach(func() {
				request = models.RetireActualLRPRequest{
					ActualLrpKey: &models.ActualLRPKey{ProcessGuid: "p-guid", Index: 2, Domain: "domain"},
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ActualLrpKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_key"}))
				})
			})

			Context("when the ActualLrpKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})
		})
	})

	Describe("FailActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.FailActualLRPRequest

			BeforeEach(func() {
				request = models.FailActualLRPRequest{
					ActualLrpKey: &models.ActualLRPKey{ProcessGuid: "p-guid", Index: 2, Domain: "domain"},
					ErrorMessage: "string",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ActualLrpKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_key"}))
				})
			})

			Context("when the ActualLrpKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the ErrorMessage is blank", func() {
				BeforeEach(func() {
					request.ErrorMessage = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"error_message"}))
				})
			})
		})
	})

	Describe("RemoveEvacuatingActualLRPRequest", func() {
		Describe("Validate", func() {
			var request models.RemoveEvacuatingActualLRPRequest

			BeforeEach(func() {
				request = models.RemoveEvacuatingActualLRPRequest{
					ActualLrpKey:         &models.ActualLRPKey{ProcessGuid: "p-guid", Index: 2, Domain: "domain"},
					ActualLrpInstanceKey: &models.ActualLRPInstanceKey{InstanceGuid: "i-guid", CellId: "cell-id"},
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the ActualLrpKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_key"}))
				})
			})

			Context("when the ActualLrpKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the ActualLrpInstanceKey is blank", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"actual_lrp_instance_key"}))
				})
			})

			Context("when the ActualLrpInstanceKey is invalid", func() {
				BeforeEach(func() {
					request.ActualLrpInstanceKey.InstanceGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"instance_guid"}))
				})
			})
		})
	})
})
