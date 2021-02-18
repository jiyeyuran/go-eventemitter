package eventemitter

import (
	"fmt"
	"os"
)

// Logger defines interface to print log information.
type Logger interface {
	Error(format string, v ...interface{})
	Warn(format string, v ...interface{})
}

// stdLogger implements Logger with standard fmt.
type stdLogger struct{}

func (logger stdLogger) Error(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] EventEmitter > "+format+"\n", v...)
}

func (logger stdLogger) Warn(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "[WARN] EventEmitter > "+format+"\n", v...)
}
