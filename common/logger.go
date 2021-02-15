package common

import (
	"fmt"
	"io"
	"os"
	"time"
)

func printout(out io.Writer, message string) {
	var stamp string
	if TraceFlag {
		stamp = time.Now().Format("02.150405.000 ")
	}
	fmt.Fprintf(out, "%s%s\n", stamp, message)
}

func Fatal(context string, err error) {
	if err != nil {
		printout(os.Stderr, fmt.Sprintf("Fatal [%s]: %v", context, err))
		os.Stderr.Sync()
	}
}

func Error(context string, err error) {
	if err != nil {
		Log("Error [%s]: %v", context, err)
	}
}

func Log(format string, details ...interface{}) {
	if !Silent {
		printout(os.Stderr, fmt.Sprintf(format, details...))
		os.Stderr.Sync()
	}
}

func Debug(format string, details ...interface{}) error {
	if DebugFlag {
		printout(os.Stderr, fmt.Sprintf(format, details...))
		os.Stderr.Sync()
	}
	return nil
}

func Trace(format string, details ...interface{}) error {
	if TraceFlag {
		printout(os.Stderr, fmt.Sprintf(format, details...))
		os.Stderr.Sync()
	}
	return nil
}

func Stdout(format string, details ...interface{}) {
	fmt.Fprintf(os.Stdout, format, details...)
	os.Stdout.Sync()
}
