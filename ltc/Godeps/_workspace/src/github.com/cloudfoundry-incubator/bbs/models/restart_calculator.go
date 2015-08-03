package models

import (
	"fmt"
	"math"
	"time"
)

const DefaultImmediateRestarts = 3
const DefaultMaxBackoffDuration = 16 * time.Minute
const DefaultMaxRestarts = 200

const CrashBackoffMinDuration = 30 * time.Second

func exponentialBackoff(exponent, max int32) time.Duration {
	if exponent > max {
		exponent = max
	}
	return CrashBackoffMinDuration * time.Duration(powerOfTwo(exponent))
}

func powerOfTwo(pow int32) int32 {
	if pow < 0 {
		panic("pow cannot be negative")
	}
	return 1 << uint(pow)
}

func calculateMaxBackoffCount(maxDuration time.Duration) int32 {
	total := math.Ceil(float64(maxDuration) / float64(CrashBackoffMinDuration))
	return int32(math.Logb(total))
}

type RestartCalculator struct {
	ImmediateRestarts  int32         `json:"immediate_restarts"`
	MaxBackoffCount    int32         `json:"max_backoff_count"`
	MaxBackoffDuration time.Duration `json:"max_backoff_duration"`
	MaxRestartAttempts int32         `json:"max_restart_attempts"`
}

func NewDefaultRestartCalculator() RestartCalculator {
	return NewRestartCalculator(DefaultImmediateRestarts, DefaultMaxBackoffDuration, DefaultMaxRestarts)
}

func NewRestartCalculator(immediateRestarts int32, maxBackoffDuration time.Duration, maxRestarts int32) RestartCalculator {
	return RestartCalculator{
		ImmediateRestarts:  immediateRestarts,
		MaxBackoffDuration: maxBackoffDuration,
		MaxBackoffCount:    calculateMaxBackoffCount(maxBackoffDuration),
		MaxRestartAttempts: maxRestarts,
	}
}

func (r RestartCalculator) Validate() error {
	var validationError ValidationError
	if r.MaxBackoffDuration < CrashBackoffMinDuration {
		err := fmt.Errorf("MaxBackoffDuration '%s' must be larger than CrashBackoffMinDuration '%s'", r.MaxBackoffDuration, CrashBackoffMinDuration)
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (r RestartCalculator) ShouldRestart(now, crashedAt int64, crashCount int32) bool {
	switch {
	case crashCount < r.ImmediateRestarts:
		return true

	case crashCount < r.MaxRestartAttempts:
		backoffDuration := exponentialBackoff(crashCount-r.ImmediateRestarts, r.MaxBackoffCount)
		if backoffDuration > r.MaxBackoffDuration {
			backoffDuration = r.MaxBackoffDuration
		}
		nextRestartTime := crashedAt + backoffDuration.Nanoseconds()
		if nextRestartTime <= now {
			return true
		}
	}

	return false
}
