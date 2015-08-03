package models

import (
	"errors"
	"net"
	"strings"
)

const (
	TCPProtocol  = "tcp"
	UDPProtocol  = "udp"
	ICMPProtocol = "icmp"
	AllProtocol  = "all"

	maxPort int = 65535
)

var errInvalidIP = errors.New("Invalid IP")

func (rule SecurityGroupRule) Validate() error {
	var validationError ValidationError

	switch rule.GetProtocol() {
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
		if rule.GetLog() == true {
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
		if rule.GetLog() == true {
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
		if rule.GetPortRange().GetStart() < 1 {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
		if rule.GetPortRange().GetEnd() < 1 {
			validationError = validationError.Append(ErrInvalidField{"port_range"})
		}
		if rule.GetPortRange().GetStart() > rule.GetPortRange().GetEnd() {
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
