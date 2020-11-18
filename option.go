package eventemitter

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
