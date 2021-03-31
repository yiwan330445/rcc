package common

import (
	"fmt"
	"time"
)

type stopwatch struct {
	message string
	started time.Time
}

type Duration time.Duration

func (it Duration) Truncate(granularity time.Duration) Duration {
	return Duration(time.Duration(it).Truncate(granularity))
}

func (it Duration) Milliseconds() int64 {
	return time.Duration(it).Milliseconds()
}

func (it Duration) String() string {
	return fmt.Sprintf("%5.3f", float64(it.Milliseconds())/1000.0)
}

func Stopwatch(form string, details ...interface{}) *stopwatch {
	message := fmt.Sprintf(form, details...)
	return &stopwatch{
		message: message,
		started: time.Now(),
	}
}

func (it *stopwatch) String() string {
	elapsed := it.Elapsed().Truncate(time.Millisecond)
	return fmt.Sprintf("%v", elapsed)
}

func (it *stopwatch) Elapsed() Duration {
	return Duration(time.Since(it.started))
}

func (it *stopwatch) Debug() Duration {
	elapsed := it.Elapsed()
	Debug("%v %v", it.message, elapsed)
	return elapsed
}

func (it *stopwatch) Log() Duration {
	elapsed := it.Elapsed()
	Log("%v %v", it.message, elapsed)
	return elapsed
}

func (it *stopwatch) Report() Duration {
	elapsed := it.Elapsed()
	Log("%v %v", it.message, elapsed)
	return elapsed
}
