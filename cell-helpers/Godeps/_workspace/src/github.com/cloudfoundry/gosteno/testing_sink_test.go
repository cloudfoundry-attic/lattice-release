package gosteno

import (
	. "launchpad.net/gocheck"
)

type TestingSinkSuite struct {
}

var _ = Suite(&TestingSinkSuite{})

func (s *TestingSinkSuite) TestAddRecord(c *C) {
	sink := NewTestingSink()

	data := map[string]interface{}{"a": "b"}
	record := NewRecord("source", LOG_INFO, "Hello, world!", data)
	sink.AddRecord(record)

	c.Assert(sink.Records(), HasLen, 1)
	c.Check(sink.Records()[0].Message, Equals, "Hello, world!")
	c.Check(sink.Records()[0].Data, DeepEquals, data)
}
