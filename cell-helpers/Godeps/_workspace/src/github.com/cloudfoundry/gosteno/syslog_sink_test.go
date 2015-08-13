// +build !windows,!plan9

package gosteno

import (
	"fmt"
	. "launchpad.net/gocheck"
	"os"
	"strings"
)

type SyslogSinkSuite struct {
}

var _ = Suite(&SyslogSinkSuite{})

func (s *SyslogSinkSuite) TestTruncate(c *C) {
	msg := generateString(MaxMessageSize - 1)
	record := &Record{
		Message: msg,
	}
	truncate(record)
	c.Check(record.Message, Equals, msg)

	msg2 := generateString(MaxMessageSize + 1)
	record2 := &Record{
		Message: msg2,
	}
	truncate(record2)
	c.Check(len(record2.Message), Equals, MaxMessageSize)
	c.Check(record2.Message[:MaxMessageSize-3], Equals, msg2[:MaxMessageSize-3])
	c.Check(strings.HasSuffix(record2.Message, "..."), Equals, true)
}

func generateString(length int) (ret string) {
	file, _ := os.Open("/dev/urandom")
	b := make([]byte, length/2)
	file.Read(b)
	file.Close()

	if length%2 == 0 {
		ret = fmt.Sprintf("%x", b)
	} else {
		ret = fmt.Sprintf("0%x", b)
	}

	return
}
