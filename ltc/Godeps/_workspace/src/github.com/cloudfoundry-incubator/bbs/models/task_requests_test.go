package models_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task requests", func() {
	Describe("TaskByGuidRequest", func() {
		Describe("Validate", func() {
			var request models.TaskByGuidRequest

			BeforeEach(func() {
				request = models.TaskByGuidRequest{
					TaskGuid: "something",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the TaskGuid is blank", func() {
				BeforeEach(func() {
					request.TaskGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"task_guid"}))
				})
			})
		})
	})

	Describe("DesireTaskRequest", func() {
		Describe("Validate", func() {
			var request models.DesireTaskRequest

			BeforeEach(func() {
				request = models.DesireTaskRequest{
					TaskGuid:       "t-guid",
					Domain:         "domain",
					TaskDefinition: model_helpers.NewValidTaskDefinition(),
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the TaskGuid is blank", func() {
				BeforeEach(func() {
					request.TaskGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"task_guid"}))
				})
			})

			Context("when the domain is blank", func() {
				BeforeEach(func() {
					request.Domain = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"domain"}))
				})
			})

			Context("when the TaskDefinition is nil", func() {
				BeforeEach(func() {
					request.TaskDefinition = nil
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"task_definition"}))
				})
			})

			Context("when the TaskDefinition has an invalid field", func() {
				BeforeEach(func() {
					request.TaskDefinition.RootFs = ""
				})

				It("bubbles up the appropriate invalid field error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"rootfs"}))
				})
			})
		})
	})

	Describe("CompleteTaskRequest", func() {
		Describe("Validate", func() {
			var request models.CompleteTaskRequest

			BeforeEach(func() {
				request = models.CompleteTaskRequest{
					TaskGuid: "t-guid",
					CellId:   "c-id",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the TaskGuid is blank", func() {
				BeforeEach(func() {
					request.TaskGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"task_guid"}))
				})
			})

			Context("when the CellId is blank", func() {
				BeforeEach(func() {
					request.CellId = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"cell_id"}))
				})
			})
		})
	})

	Describe("FailTaskRequest", func() {
		Describe("Validate", func() {
			var request models.FailTaskRequest

			BeforeEach(func() {
				request = models.FailTaskRequest{
					TaskGuid:      "t-guid",
					FailureReason: "some-failure",
				}
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Expect(request.Validate()).To(BeNil())
				})
			})

			Context("when the TaskGuid is blank", func() {
				BeforeEach(func() {
					request.TaskGuid = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"task_guid"}))
				})
			})

			Context("when the FailureReason is blank", func() {
				BeforeEach(func() {
					request.FailureReason = ""
				})

				It("returns a validation error", func() {
					Expect(request.Validate()).To(ConsistOf(models.ErrInvalidField{"failure_reason"}))
				})
			})
		})
	})
})
