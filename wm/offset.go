package wm

import (
	"context"
	"sync/atomic"
)

type wait struct {
	expectOffset int64
	waiter       chan struct{}
}

// Offset watermark model
type Offset struct {
	doneOffset int64
	waitCh     chan wait
	doneCh     chan struct{}
}

// NewOffset is ctor for Offset
func NewOffset() *Offset {
	o := &Offset{waitCh: make(chan wait), doneCh: make(chan struct{}, 1)}
	go o.process()
	return o
}

// Done for update doneOffset
func (o *Offset) Done(offset int64) {
	atomic.StoreInt64(&o.doneOffset, offset)
	select {
	case o.doneCh <- struct{}{}:
	default:
	}
}

// Wait for doneOffset>= expectOffset
func (o *Offset) Wait(ctx context.Context, expectOffset int64) error {
	if atomic.LoadInt64(&o.doneOffset) >= expectOffset {
		return nil
	}

	waiter := make(chan struct{})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case o.waitCh <- wait{expectOffset: expectOffset, waiter: waiter}:
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-waiter:
			return nil
		}
	}
}

func (o *Offset) process() {

	saveWait := func(w wait) {

	}
	notifyUntil := func(doneOffset int64) {

	}
	for {
		select {
		case w := <-o.waitCh:
			doneOffset := atomic.LoadInt64(&o.doneOffset)
			if w.expectOffset <= doneOffset {
				close(w.waiter)
			} else {
				saveWait(w)
			}
		case <-o.doneCh:
			doneOffset := atomic.LoadInt64(&o.doneOffset)

			// notify all waiters until doneOffset
			notifyUntil(doneOffset)
		}
	}
}
