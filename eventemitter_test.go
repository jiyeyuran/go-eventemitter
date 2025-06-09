package eventemitter

import (
	"encoding/json"
	"runtime"
	"testing"
	"time"
)

func TestEventEmitter_AddListener(t *testing.T) {
	evName := "test"
	emitter := New()

	emitter.On(evName, func() {})
	emitter.On(evName, func() {})
	emitter.Once(evName, func() {})

	if emitter.ListenerCount(evName) != 3 {
		t.Errorf("listener count is %d, expected 3", emitter.ListenerCount(evName))
	}
}

func TestEventEmitter_Once(t *testing.T) {
	evName := "test"
	emitter := New()

	count := 0
	emitter.Once(evName, func() { count++ })
	if err := emitter.AsyncEmit(evName).Wait(); err != nil {
		t.Error(err)
	}
	if err := emitter.AsyncEmit(evName).Wait(); err != nil {
		t.Error(err)
	}
	if emitter.ListenerCount(evName) != 0 {
		t.Errorf("listener count is %d, expected 0", emitter.ListenerCount(evName))
	}

	if count != 1 {
		t.Errorf("onceObserver is called %d times, expected 1", count)
	}

	count = 0
	emitter.Once(evName, func() { count++ })
	emitter.Emit(evName)
	emitter.Emit(evName)

	if emitter.ListenerCount(evName) != 0 {
		t.Errorf("listener count is %d, expected 0", emitter.ListenerCount(evName))
	}
	if count != 1 {
		t.Errorf("onceObserver is called %d times, expected 1", count)
	}
}

func TestEventEmitter_AlignArguments(t *testing.T) {
	evName := "test"
	emitter := New()

	count := 0
	emitter.On(evName, func() { count++ })
	emitter.On(evName, func(i, j int) {})
	emitter.Emit(evName)
	emitter.Emit(evName, 1)
	emitter.Emit(evName, 1, 2)
	emitter.Emit(evName, 1, 2, 3)

	if count != 4 {
		t.Errorf("observer is called %d times, expected 4", count)
	}
}

func TestEventEmitter_UnmarshalArguments(t *testing.T) {
	emitter := New()

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

	if called != 3*2 {
		t.Errorf("observer is called %d times, expected 6", called)
	}

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

	if called != 3*2 {
		t.Errorf("observer is called %d times, expected 6", called)
	}
}

func TestEventEmitter_ReceiveArgumentsAreTheSameAsEmiting(t *testing.T) {
	emitter := New()
	evName := "test"
	var i, j int
	emitter.On(evName, func(arg1, arg2 int) {
		i, j = arg1, arg2
	})
	emitter.Emit(evName, 1, 2)
	if i != 1 || j != 2 {
		t.Errorf("observer is called with %d, %d, expected 1, 2", i, j)
	}
}

func TestEventEmitter_AsyncEmit(t *testing.T) {
	evName := "test"
	emitter := New()

	count := 0
	emitter.On(evName, func() { count++ })
	if err := emitter.AsyncEmit(evName).Wait(); err != nil {
		t.Errorf("emit error: %v", err)
	}
	if count != 1 {
		t.Errorf("observer is called %d times, expected 1", count)
	}
}

func TestEventEmitter_RemoveListener(t *testing.T) {
	evName := "test"
	emitter := New()

	fn := func() {}
	emitter.On(evName, fn)
	emitter.Off(evName, fn)

	if emitter.ListenerCount(evName) != 0 {
		t.Errorf("listener count is %d, expected 0", emitter.ListenerCount(evName))
	}
}

func TestEventEmitter_RemoveAllListeners(t *testing.T) {
	evName := "test"
	emitter := New()

	fn := func() {}

	emitter.On(evName, fn)
	emitter.On(evName, fn)
	emitter.RemoveAllListeners(evName)

	if emitter.ListenerCount(evName) != 0 {
		t.Errorf("listener count is %d, expected 0", emitter.ListenerCount(evName))
	}

	emitter.On(evName, fn)
	emitter.On("abc", fn)
	emitter.RemoveAllListeners()
	if emitter.ListenerCount(evName) != 0 || emitter.ListenerCount("abc") != 0 {
		t.Errorf("listener count is %d, expected 0", emitter.ListenerCount(evName))
	}
}

func TestEventEmitter_LoopExitedAfterIdleDuration(t *testing.T) {
	evName := "test"
	duration := time.Millisecond * 10
	emitter := New(WithIdleLoopExitingDuration(duration))

	goroutines := runtime.NumGoroutine()

	emitter.On(evName, func() {})

	if err := emitter.AsyncEmit(evName).Wait(); err != nil {
		t.Errorf("emit error: %v", err)
	}

	if runtime.NumGoroutine() != goroutines+1 {
		t.Errorf("goroutine count is %d, expected %d", runtime.NumGoroutine(), goroutines+1)
	}

	time.Sleep(duration + time.Millisecond*10)

	if runtime.NumGoroutine() != goroutines {
		t.Errorf("goroutine count is %d, expected %d", runtime.NumGoroutine(), goroutines)
	}
}
