package gosteno

import (
	"fmt"
	. "launchpad.net/gocheck"
)

type JsonCodecSuite struct {
	JsonCodec
}

var _ = Suite(&JsonCodecSuite{})

func (s *JsonCodecSuite) TestJsonCodec(c *C) {
	r := NewRecord("source", LOG_INFO, "Hello world", nil)
	m, err := s.EncodeRecord(r)
	c.Assert(err, IsNil)

	// The names of these fields are a direct copy of the fields used by the Ruby
	// version of steno to prevent breaking the prettifiers. Some of these fields
	// can be changed (e.g.  `process_id` -> `pid`, `log_level` -> `level), but
	// only as long as the prettifiers are also updated.
	fields := []string{
		"timestamp",
		"process_id",
		"source",
		"log_level",
		"message",
		"data",
	}

	for _, f := range fields {
		c.Check(string(m), Matches, fmt.Sprintf(`{.*"%s":.*}`, f))
	}
}

func (s *JsonCodecSuite) TestTimestampIsFormattedAsFloat(c *C) {
	r := NewRecord("source", LOG_INFO, "Hello world", nil)
	m, err := s.EncodeRecord(r)
	c.Assert(err, IsNil)
	c.Assert(string(m), Matches, `.*"timestamp":\d{10}\.\d{9},.*`)
}

func (s *JsonCodecSuite) TestEmptyFileLineMethodNotIncluded(c *C) {
	r := NewRecord("source", LOG_INFO, "Hello world", nil)
	m, err := s.EncodeRecord(r)
	c.Assert(err, IsNil)
	c.Check(string(m), Not(Matches), `.*"file":.*`)
	c.Check(string(m), Not(Matches), `.*"line":.*`)
	c.Check(string(m), Not(Matches), `.*"method":.*`)
}
