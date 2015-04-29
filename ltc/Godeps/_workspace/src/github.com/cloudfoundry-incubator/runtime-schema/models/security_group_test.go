package models_test

import (
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroupRule", func() {
	var rule models.SecurityGroupRule

	rulePayload := `{
		"protocol": "tcp",
		"destinations": ["1.2.3.4/16"],
		"port_range": {
			"start": 1,
			"end": 1024
		},
		"log": false
	}`

	BeforeEach(func() {
		rule = models.SecurityGroupRule{
			Protocol:     models.TCPProtocol,
			Destinations: []string{"1.2.3.4/16"},
			PortRange: &models.PortRange{
				Start: 1,
				End:   1024,
			},
		}
	})

	Describe("To JSON", func() {
		It("should JSONify a rule", func() {
			json, err := models.ToJSON(&rule)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(json)).To(MatchJSON(rulePayload))
		})

		It("should JSONify icmp info", func() {
			icmpRule := models.SecurityGroupRule{
				Protocol:     models.ICMPProtocol,
				Destinations: []string{"1.2.3.4/16"},
				IcmpInfo:     &models.ICMPInfo{},
			}

			json, err := models.ToJSON(&icmpRule)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(json)).To(MatchJSON(`{"protocol": "icmp", "destinations": ["1.2.3.4/16"], "icmp_info": {"type":0,"code":0}, "log":false }`))
		})
	})

	Describe("Validation", func() {
		var (
			validationErr error

			protocol    string
			destination string

			ports     []uint16
			portRange *models.PortRange

			icmpInfo *models.ICMPInfo

			log bool
		)

		BeforeEach(func() {
			protocol = "tcp"
			destination = "8.8.8.8/16"

			ports = nil
			portRange = nil
			icmpInfo = nil
			log = false
		})

		JustBeforeEach(func() {
			rule = models.SecurityGroupRule{
				Protocol:     models.ProtocolName(protocol),
				Destinations: []string{destination},
				Ports:        ports,
				PortRange:    portRange,
				IcmpInfo:     icmpInfo,
				Log:          log,
			}

			validationErr = rule.Validate()
		})

		itAllowsPorts := func() {
			Describe("ports", func() {
				Context("with a valid port", func() {
					BeforeEach(func() {
						portRange = nil
						ports = []uint16{1}
					})

					It("passes validation and does not return an error", func() {
						Expect(validationErr).NotTo(HaveOccurred())
					})
				})

				Context("with an empty ports list", func() {
					BeforeEach(func() {
						portRange = nil
						ports = []uint16{}
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("ports")))
					})
				})

				Context("with an invalid port", func() {
					BeforeEach(func() {
						portRange = nil
						ports = []uint16{0}
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("ports")))
					})
				})

			})

			Describe("port range", func() {
				Context("when it is a valid port range", func() {
					BeforeEach(func() {
						ports = nil
						portRange = &models.PortRange{1, 65535}
					})

					It("passes validation and does not return an error", func() {
						Expect(validationErr).NotTo(HaveOccurred())
					})
				})

				Context("when port range has a start value greater than the end value", func() {
					BeforeEach(func() {
						ports = nil
						portRange = &models.PortRange{1024, 1}
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("port_range")))
					})
				})
			})

			Context("when ports and port range are provided", func() {
				BeforeEach(func() {
					portRange = &models.PortRange{1, 65535}
					ports = []uint16{1}
				})

				It("returns an error", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("Invalid: ports and port_range provided")))
				})
			})

			Context("when ports and port range are not provided", func() {
				BeforeEach(func() {
					portRange = nil
					ports = nil
				})

				It("returns an error", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("Missing required field: ports or port_range")))
				})
			})
		}

		itExpectsADestination := func() {
			Describe("destination", func() {
				Context("when the destination is valid", func() {
					BeforeEach(func() {
						destination = "1.2.3.4/32"
					})

					It("passes validation and does not return an error", func() {
						Expect(validationErr).NotTo(HaveOccurred())
					})
				})

				Context("when the destination is invalid", func() {
					BeforeEach(func() {
						destination = "garbage/32"
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("destination")))
					})
				})
			})
		}

		itFailsWithPorts := func() {
			Context("when Port range is provided", func() {
				BeforeEach(func() {
					ports = nil
					portRange = &models.PortRange{1, 65535}
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("port_range")))
				})
			})

			Context("when Ports are provided", func() {
				BeforeEach(func() {
					portRange = nil
					ports = []uint16{1}
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("ports")))
				})
			})
		}

		itFailsWithICMPInfo := func() {
			Context("when ICMP info is provided", func() {
				BeforeEach(func() {
					icmpInfo = &models.ICMPInfo{}
				})
				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("icmp_info")))
				})
			})
		}

		itAllowsLogging := func() {
			Context("when log is true", func() {
				BeforeEach(func() {
					log = true
				})

				It("succeeds", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})
			})
		}

		itDisallowsLogging := func() {
			Context("when log is true", func() {
				BeforeEach(func() {
					log = true
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("log")))
				})
			})
		}

		Describe("destination", func() {
			BeforeEach(func() {
				ports = []uint16{1}
			})

			Context("when its an IP Address", func() {
				BeforeEach(func() {
					destination = "8.8.8.8"
				})

				It("passes validation and does not return an error", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})
			})

			Context("when its a range of IP Addresses", func() {
				BeforeEach(func() {
					destination = "8.8.8.8-8.8.8.9"
				})

				It("passes validation and does not return an error", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})

				Context("and the range is not valid", func() {
					BeforeEach(func() {
						destination = "1.2.3.4 - 1.2.1.3"
					})

					It("fails", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("destination")))
					})
				})
			})

			Context("when its a CIDR", func() {
				BeforeEach(func() {
					destination = "8.8.8.8/16"
				})

				It("passes validation and does not return an error", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})
			})

			Context("when its not valid", func() {
				BeforeEach(func() {
					destination = "8.8"
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("destination")))
				})
			})
		})

		Describe("protocol", func() {
			Context("when the protocol is tcp", func() {
				BeforeEach(func() {
					protocol = "tcp"
					ports = []uint16{1}
				})

				itFailsWithICMPInfo()
				itAllowsPorts()
				itExpectsADestination()
				itAllowsLogging()
			})

			Context("when the protocol is udp", func() {
				BeforeEach(func() {
					protocol = "udp"
					ports = []uint16{1}
				})

				itFailsWithICMPInfo()
				itAllowsPorts()
				itExpectsADestination()
				itDisallowsLogging()
			})

			Context("when the protocol is icmp", func() {
				BeforeEach(func() {
					protocol = "icmp"
					icmpInfo = &models.ICMPInfo{}
				})

				itExpectsADestination()
				itFailsWithPorts()
				itDisallowsLogging()

				Context("when no ICMPInfo is provided", func() {
					BeforeEach(func() {
						icmpInfo = nil
					})

					It("fails", func() {
						Expect(validationErr).To(HaveOccurred())
					})
				})
			})

			Context("when the protocol is all", func() {
				BeforeEach(func() {
					protocol = "all"
				})

				itFailsWithICMPInfo()
				itExpectsADestination()
				itFailsWithPorts()
				itAllowsLogging()
			})

			Context("when the protocol is invalid", func() {
				BeforeEach(func() {
					protocol = "foo"
				})

				It("returns an error", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("protocol")))
				})
			})
		})

		Context("when thre are multiple field validations", func() {
			BeforeEach(func() {
				protocol = "tcp"
				destination = "garbage"
				portRange = &models.PortRange{443, 80}
			})

			It("aggregates validation errors", func() {
				Expect(validationErr).To(MatchError(ContainSubstring("port_range")))
				Expect(validationErr).To(MatchError(ContainSubstring("destination")))
			})
		})
	})
})
