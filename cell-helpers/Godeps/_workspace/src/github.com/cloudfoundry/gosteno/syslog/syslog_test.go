// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows,!plan9

package syslog

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"testing"
	"time"
)

var serverAddr string

func runSyslog(c net.PacketConn, done chan<- string) {
	var buf [4096]byte
	var rcvd string = ""
	for {
		n, _, err := c.ReadFrom(buf[0:])
		if err != nil || n == 0 {
			break
		}
		rcvd += string(buf[0:n])
	}
	done <- rcvd
}

func startServer(done chan<- string) {
	c, e := net.ListenPacket("udp", "127.0.0.1:0")
	if e != nil {
		log.Fatalf("net.ListenPacket failed udp :0 %v", e)
	}
	serverAddr = c.LocalAddr().String()
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	go runSyslog(c, done)
}

func skipNetTest(t *testing.T) bool {
	if testing.Short() {
		// Depends on syslog daemon running, and sometimes it's not.
		t.Logf("skipping syslog test during -short")
		return true
	}
	return false
}

func TestNew(t *testing.T) {
	if skipNetTest(t) {
		return
	}
	s, err := New(LOG_INFO, "")
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	// Don't send any messages.
	s.Close()
}

func TestNewLogger(t *testing.T) {
	if skipNetTest(t) {
		return
	}
	f, err := NewLogger(LOG_INFO, 0)
	if f == nil {
		t.Error(err)
	}
}

func TestDial(t *testing.T) {
	if skipNetTest(t) {
		return
	}
	l, err := Dial("", "", LOG_ERR, "syslog_test")
	if err != nil {
		t.Fatalf("Dial() failed: %s", err)
	}
	l.Close()
}

func TestUDPDial(t *testing.T) {
	done := make(chan string)
	startServer(done)
	l, err := Dial("udp", serverAddr, LOG_INFO, "syslog_test")
	if err != nil {
		t.Fatalf("syslog.Dial() failed: %s", err)
	}
	msg := "udp test"
	l.Info(msg)
	expected := "<6>syslog_test: udp test\n"
	rcvd := <-done
	if rcvd != expected {
		t.Fatalf("s.Info() = '%q', but wanted '%q'", rcvd, expected)
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		pri Priority
		pre string
		msg string
		exp string
	}{
		{LOG_ERR, "syslog_test", "", "<3>syslog_test: \n"},
		{LOG_ERR, "syslog_test", "write test", "<3>syslog_test: write test\n"},
		// Write should not add \n if there already is one
		{LOG_ERR, "syslog_test", "write test 2\n", "<3>syslog_test: write test 2\n"},
	}

	for _, test := range tests {
		done := make(chan string)
		startServer(done)
		l, err := Dial("udp", serverAddr, test.pri, test.pre)
		if err != nil {
			t.Fatalf("syslog.Dial() failed: %s", err)
		}
		_, err = io.WriteString(l, test.msg)
		if err != nil {
			t.Fatalf("WriteString() failed: %s", err)
		}
		rcvd := <-done
		if rcvd != test.exp {
			t.Fatalf("s.Info() = '%q', but wanted '%q'", rcvd, test.exp)
		}
	}
}

func TestWriteTimesOut(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "write-timeout-test")
	if err != nil {
		t.Fatalf("ioutil.TempFile() failed: %s", err)
	}

	socketFile := path.Join(tmpdir, "test.sock")

	_, err = net.Listen("unix", socketFile)
	if err != nil {
		t.Fatalf("net.Listen() failed to listen on socket: %s", err)
	}

	w, err := Dial("unix", socketFile, LOG_INFO, "syslog_test")

	if err != nil {
		t.Fatalf("syslog.Dial() failed: %s", err)
	}

	firstFailure := make(chan error)

	go func() {
		for {
			_, err = io.WriteString(w, "timing out")
			if err != nil {
				firstFailure <- err
				break
			}
		}
	}()

	select {
	case err := <-firstFailure:
		switch err.(type) {
		case *net.OpError:
			if !err.(*net.OpError).Timeout() {
				t.Fatalf("WriteString() error = '%#v', but wanted a timeout error", err)
			}
		default:
			t.Fatalf("WriteString() error = '%#v', but wanted a timeout error", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("WriteString() never returned an error")
	}
}
