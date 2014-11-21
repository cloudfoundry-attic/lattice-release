package logs_helpers

func LoggregatorUrl(loggregatorTarget string) string {
	return "ws://" + loggregatorTarget
}
