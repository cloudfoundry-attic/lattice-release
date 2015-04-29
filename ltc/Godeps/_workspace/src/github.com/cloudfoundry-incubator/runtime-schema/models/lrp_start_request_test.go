package models_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRPStartRequest", func() {
	var lrpStart models.LRPStartRequest
	var lrpStartPayload string

	BeforeEach(func() {
		lrpStartPayload = `{
    "desired_lrp": {
      "process_guid": "some-guid",
      "domain": "tests",
      "instances": 1,
      "start_timeout": 0,
      "rootfs": "docker:///docker.com/docker",
      "action": {"download": {
          "from": "http://example.com",
          "to": "/tmp/internet",
          "cache_key": ""
        }
      },
      "disk_mb": 512,
      "memory_mb": 1024,
      "cpu_weight": 42,
      "privileged": false,
      "ports": [
        5678
      ],
      "routes": {
        "router": {"port":5678,"hosts":["route-1","route-2"]}
      },
      "log_guid": "log-guid",
      "log_source": "the cloud",
      "metrics_guid": "metrics-guid",
      "modification_tag": {
        "epoch": "some-epoch",
        "index": 50
      }
    },
    "indices": [2]
  }`

		rawMessage := json.RawMessage([]byte(`{"port":5678,"hosts":["route-1","route-2"]}`))
		lrpStart = models.LRPStartRequest{
			Indices: []uint{2},

			DesiredLRP: models.DesiredLRP{
				Domain:      "tests",
				ProcessGuid: "some-guid",

				RootFS:    "docker:///docker.com/docker",
				Instances: 1,
				MemoryMB:  1024,
				DiskMB:    512,
				CPUWeight: 42,
				Routes: map[string]*json.RawMessage{
					"router": &rawMessage,
				},
				Ports: []uint16{
					5678,
				},
				LogGuid:     "log-guid",
				LogSource:   "the cloud",
				MetricsGuid: "metrics-guid",
				Action: &models.DownloadAction{
					From: "http://example.com",
					To:   "/tmp/internet",
				},
				ModificationTag: models.ModificationTag{
					Epoch: "some-epoch",
					Index: 50,
				},
			},
		}
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			jsonPayload, err := models.ToJSON(&lrpStart)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(jsonPayload)).To(MatchJSON(lrpStartPayload))
		})
	})

	Describe("FromJSON", func() {
		var decodedLRPStartRequest *models.LRPStartRequest
		var err error

		JustBeforeEach(func() {
			decodedLRPStartRequest = &models.LRPStartRequest{}
			err = models.FromJSON([]byte(lrpStartPayload), decodedLRPStartRequest)
		})

		It("returns a LRP with correct fields", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedLRPStartRequest).To(Equal(&lrpStart))
		})

		Context("with an invalid payload", func() {
			BeforeEach(func() {
				lrpStartPayload = "aliens lol"
			})

			It("returns the error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with an invalid desired lrp", func() {
			BeforeEach(func() {
				lrpStartPayload = `{
    "desired_lrp": {
      "domain": "tests",
      "instances": 1,
      "stack": "some-stack",
      "start_timeout": 0,
      "rootfs": "docker:///docker.com/docker",
      "action": {"download": {
          "from": "http://example.com",
          "to": "/tmp/internet",
          "cache_key": ""
        }
      },
      "disk_mb": 512,
      "memory_mb": 1024,
      "cpu_weight": 42,
      "ports": [
        5678
      ],
      "routes": {
        "router": {"port":5678,"hosts":["route-1","route-2"]}
      },
      "log_guid": "log-guid",
      "log_source": "the cloud"
    },
    "index": 2
  }`
			})

			It("returns a validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(ContainElement(models.ErrInvalidField{"process_guid"}))
			})
		})

		Context("with no indices", func() {
			BeforeEach(func() {
				lrpStartPayload = `{
    "desired_lrp": {
      "process_guid": "some-guid",
      "domain": "tests",
      "instances": 1,
      "stack": "some-stack",
      "start_timeout": 0,
      "rootfs": "docker:///docker.com/docker",
      "action": {"download": {
          "from": "http://example.com",
          "to": "/tmp/internet",
          "cache_key": ""
        }
      },
      "disk_mb": 512,
      "memory_mb": 1024,
      "cpu_weight": 42,
      "ports": [
        5678
      ],
      "routes": {
        "router": {"port":5678,"hosts":["route-1","route-2"]}
      },
      "log_guid": "log-guid",
      "log_source": "the cloud"
    }
  }`
			})

			It("returns a validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(ContainElement(models.ErrInvalidField{"indices"}))
			})
		})

		Context("with an invalid index", func() {
			BeforeEach(func() {
				lrpStartPayload = `{
    "desired_lrp": {
      "process_guid": "some-guid",
      "domain": "tests",
      "instances": 1,
      "stack": "some-stack",
      "start_timeout": 0,
      "rootfs": "docker:///docker.com/docker",
      "action": {"download": {
          "from": "http://example.com",
          "to": "/tmp/internet",
          "cache_key": ""
        }
      },
      "disk_mb": 512,
      "memory_mb": 1024,
      "cpu_weight": 42,
      "ports": [
        5678
      ],
      "routes": {
        "router": {"port":5678,"hosts":["route-1","route-2"]}
      },
      "log_guid": "log-guid",
      "log_source": "the cloud"
    },
    "indices": [-1]
  }`
			})

			It("returns a validation error", func() {
				Expect(err).To(BeAssignableToTypeOf(&json.UnmarshalTypeError{}))
			})
		})

	})
})
