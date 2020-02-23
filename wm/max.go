package wm

import (
	"context"

	"golang.org/x/sync/semaphore"
)

// Max watermark model
type Max struct {
	sema *semaphore.Weighted
}

// NewMax is ctor for Max
func NewMax(max int64) *Max {
	m := &Max{sema: semaphore.NewWeighted(max)}
	return m
}

// Enter returns nil if not canceled
func (m *Max) Enter(ctx context.Context) error {
	return m.sema.Acquire(ctx, 1)
}

// TryEnter returns true if succeed
func (m *Max) TryEnter() bool {
	return m.sema.TryAcquire(1)
}

// Exit should only be called if Enter returns nil
func (m *Max) Exit() {
	m.sema.Release(1)
}
