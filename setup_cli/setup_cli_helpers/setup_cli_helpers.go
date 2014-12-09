package setup_cli_helpers

import (
	"strconv"
	"time"
)

func Timeout(timeout string) time.Duration {
	if timeout, err := strconv.Atoi(timeout); err == nil {
		return time.Second * time.Duration(timeout)
	}

	return time.Minute
}

func LoggregatorUrl(loggregatorTarget string) string {
	return "ws://" + loggregatorTarget
}
