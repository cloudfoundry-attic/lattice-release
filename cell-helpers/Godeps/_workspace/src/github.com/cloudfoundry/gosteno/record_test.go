package gosteno

import (
	. "launchpad.net/gocheck"
)

type RecordSuite struct{}

var _ = Suite(&RecordSuite{})

func (s *RecordSuite) TestNewRecordWithLOC(c *C) {
	config.EnableLOC = true
	r := NewRecord("source", LOG_INFO, "hello", nil)
	config.EnableLOC = false

	c.Check(r.File, Matches, ".*record_test.go$")
	c.Check(r.Line, Equals, 13)
	c.Check(r.Method, Matches, `.*\.\(\*RecordSuite\)\.TestNewRecordWithLOC`)
}

func (s *RecordSuite) TestNewRecordWithoutLOC(c *C) {
	r := NewRecord("source", LOG_INFO, "hello", nil)

	c.Check(r.File, Equals, "")
	c.Check(r.Line, Equals, 0)
	c.Check(r.Method, Equals, "")
}

func (s *RecordSuite) TestRecordPid(c *C) {
	r := NewRecord("source", LOG_INFO, "hello", nil)

	c.Check(r.Pid, Not(Equals), 0)
}
