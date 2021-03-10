package eventemitter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type MockFunc struct {
	require *require.Assertions
	called  int
	args    []interface{}
}

func NewMockFunc(t *testing.T) *MockFunc {
	return &MockFunc{
		require: require.New(t),
	}
}

func (w *MockFunc) Fn() func(...interface{}) {
	w.Reset()

	return func(args ...interface{}) {
		w.args = args
		w.called++
	}
}

func (w *MockFunc) ExpectCalledWith(args ...interface{}) {
	for i, arg := range args {
		w.require.EqualValues(arg, w.args[i])
	}
}

func (w *MockFunc) ExpectCalled() {
	w.require.NotZero(w.called)
}

func (w *MockFunc) ExpectCalledTimes(called int) {
	w.require.Equal(called, w.called)
}

func (w *MockFunc) CalledTimes() int {
	return int(w.called)
}

func (w *MockFunc) Reset() {
	w.args = nil
	w.called = 0
}
