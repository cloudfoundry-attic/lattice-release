package gosteno

import (
	"encoding/json"
	. "launchpad.net/gocheck"
)

type LogLevelSuite struct {
}

var _ = Suite(&LogLevelSuite{})

func (s *LogLevelSuite) TestNewLogLevel(c *C) {
	level := defineLogLevel("foobar", 100)
	c.Assert(level, NotNil)
	c.Assert(level.Name, Equals, "foobar")
	c.Assert(level.Priority, Equals, 100)
}

func (s *LogLevelSuite) TestGetLevel(c *C) {
	l, err := GetLogLevel("info")
	c.Assert(l, Equals, LOG_INFO)
	c.Assert(err, IsNil)
}

func (s *LogLevelSuite) TestGetNotExistLevel(c *C) {
	l, err := GetLogLevel("foobar")
	c.Assert(l, Equals, LogLevel{})
	c.Assert(err, NotNil)
}

func (s *LogLevelSuite) TestMarshal(c *C) {
	m, err := json.Marshal(LOG_INFO)

	c.Assert(err, IsNil)
	c.Assert(string(m), Equals, "\"info\"")
}

func (s *LogLevelSuite) TestUnmarshal(c *C) {
	var l LogLevel

	err := json.Unmarshal([]byte("\"info\""), &l)

	c.Assert(err, IsNil)
	c.Assert(l, Equals, LOG_INFO)
}

func (s *LogLevelSuite) TestUnmarshalError(c *C) {
	var l LogLevel

	err := json.Unmarshal([]byte("\"undefined\""), &l)

	c.Assert(err.Error(), Matches, "Undefined log level: .*")
	c.Assert(l, Equals, LogLevel{})
}
