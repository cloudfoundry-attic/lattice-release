// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows,!plan9

package syslog

import (
	"errors"
	"net"
)

// unixSyslog opens a connection to the syslog daemon running on the
// local machine using a Unix domain socket.

func unixSyslog() (conn serverConn, err error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog"}
	var raddr string
	for _, network := range logTypes {
		for _, path := range logPaths {
			raddr = path
			conn, err := net.Dial(network, raddr)
			if err != nil {
				continue
			} else {
				return netConn{conn}, nil
			}
		}
	}
	return nil, errors.New("Could not connect to local syslog socket")
}
