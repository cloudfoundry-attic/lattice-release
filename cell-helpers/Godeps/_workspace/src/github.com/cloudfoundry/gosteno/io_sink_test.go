package gosteno

import (
	"bufio"
	"io"
	. "launchpad.net/gocheck"
)

type IOSinkSuite struct {
}

var _ = Suite(&IOSinkSuite{})

func (s *IOSinkSuite) TestAddRecord(c *C) {
	pReader, pWriter := io.Pipe()
	sink := NewIOSink(nil)
	sink.writer = bufio.NewWriter(pWriter)
	sink.SetCodec(NewJsonCodec())

	go func(msg string) {
		record := NewRecord("source", LOG_INFO, msg, nil)
		sink.AddRecord(record)
		sink.Flush()
		pWriter.Close()
	}("Hello, \nworld")

	bufReader := bufio.NewReader(pReader)
	msg, err := bufReader.ReadString('\n')
	c.Assert(err, IsNil)
	c.Assert(msg, Matches, `{.*"Hello, \\nworld".*}\n`)
}
