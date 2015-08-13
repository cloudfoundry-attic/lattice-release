package gosteno

import (
	"fmt"
	. "launchpad.net/gocheck"
)

type JsonPrettifierSuite struct {
}

var _ = Suite(&JsonPrettifierSuite{})

func (s *JsonPrettifierSuite) TestConst(c *C) {
	c.Assert(EXCLUDE_NONE, Equals, 0)
	c.Assert(EXCLUDE_LEVEL, Equals, 1<<0)
	c.Assert(EXCLUDE_TIMESTAMP, Equals, 1<<1)
	c.Assert(EXCLUDE_LINE, Equals, 1<<3)
}

func (s *JsonPrettifierSuite) TestConstOrder(c *C) {
	record := NewRecord("source", LOG_INFO, "Hello, world", map[string]interface{}{"foo": "bar"})

	prettifier1 := NewJsonPrettifier(EXCLUDE_FILE | EXCLUDE_DATA)
	bytes1, _ := prettifier1.EncodeRecord(record)

	prettifier2 := NewJsonPrettifier(EXCLUDE_DATA | EXCLUDE_FILE)
	bytes2, _ := prettifier2.EncodeRecord(record)

	c.Assert(string(bytes1), Equals, string(bytes2))
}

func (s *JsonPrettifierSuite) TestEncodeRecord(c *C) {
	config.EnableLOC = true
	record := NewRecord("source", LOG_INFO, "Hello, world", map[string]interface{}{"foo": "bar"})
	config.EnableLOC = false
	l := record.Line

	prettifier := NewJsonPrettifier(EXCLUDE_NONE)
	b, err := prettifier.EncodeRecord(record)
	c.Assert(err, IsNil)

	// One example:
	// INFO 2012-09-27 16:53:40 json_prettifier_test.go:34:TestEncodeRecord {"foo":"bar"} Hello, world
	c.Assert(string(b), Matches, fmt.Sprintf(`INFO .*son_prettifier_test.go:%d:TestEncodeRecord.*{"foo":"bar"}.*Hello, world`, l))
}

func (s *JsonPrettifierSuite) TestExclude(c *C) {
	config.EnableLOC = true
	record := NewRecord("source", LOG_INFO, "Hello, world", map[string]interface{}{"foo": "bar"})
	config.EnableLOC = false

	prettifier := NewJsonPrettifier(EXCLUDE_DATA | EXCLUDE_LINE)
	bytes, _ := prettifier.EncodeRecord(record)

	// One example:
	// INFO Wed, 19 Sep 2012 10:51:57 CST json_prettifier_test.go:TestExclude Hello, world
	c.Assert(string(bytes), Matches, `INFO .*son_prettifier_test.go:TestExclude Hello, world`)
}

func (s *JsonPrettifierSuite) TestDecodeLogEntry(c *C) {
	config.EnableLOC = true
	record := NewRecord("source", LOG_INFO, "Hello, world", map[string]interface{}{"foo": "bar"})
	config.EnableLOC = false
	l := record.Line
	t := record.Timestamp
	b, _ := NewJsonCodec().EncodeRecord(record)
	entry := string(b)

	prettifier := NewJsonPrettifier(EXCLUDE_NONE)
	record, err := prettifier.DecodeJsonLogEntry(entry)

	c.Assert(err, IsNil)
	c.Assert(record.Timestamp, Equals, t)
	c.Assert(record.Line, Equals, l)
	c.Assert(record.Level, Equals, LOG_INFO)
	c.Assert(record.Method, Matches, ".*TestDecodeLogEntry$")
	c.Assert(record.Message, Equals, "Hello, world")
	c.Assert(record.File, Matches, ".*json_prettifier_test.go")
	c.Assert(record.Data["foo"], Equals, "bar")

	// test err with invalid log level
	entry = `{"Message":"hi","Log_Level":"enoent"}`
	record, err = prettifier.DecodeJsonLogEntry(entry)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Matches, ".*enoent.*")
}
