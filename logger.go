package eventemitter

import "log"

// Logger defines interface to print log information.
type Logger interface {
	Error(format string, v ...interface{})
	Warn(format string, v ...interface{})
}

// stdLogger implements Logger with standard log.
type stdLogger struct{}

func (logger stdLogger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] EventEmitter > "+format, v...)
}

func (logger stdLogger) Warn(format string, v ...interface{}) {
	log.Printf("[WARN] EventEmitter > "+format, v...)
}
