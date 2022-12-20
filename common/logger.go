package common

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	logsource  = make(logwriters)
	logbarrier = sync.WaitGroup{}
)

type logwriter func() (*os.File, string)
type logwriters chan logwriter

func loggerLoop(writers logwriters) {
	var stamp string
	line := uint64(0)
	for {
		line += 1
		todo, ok := <-writers
		if !ok {
			continue
		}
		out, message := todo()

		if TraceFlag {
			stamp = time.Now().Format("02.150405.000 ")
		} else if LogLinenumbers {
			stamp = fmt.Sprintf("%3d ", line)
		} else {
			stamp = ""
		}
		fmt.Fprintf(out, "%s%s\n", stamp, message)
		out.Sync()
		logbarrier.Done()
	}
}

func init() {
	go loggerLoop(logsource)
}

func printout(out *os.File, message string) {
	logbarrier.Add(1)
	logsource <- func() (*os.File, string) {
		return out, message
	}
}

func Fatal(context string, err error) {
	if err != nil {
		printout(os.Stderr, fmt.Sprintf("Fatal [%s]: %v", context, err))
	}
}

func Error(context string, err error) {
	if err != nil {
		Log("Error [%s]: %v", context, err)
	}
}

func Log(format string, details ...interface{}) {
	if !Silent {
		prefix := ""
		if DebugFlag || TraceFlag {
			prefix = "[N] "
		}
		printout(os.Stderr, fmt.Sprintf(prefix+format, details...))
	}
}

func Debug(format string, details ...interface{}) error {
	if DebugFlag {
		printout(os.Stderr, fmt.Sprintf("[D] "+format, details...))
	}
	return nil
}

func Trace(format string, details ...interface{}) error {
	if TraceFlag {
		printout(os.Stderr, fmt.Sprintf("[T] "+format, details...))
	}
	return nil
}

func Stdout(format string, details ...interface{}) {
	fmt.Fprintf(os.Stdout, format, details...)
	os.Stdout.Sync()
}

func WaitLogs() {
	defer Timeline("wait logs done")

	runtime.Gosched()
	logbarrier.Wait()
}

func Progress(step int, form string, details ...interface{}) {
	previous := ProgressMark
	ProgressMark = time.Now()
	delta := ProgressMark.Sub(previous).Round(1 * time.Millisecond).Seconds()
	message := fmt.Sprintf(form, details...)
	Log("####  Progress: %02d/13  %s  %8.3fs  %s", step, Version, delta, message)
	Timeline("%d/13 %s", step, message)
}
