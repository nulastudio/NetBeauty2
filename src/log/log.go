package log

import (
	"fmt"
	"os"
)

type LogLevel int

const (
	Error LogLevel = iota
	Detail
	Info
)

type Logger struct {
	LogLevel LogLevel
}

var DefaultLogger = &Logger{Info}

func (logger *Logger) Log(message string, level LogLevel) {
	if logger.LogLevel >= level {
		if logger.LogLevel == Error {
			message = "Error: " + message
		}
		fmt.Println(message)
	}
}

func (logger *Logger) PanicLog(message string, level LogLevel, code int) {
	logger.Log(message, level)
	os.Exit(code)
}

func LogError(err error, panic bool) {
	code := 0
	if panic {
		code = 1
	}
	LogPanic(err, code)
}

func LogPanic(err error, code int) {
	if err != nil {
		if code != 0 {
			DefaultLogger.PanicLog(err.Error(), Error, code)
		} else {
			DefaultLogger.Log(err.Error(), Error)
		}
	}
}

func LogInfo(message string) {
	DefaultLogger.Log(message, Info)
}

func LogDetail(message string) {
	DefaultLogger.Log(message, Detail)
}
