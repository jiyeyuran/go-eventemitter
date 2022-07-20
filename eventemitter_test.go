package eventemitter

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventEmitter_AddListener(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	emitter.On(evName, func() {})
	emitter.On(evName, func() {})
	emitter.Once(evName, func() {})
	assert.Equal(t, 3, emitter.ListenerCount(evName))
}

func TestEventEmitter_Once(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	onceObserver := NewMockFunc(t)
	emitter.Once(evName, onceObserver.Fn())
	emitter.SafeEmit(evName).Wait()
	assert.Equal(t, 0, emitter.ListenerCount(evName))

	emitter.Once(evName, onceObserver.Fn())
	emitter.Emit(evName)
	assert.Equal(t, 0, emitter.ListenerCount(evName))

	emitter.Once(evName, onceObserver.Fn())
	wg := sync.WaitGroup{}
	// Emit
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			emitter.Emit(evName)
		}()
	}
	wg.Wait()
	onceObserver.ExpectCalledTimes(1)

	onceObserver = NewMockFunc(t)

	emitter.Once(evName, onceObserver.Fn())
	wg = sync.WaitGroup{}
	// SafeEmit
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			emitter.SafeEmit(evName).Wait()
		}()
	}
	wg.Wait()
	onceObserver.ExpectCalledTimes(1)
	assert.Equal(t, 0, emitter.ListenerCount(evName))

	evName1 := "test1"
	evName2 := "test2"
	evName1Observer := NewMockFunc(t)
	evName2Observer := NewMockFunc(t)
	emitter.Once(evName1, evName1Observer.Fn())
	emitter.Once(evName2, evName2Observer.Fn())

	emitter.SafeEmit(evName1)
	emitter.SafeEmit(evName2)

	evName1Observer.ExpectCalledTimes(1)
	evName2Observer.ExpectCalledTimes(1)
}

func TestEventEmitter_AlignArguments(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	onObserver := NewMockFunc(t)
	emitter.On(evName, onObserver.Fn())
	emitter.On(evName, func(i, j int) {})
	emitter.Emit(evName)
	emitter.Emit(evName, 1)
	emitter.Emit(evName, 1, 2)
	emitter.Emit(evName, 1, 2, 3)

	onObserver.ExpectCalledTimes(4)
}

func TestEventEmitter_UnmarshalArguments(t *testing.T) {
	emitter := NewEventEmitter()

	type struct1 struct {
		A int
	}

	s := struct1{A: 1}
	data, _ := json.Marshal(s)
	evName := "test"
	called := 0

	// test unmarshal to struct
	emitter.On(evName, func(_ struct1) {
		called++
	})
	// test unmarshal to *struct
	emitter.On(evName, func(_ *struct1) {
		called++
	})
	// test unmarshal to json.RawMessage
	emitter.On(evName, func(_ json.RawMessage) {
		called++
	})
	emitter.Emit(evName, data)
	emitter.Emit(evName, json.RawMessage(data))

	assert.Equal(t, 3*2, called)

	called = 0
	evName2 := "test2"
	ss := []struct1{{A: 1}}
	data, _ = json.Marshal(ss)

	// test unmarshal []struct
	emitter.On(evName2, func(_ []struct1) {
		called++
	})
	// test unmarshal []byte
	emitter.On(evName2, func(_ []byte) {
		called++
	})
	// test unmarshal json.RawMessage
	emitter.On(evName2, func(_ json.RawMessage) {
		called++
	})
	emitter.Emit(evName2, data)
	emitter.Emit(evName2, json.RawMessage(data))

	assert.Equal(t, 3*2, called)
}

func TestEventEmitter_ReceiveArgumentsAreTheSameAsEmiting(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()
	observer := NewMockFunc(t)

	emitter.On(evName, observer.Fn())
	emitter.Emit(evName, 1, 2)
	observer.ExpectCalledWith(1, 2)
}

func TestEventEmitter_SafeEmit(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()
	onObserver := NewMockFunc(t)

	emitter.On(evName, onObserver.Fn())
	emitter.SafeEmit(evName).Wait()
	onObserver.ExpectCalled()
}

func TestEventEmitter_RemoveListener(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	onObserver := NewMockFunc(t)
	fn := onObserver.Fn()

	emitter.On(evName, fn)
	emitter.Off(evName, fn)

	onObserver.ExpectCalledTimes(0)

	assert.Equal(t, 0, emitter.ListenerCount(evName))
}

func TestEventEmitter_RemoveAllListeners(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	onObserver := NewMockFunc(t)
	n := 10

	emitter.On(evName, onObserver.Fn())
	emitter.Emit(evName)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			emitter.SafeEmit(evName, index).Wait()
		}(i)
	}
	wg.Wait()
	emitter.RemoveAllListeners(evName)
	emitter.SafeEmit(evName).Wait()
	onObserver.ExpectCalledTimes(n + 1)

	onObserver.Reset()
	emitter.RemoveAllListeners(evName)
	emitter.Emit(evName)
	emitter.SafeEmit(evName).Wait()
	onObserver.ExpectCalledTimes(0)

	emitter.On(evName, onObserver.Fn())
	emitter.Emit(evName)
	wg = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			emitter.SafeEmit(evName).Wait()
		}()
	}
	wg.Wait()
	emitter.RemoveAllListeners()
	emitter.SafeEmit(evName).Wait()
	onObserver.ExpectCalledTimes(n + 1)

	assert.Equal(t, 0, emitter.ListenerCount(evName))
}

func TestEventEmitter_LoopExitedAfterIdleDuration(t *testing.T) {
	evName := "test"
	duration := time.Millisecond * 10
	sleepDur := duration + time.Millisecond
	emitter := NewEventEmitter(WithIdleLoopExitingDuration(duration)).(*EventEmitter)

	onObserver := NewMockFunc(t)

	emitter.On(evName, onObserver.Fn())
	emitter.SafeEmit(evName).Wait()

	assert.EqualValues(t, 1, emitter.loopStarted)
	time.Sleep(sleepDur)
	assert.EqualValues(t, 0, emitter.loopStarted)
	emitter.SafeEmit(evName).Wait()
	assert.EqualValues(t, 1, emitter.loopStarted)
}
