package common

import (
	"fmt"
	"log"
	"os"
)

var (
	logTool logging
)

type logging interface {
	Println(...interface{})
}

type flatLog bool

func (it flatLog) Println(values ...interface{}) {
	fmt.Println(values...)
}

func init() {
	logTool = flatLog(true)
}

func TrueLog() {
	logTool = log.New(os.Stdout, "| ", log.LstdFlags)
}

func Log(format string, details ...interface{}) {
	if Silent {
		return
	}
	logTool.Println(fmt.Sprintf(format, details...))
}
