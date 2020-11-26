package eventemitter

import (
	"context"
	"sync"
)

// Waiting all aysnc listeners to be done.
type AysncResult interface {
	Wait()
	WaitCtx(ctx context.Context) error
}

type AysncResultImpl struct {
	wg *sync.WaitGroup
}

func NewAysncResultImpl(wg *sync.WaitGroup) AysncResult {
	return &AysncResultImpl{
		wg: wg,
	}
}

func (s *AysncResultImpl) Wait() {
	s.wg.Wait()
}

func (s *AysncResultImpl) WaitCtx(ctx context.Context) error {
	doneCh := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
