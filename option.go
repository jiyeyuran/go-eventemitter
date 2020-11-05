package eventemitter

type Option func(*EventEmitter)

func WithLogger(logger Logger) Option {
	return func(ee *EventEmitter) {
		ee.logger = logger
	}
}

func WithDecoder(decoder Decoder) Option {
	return func(ee *EventEmitter) {
		ee.decoder = decoder
	}
}

func WithQueueSize(queueSize int) Option {
	return func(ee *EventEmitter) {
		ee.queueSize = queueSize
	}
}
