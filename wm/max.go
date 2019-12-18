package wm

import (
	"context"
	"sync/atomic"
	"time"
)

// Max watermark model
type Max struct {
	count int64
	max   int64
	ch    chan struct{}
}

// NewMax is ctor for Max
func NewMax(max int64) *Max {
	m := &Max{max: max, ch: make(chan struct{})}
	go m.monitor()
	return m
}

// Enter returns nil if not canceled
func (m *Max) Enter(ctx context.Context) error {
	count := atomic.AddInt64(&m.count, 1)

	if count > m.max {
		select {
		case m.ch <- struct{}{}:
			return nil
		case <-ctx.Done():
			m.exit()
			return ctx.Err()
		}
	}

	return nil
}

// Exit should only be called if Enter returns nil
func (m *Max) Exit() {
	count := m.exit()

	if count >= m.max {
		select {
		case <-m.ch:
		default:
		}
	}

}

func (m *Max) exit() int64 {
	count := atomic.AddInt64(&m.count, -1)
	if count < 0 {
		panic("Max: count < 0")
	}
	return count
}

func (m *Max) monitor() {
	for {
		if atomic.LoadInt64(&m.count) < m.max {
			select {
			case <-m.ch:
				continue
			default:
			}
		}

		time.Sleep(time.Second)
	}
}
