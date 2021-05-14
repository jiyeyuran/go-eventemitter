package eventemitter

import "time"

// Option defines variadic parameter to create EventEmitter
type Option func(*EventEmitter)

// WithLogger set logger. By default, golang standard log will be used.
func WithLogger(logger Logger) Option {
	return func(ee *EventEmitter) {
		ee.logger = logger
	}
}

// WithDecoder set bytes decode. By default, bytes will be decoded by json decoder.
func WithDecoder(decoder Decoder) Option {
	return func(ee *EventEmitter) {
		ee.decoder = decoder
	}
}

// WithQueueSize set max queue buffer size for asynchronous event. By default, a maximum of 128 buffer size for any asynchronous event.
func WithQueueSize(queueSize int) Option {
	return func(ee *EventEmitter) {
		ee.queueSize = queueSize
	}
}

// WithMaxListeners set max listeners for single event. By default, a maximum of 10 listeners can be registered for any single event.
func WithMaxListeners(num int) Option {
	return func(ee *EventEmitter) {
		ee.maxListeners = num
	}
}

// WithPanicHandler set panic handler to catch panic from listener calling.
func WithPanicHandler(handler PanicHandler) Option {
	return func(ee *EventEmitter) {
		ee.panicHandler = handler
	}
}

// WithIdleLoopExitingDuration set the duration to exit loop which is idle.
func WithIdleLoopExitingDuration(d time.Duration) Option {
	return func(ee *EventEmitter) {
		ee.idleLoopExitingDur = d
	}
}
