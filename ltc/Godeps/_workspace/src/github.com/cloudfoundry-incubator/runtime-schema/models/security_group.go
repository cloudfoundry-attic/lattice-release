package models

import (
	"errors"
	"net"
	"strings"
)

type ProtocolName string

const (
	TCPProtocol  ProtocolName = "tcp"
	UDPProtocol  ProtocolName = "udp"
	ICMPProtocol ProtocolName = "icmp"
	AllProtocol  ProtocolName = "all"

	maxPort int = 65535
)

var errInvalidIP = errors.New("Invalid IP")

type PortRange struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
}

type SecurityGroupRule struct {
	Protocol     ProtocolName `json:"protocol"`
	Destinations []string     `json:"destinations"`
	Ports        []uint16     `json:"ports,omitempty"`
	PortRange    *PortRange   `json:"port_range,omitempty"`
	IcmpInfo     *ICMPInfo    `json:"icmp_info,omitempty"`
	Log          bool         `json:"log"`
}

type ICMPInfo struct {
	Type int32 `json:"type"`
	Code int32 `json:"code"`
}

func (rule SecurityGroupRule) Validate() error {
	var validationError ValidationError

	switch rule.Protocol {
	case TCPProtocol:
		validationError = rule.validatePorts()
		if rule.IcmpInfo != nil {
			validationError = validationError.Append(ErrInvalidField{"icmp_info"})
		}
	case UDPProtocol:
		validationError = rule.validatePorts()
		if rule.IcmpInfo != nil {
			validationError = validationError.Append(ErrInvalidField{"icmp_info"})
		}
		if rule.Log == true {
			validationError = validationError.Append(ErrInvalidField{"log"})
		}
	case ICMPProtocol:
		if rule.PortRange != nil {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
		if rule.Ports != nil {
			validationError = validationError.Append(ErrInvalidField{"ports"})
		}
		if rule.IcmpInfo == nil {
			validationError = validationError.Append(ErrInvalidField{"icmp_info"})
		}
		if rule.Log == true {
			validationError = validationError.Append(ErrInvalidField{"log"})
		}
	case AllProtocol:
		if rule.PortRange != nil {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
		if rule.Ports != nil {
			validationError = validationError.Append(ErrInvalidField{"ports"})
		}
		if rule.IcmpInfo != nil {
			validationError = validationError.Append(ErrInvalidField{"icmp_info"})
		}
	default:
		validationError = validationError.Append(ErrInvalidField{"protocol"})
	}

	if err := rule.validateDestinations(); err != nil {
		validationError = validationError.Append(ErrInvalidField{"destinations [ " + err.Error() + " ]"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (rule SecurityGroupRule) validatePorts() ValidationError {
	var validationError ValidationError

	if rule.PortRange == nil && rule.Ports == nil {
		return validationError.Append(errors.New("Missing required field: ports or port_range"))
	}

	if rule.PortRange != nil && rule.Ports != nil {
		return validationError.Append(errors.New("Invalid: ports and port_range provided"))
	}

	if rule.PortRange != nil {
		if rule.PortRange.Start < 1 {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
		if rule.PortRange.End < 1 {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
		if rule.PortRange.Start > rule.PortRange.End {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
	}

	if rule.Ports != nil {
		if len(rule.Ports) == 0 {
			validationError = validationError.Append(ErrInvalidField{"ports"})
		}

		for _, p := range rule.Ports {
			if p < 1 {
				validationError = validationError.Append(ErrInvalidField{"ports"})
			}
		}
	}

	return validationError
}

func (rule SecurityGroupRule) validateDestinations() error {
	if len(rule.Destinations) == 0 {
		return errors.New("Must have at least 1 destination")
	}

	var validationError ValidationError

	for _, d := range rule.Destinations {
		n := strings.IndexAny(d, "-/")
		if n == -1 {
			if net.ParseIP(d) == nil {
				validationError = validationError.Append(errInvalidIP)
				continue
			}
		} else if d[n] == '/' {
			_, _, err := net.ParseCIDR(d)
			if err != nil {
				validationError = validationError.Append(err)
				continue
			}
		} else {
			firstIP := net.ParseIP(d[:n])
			secondIP := net.ParseIP(d[n+1:])
			if firstIP == nil || secondIP == nil {
				validationError = validationError.Append(errInvalidIP)
				continue
			}
			for i, b := range firstIP {
				if b < secondIP[i] {
					break
				}

				if b == secondIP[i] {
					continue
				}

				validationError = validationError.Append(errInvalidIP)
				continue
			}
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
