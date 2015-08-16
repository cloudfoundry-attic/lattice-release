package gosteno

import (
	. "launchpad.net/gocheck"
)

type ConfigSuite struct {
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) TestReInitLevel(c *C) {
	levels := []LogLevel{LOG_INFO, LOG_DEBUG, LOG_WARN}

	for _, level := range levels {
		Init(&Config{Level: level})

		l := NewLogger("reinit")
		c.Assert(l.Level(), Equals, level)
	}
}
