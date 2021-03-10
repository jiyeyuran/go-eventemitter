package eventemitter

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
)

var DefaultQueueSize = 128

// IEventEmitter defines event emitter interface
type IEventEmitter interface {
	// AddListener is the alias for emitter.On(eventName, listener).
	AddListener(evt string, listener interface{}) IEventEmitter

	// Once adds a one-time listener function for the event named eventName.
	// The next time eventName is triggered, this listener is removed and then invoked.
	Once(evt string, listener interface{}) IEventEmitter

	// Emit synchronously calls each of the listeners registered for the event named eventName,
	// in the order they were registered, passing the supplied arguments to each.
	// Returns true if the event had listeners, false otherwise.
	Emit(evt string, argv ...interface{}) bool

	// SafeEmit asynchronously calls each of the listeners registered for the event named eventName.
	// By default, a maximum of 128 events can be buffered.
	// Panic will be catched and logged as error.
	// Returns AysncResult.
	SafeEmit(evt string, argv ...interface{}) AysncResult

	// RemoveListener is the alias for emitter.Off(eventName, listener).
	RemoveListener(evt string, listener interface{}) IEventEmitter

	// RemoveAllListeners removes all listeners, or those of the specified eventNames.
	RemoveAllListeners(evts ...string) IEventEmitter

	// On adds the listener function to the end of the listeners array for the event named eventName.
	// No checks are made to see if the listener has already been added.
	// Multiple calls passing the same combination of eventName and listener will result in the listener
	// being added, and called, multiple times.
	// By default, a maximum of 10 listeners can be registered for any single event.
	// This is a useful default that helps finding memory leaks. Note that this is not a hard limit.
	// The EventEmitter instance will allow more listeners to be added but will output a trace warning
	// to log indicating that a "possible EventEmitter memory leak" has been detected.
	On(evt string, listener interface{}) IEventEmitter

	// Off removes the specified listener from the listener array for the event named eventName.
	Off(evt string, listener interface{}) IEventEmitter

	// ListenerCount returns the number of listeners listening to the event named eventName.
	ListenerCount(evt string) int
}

type intervalListener struct {
	Once          *sync.Once
	evt           string
	listenerValue reflect.Value
	argTypes      []reflect.Type
	decoder       Decoder
}

type callWrapper struct {
	wg       *sync.WaitGroup
	listener *intervalListener
	values   []reflect.Value
}

func newInternalListener(evt string, listener interface{}, once bool, decoder Decoder) *intervalListener {
	var argTypes []reflect.Type
	listenerValue := reflect.ValueOf(listener)
	listenerType := listenerValue.Type()

	for i := 0; i < listenerType.NumIn(); i++ {
		argTypes = append(argTypes, listenerType.In(i))
	}

	l := &intervalListener{
		evt:           evt,
		argTypes:      argTypes,
		listenerValue: listenerValue,
		decoder:       decoder,
	}

	if once {
		l.Once = &sync.Once{}
	}

	return l
}

func (l *intervalListener) Event() string {
	return l.evt
}

func (l *intervalListener) Call(callArgs []reflect.Value) {
	if !l.listenerValue.Type().IsVariadic() {
		callArgs = l.alignArguments(callArgs)
	}
	callArgs = l.convertArguments(callArgs)
	l.listenerValue.Call(callArgs)
}

func (l *intervalListener) CallWrapper(wg *sync.WaitGroup, callArgs []reflect.Value) callWrapper {
	return callWrapper{
		wg:       wg,
		listener: l,
		values:   callArgs,
	}
}

func (l intervalListener) convertArguments(args []reflect.Value) []reflect.Value {
	if len(args) != len(l.argTypes) {
		return args
	}
	actualArgs := make([]reflect.Value, len(args))

	for i, arg := range args {
		// Unmarshal bytes to golang type
		if isBytesType(arg.Type()) && !isBytesType(l.argTypes[i]) {
			val := reflect.New(l.argTypes[i]).Interface()
			if err := l.decoder.Decode(arg.Bytes(), val); err == nil {
				actualArgs[i] = reflect.ValueOf(val).Elem()
			}
		} else if arg.Type() != l.argTypes[i] &&
			arg.Type().ConvertibleTo(l.argTypes[i]) {
			actualArgs[i] = arg.Convert(l.argTypes[i])
		} else {
			actualArgs[i] = arg
		}
	}

	return actualArgs
}

func (l intervalListener) alignArguments(args []reflect.Value) (actualArgs []reflect.Value) {
	// delete unwanted arguments
	if argLen := len(l.argTypes); len(args) >= argLen {
		actualArgs = args[0:argLen]
	} else {
		actualArgs = args[:]

		// append missing arguments with zero value
		for _, argType := range l.argTypes[len(args):] {
			actualArgs = append(actualArgs, reflect.Zero(argType))
		}
	}

	return actualArgs
}

// The EventEmitter implements IEventEmitter
type EventEmitter struct {
	mu            sync.Mutex
	logger        Logger
	decoder       Decoder
	queueSize     int
	maxListeners  int
	callWrapperCh chan callWrapper
	evtListeners  map[string][]*intervalListener
	loopStarted   bool
}

func NewEventEmitter(options ...Option) IEventEmitter {
	ee := &EventEmitter{
		logger:       stdLogger{},
		decoder:      JsonDecoder{},
		queueSize:    DefaultQueueSize,
		maxListeners: 10,
		evtListeners: make(map[string][]*intervalListener),
	}

	for _, option := range options {
		option(ee)
	}

	ee.callWrapperCh = make(chan callWrapper, ee.queueSize)

	return ee
}

func (e *EventEmitter) AddListener(evt string, listener interface{}) IEventEmitter {
	return e.On(evt, listener)
}

// Emit fires a particular event
func (e *EventEmitter) Emit(evt string, args ...interface{}) bool {
	onceListeners := []*intervalListener{}
	defer func() {
		for _, listener := range onceListeners {
			e.RemoveListener(evt, listener.listenerValue.Interface())
		}
	}()

	e.mu.Lock()
	listeners := e.evtListeners[evt][:]
	e.mu.Unlock()

	callArgs := make([]reflect.Value, 0, len(args))

	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}

	for _, listener := range listeners {
		if listener.Once != nil {
			listener.Once.Do(func() {
				listener.Call(callArgs)
				onceListeners = append(onceListeners, listener)
			})
		} else {
			listener.Call(callArgs)
		}
	}

	return len(listeners) > 0
}

// SafaEmit fires a particular event asynchronously.
func (e *EventEmitter) SafeEmit(evt string, args ...interface{}) AysncResult {
	onceListeners := []*intervalListener{}
	defer func() {
		for _, listener := range onceListeners {
			e.RemoveListener(evt, listener.listenerValue.Interface())
		}
	}()

	e.mu.Lock()
	defer e.mu.Unlock()

	listeners := e.evtListeners[evt][:]

	wg := &sync.WaitGroup{}
	wg.Add(len(listeners))

	callArgs := make([]reflect.Value, 0, len(args))

	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}
	for _, listener := range listeners {
		e.asyncCall(wg, listener, callArgs)
		if listener.Once != nil {
			onceListeners = append(onceListeners, listener)
		}
	}

	return NewAysncResultImpl(wg)
}

func (e *EventEmitter) RemoveListener(evt string, listener interface{}) IEventEmitter {
	return e.Off(evt, listener)
}

func (e *EventEmitter) RemoveAllListeners(evts ...string) IEventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(evts) == 0 {
		e.evtListeners = make(map[string][]*intervalListener)
	} else {
		for _, evt := range evts {
			delete(e.evtListeners, evt)
		}
	}

	if len(e.evtListeners) == 0 {
		e.stopLoop()
	}

	return e
}

func (e *EventEmitter) Once(evt string, listener interface{}) IEventEmitter {
	if err := isValidListener(listener); err != nil {
		panic(err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.evtListeners == nil {
		e.evtListeners = make(map[string][]*intervalListener)
	}
	e.evtListeners[evt] = append(e.evtListeners[evt], newInternalListener(evt, listener, true, e.decoder))
	e.startLoop()

	return e
}

func (e *EventEmitter) On(evt string, listener interface{}) IEventEmitter {
	if err := isValidListener(listener); err != nil {
		panic(err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.maxListeners > 0 && len(e.evtListeners[evt]) >= e.maxListeners {
		e.logger.Warn(`AddListener | max listeners (%d) for event: "%s" are reached!`, e.maxListeners, evt)
	}
	internalListener := newInternalListener(evt, listener, false, e.decoder)
	e.evtListeners[evt] = append(e.evtListeners[evt], internalListener)

	e.startLoop()

	return e
}

func (e *EventEmitter) Off(evt string, listener interface{}) IEventEmitter {
	if err := isValidListener(listener); err != nil {
		panic(err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	idx := -1
	pointer := reflect.ValueOf(listener).Pointer()
	listeners := e.evtListeners[evt]

	for index, item := range listeners {
		if item.listenerValue.Pointer() == pointer {
			idx = index
			break
		}
	}

	if idx < 0 {
		return e
	}

	e.evtListeners[evt] = append(listeners[:idx], listeners[idx+1:]...)

	if len(e.evtListeners[evt]) == 0 {
		e.stopLoop()
	}

	return e
}

func (e *EventEmitter) ListenerCount(evt string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.evtListeners == nil {
		return 0
	}

	return len(e.evtListeners[evt])
}

func (e *EventEmitter) Len() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return len(e.evtListeners)
}

func (e *EventEmitter) asyncCall(wg *sync.WaitGroup, listener *intervalListener, callArgs []reflect.Value) {
	call := func() {
		if !e.loopStarted {
			return
		}
		e.callWrapperCh <- listener.CallWrapper(wg, callArgs)
	}
	if listener.Once != nil {
		listener.Once.Do(call)
	} else {
		call()
	}
}

func (e *EventEmitter) startLoop() {
	if !e.loopStarted {
		e.loopStarted = true
		e.callWrapperCh = make(chan callWrapper, e.queueSize)
		go e.runLoop(e.callWrapperCh)
	}
}

func (e *EventEmitter) stopLoop() {
	if e.loopStarted {
		e.loopStarted = false
		close(e.callWrapperCh)
	}
}

func (e *EventEmitter) runLoop(callWrapperCh chan callWrapper) {
	call := func(argument callWrapper) {
		defer func() {
			argument.wg.Done()

			if r := recover(); r != nil {
				e.logger.Error("SafeEmit() | event listener threw an error [event:%s]: %s", argument.listener.Event(), r)
				debug.PrintStack()
			}
		}()
		argument.listener.Call(argument.values)
	}

	for callWrapper := range callWrapperCh {
		call(callWrapper)
	}
}

func isValidListener(fn interface{}) error {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(fn))
	}

	return nil
}

func isBytesType(tp reflect.Type) bool {
	return tp.Kind() == reflect.Slice && tp.Elem().Kind() == reflect.Uint8
}
