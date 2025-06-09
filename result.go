package eventemitter

import (
	"context"
	"sync"
)

// Waiting all aysnc listeners to be done.
type AysncResult interface {
	Wait() error
	WaitContext(ctx context.Context) error
}

type aysncResultImpl struct {
	wg  *sync.WaitGroup
	err error
}

func newAysncResultImpl(wg *sync.WaitGroup) *aysncResultImpl {
	return &aysncResultImpl{
		wg: wg,
	}
}

func (s *aysncResultImpl) Wait() error {
	s.wg.Wait()
	return s.err
}

func (s *aysncResultImpl) WaitContext(ctx context.Context) error {
	doneCh := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}
