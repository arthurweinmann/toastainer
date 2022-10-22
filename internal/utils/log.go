package utils

import (
	"fmt"
	"os"
	"runtime/debug"
)

var loglevel int

func InitLogging(level string) {
	switch level {
	case "all":
		loglevel = -1
	case "debug":
		loglevel = 0
	case "info":
		loglevel = 1
	case "warn":
		loglevel = 2
	case "error":
		loglevel = 3
	case "fatal":
		loglevel = 4
	case "none":
		loglevel = 5
	default:
		loglevel = 2
	}
}

func Debug(keyvals ...interface{}) {
	if loglevel <= 0 {
		fmt.Printf("DEBUG >\n    ")
		fmt.Println(keyvals...)
	}
}

func Info(keyvals ...interface{}) {
	if loglevel <= 1 {
		fmt.Printf("INFO >\n    ")
		fmt.Println(keyvals...)
	}
}

func Warn(keyvals ...interface{}) {
	if loglevel <= 2 {
		fmt.Printf("WARNING >\n    ")
		fmt.Println(keyvals...)
	}
}

func Error(keyvals ...interface{}) {
	if loglevel <= 3 {
		fmt.Printf("ERROR >\n    ")
		fmt.Println(keyvals...)
	}
}

func Fatal(keyvals ...interface{}) {
	if loglevel <= 4 {
		fmt.Printf("FATAL >\n    ")
		fmt.Println(keyvals...)
		debug.PrintStack()
	}

	os.Exit(1)
}
