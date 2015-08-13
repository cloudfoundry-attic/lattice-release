package gosteno

import (
	"testing"
)

func BenchmarkNoSink(b *testing.B) {
	Init(&Config{})

	loggers = make(map[string]*BaseLogger)
	logger := NewLogger("nosink")

	performBenchmark(logger, b)
}

func BenchmarkDevNullSink(b *testing.B) {
	Init(&Config{
		Sinks: []Sink{NewFileSink("/dev/null")},
	})

	loggers = make(map[string]*BaseLogger)
	logger := NewLogger("dev_null_sink")

	performBenchmark(logger, b)
}

func BenchmarkDevNullSinkWithLOC(b *testing.B) {
	Init(&Config{
		Sinks:     []Sink{NewFileSink("/dev/null")},
		EnableLOC: true,
	})

	loggers = make(map[string]*BaseLogger)
	logger := NewLogger("dev_null_sink_with_loc")

	performBenchmark(logger, b)
}

func BenchmarkDevNullSinkWithData(b *testing.B) {
	Init(&Config{
		Sinks: []Sink{NewFileSink("/dev/null")},
	})

	loggers = make(map[string]*BaseLogger)
	logger := NewLogger("dev_null_sink_with_data")
	logger.Set("key1", "value1")
	logger.Set("key2", "value2")

	performBenchmark(logger, b)
}

func performBenchmark(logger *Logger, b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Hello, world.")
	}
}
