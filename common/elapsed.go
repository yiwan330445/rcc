package common

import (
	"fmt"
	"time"
)

type stopwatch struct {
	message string
	started time.Time
}

func Stopwatch(form string, details ...interface{}) *stopwatch {
	message := fmt.Sprintf(form, details...)
	return &stopwatch{
		message: message,
		started: time.Now(),
	}
}

func (it *stopwatch) String() string {
	elapsed := time.Now().Sub(it.started)
	return fmt.Sprintf("%v", elapsed)
}

func (it *stopwatch) Log() time.Duration {
	elapsed := time.Now().Sub(it.started)
	Log("%v %v", it.message, elapsed)
	return elapsed
}

func (it *stopwatch) Report() time.Duration {
	elapsed := time.Now().Sub(it.started)
	Log("%v %v", it.message, elapsed)
	return elapsed
}
