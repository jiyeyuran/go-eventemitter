package eventemitter

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventEmitter_AddListener(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	emitter.On(evName, func() {})
	emitter.On(evName, func() {})
	assert.Equal(t, 2, emitter.ListenerCount(evName))
}

func TestEventEmitter_Once(t *testing.T) {
	evName := "test"
	emitter := NewEventEmitter()

	onceObserver := NewMockFunc(t)
	emitter.Once(evName, onceObserver.Fn())
	result := emitter.SafeEmit(evName)
	assert.Equal(t, 0, emitter.ListenerCount(evName))
	result.Wait()

	emitter.Once(evName, onceObserver.Fn())
	emitter.Emit(evName)
	assert.Equal(t, 0, emitter.ListenerCount(evName))

	emitter.Once(evName, onceObserver.Fn())
	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go (func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				emitter.SafeEmit(evName)
			} else {
				emitter.Emit(evName)
			}
		})(i)
	}

	wg.Wait()

	onceObserver.ExpectCalledTimes(1)
	assert.Equal(t, 0, emitter.ListenerCount(evName))
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
	emitter.On(evName, func(s struct1) {
		called++
	})
	// test unmarshal to *struct
	emitter.On(evName, func(s *struct1) {
		called++
	})
	// test unmarshal to json.RawMessage
	emitter.On(evName, func(s json.RawMessage) {
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
	emitter.On(evName2, func(s []struct1) {
		called++
	})
	// test unmarshal []byte
	emitter.On(evName2, func(s []byte) {
		called++
	})
	// test unmarshal json.RawMessage
	emitter.On(evName2, func(s json.RawMessage) {
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
	n := DefaultQueueSize
	results := []AysncResult{}

	emitter.On(evName, onObserver.Fn())
	emitter.Emit(evName)
	for i := 0; i < n; i++ {
		result := emitter.SafeEmit(evName)
		results = append(results, result)
	}
	emitter.RemoveAllListeners(evName)
	results = append(results, emitter.SafeEmit(evName))
	for _, result := range results {
		result.Wait()
	}
	onObserver.ExpectCalledTimes(n + 1)

	onObserver.Reset()
	emitter.RemoveAllListeners(evName)
	emitter.Emit(evName)
	emitter.SafeEmit(evName).Wait()
	onObserver.ExpectCalledTimes(0)

	results = []AysncResult{}
	emitter.On(evName, onObserver.Fn())
	emitter.Emit(evName)
	for i := 0; i < n; i++ {
		result := emitter.SafeEmit(evName)
		results = append(results, result)
	}
	emitter.RemoveAllListeners()
	results = append(results, emitter.SafeEmit(evName))
	for _, result := range results {
		result.Wait()
	}
	onObserver.ExpectCalledTimes(n + 1)

	assert.Equal(t, 0, emitter.ListenerCount(evName))
}
