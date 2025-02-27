package log

import (
	"fmt"
	"strings"
)

const (
	// Nothing at all
	LogLevelDisabled = iota
	// Errors and Warnings only
	LogLevelQuiet
	// generic info like "opening workspace ..."
	LogLevelInfo
	// obvious
	LogLevelDebug
)

var LogLevel int = LogLevelInfo

func addNewLineIfMissing(fmtString string) string {
	if !strings.HasSuffix(fmtString, "\n") {
		return fmtString + "\n"
	}
	return fmtString
}

func customPrintF(fmtString string, args ...interface{}) {
	fmtString = addNewLineIfMissing(fmtString)
	fmt.Printf(fmtString, args...)
}

func Info(fmtString string, args ...interface{}) {
	if LogLevel < LogLevelInfo {
		return
	}
	customPrintF(fmtString, args...)
}

func Error(fmtString string, args ...interface{}) {
	if LogLevel < LogLevelQuiet {
		return
	}
	customPrintF("[ERROR]: "+fmtString, args...)
}

func Warn(fmtString string, args ...interface{}) {
	if LogLevel < LogLevelQuiet {
		return
	}
	customPrintF(fmtString, args...)
}

func Debug(fmtString string, args ...interface{}) {
	if LogLevel < LogLevelDebug {
		return
	}
	customPrintF("[DEBUG]: "+fmtString, args...)
}
