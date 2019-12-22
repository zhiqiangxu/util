package wm

import (
	"context"
	"sync/atomic"

	"github.com/zhiqiangxu/rpheap"
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
	heap       *rpheap.Heap
}

// NewOffset is ctor for Offset
func NewOffset() *Offset {
	o := &Offset{waitCh: make(chan wait), doneCh: make(chan struct{}, 1), heap: rpheap.New()}
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

	waits := make(map[int64][]chan struct{})
	saveWait := func(w wait) {
		ws := waits[w.expectOffset]
		if ws == nil {
			o.heap.Insert(w.expectOffset)
			waits[w.expectOffset] = []chan struct{}{w.waiter}
		} else {
			waits[w.expectOffset] = append(ws, w.waiter)
		}

	}
	notifyUntil := func(doneOffset int64) {
		if o.heap.Size() == 0 {
			return
		}
		for {
			minOffset := o.heap.FindMin()
			if minOffset <= doneOffset {
				for _, w := range waits[minOffset] {
					close(w)
				}
				delete(waits, minOffset)
				o.heap.DeleteMin()
			} else {
				break
			}
		}
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
