package eventemitter

import "log"

type Logger interface {
	Error(format string, v ...interface{})
}

type stdLogger struct{}

func (logger stdLogger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] EventEmitter > "+format, v...)
}
