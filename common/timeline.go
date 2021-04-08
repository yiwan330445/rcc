package common

import (
	"fmt"
)

var (
	TimelineEnabled bool
	pipe            chan string
	done            chan bool
)

type timevent struct {
	when Duration
	what string
}

func timeliner(events chan string, done chan bool) {
	history := make([]*timevent, 0, 100)
	for {
		event, ok := <-events
		if !ok {
			break
		}
		history = append(history, &timevent{Clock.Elapsed(), event})
	}
	death := Clock.Elapsed()
	if TimelineEnabled && death.Milliseconds() > 0 {
		history = append(history, &timevent{death, "Now."})
		Log("----  rcc timeline  ----")
		Log(" #  percent  seconds  event")
		for at, event := range history {
			permille := event.when * 1000 / death
			percent := float64(permille) / 10.0
			Log("%2d:  %5.1f%%  %7s  %s", at+1, percent, event.when, event.what)
		}
		Log("----  rcc timeline  ----")
	}
	close(done)
}

func init() {
	pipe = make(chan string)
	done = make(chan bool)
	go timeliner(pipe, done)
}

func IgnoreAllPanics() {
	recover()
}

func Timeline(form string, details ...interface{}) {
	defer IgnoreAllPanics()
	pipe <- fmt.Sprintf(form, details...)
}

func EndOfTimeline() {
	close(pipe)
	<-done
}
