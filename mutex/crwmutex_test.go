package mutex

import (
	"context"
	"testing"
	"time"
)

func TestCRWMutex(t *testing.T) {
	crwm := NewCRWMutex()
	// test RLock/RUnlock
	err := crwm.RLock(context.Background())
	if err != nil {
		t.FailNow()
	}

	crwm.RUnlock()

	// test Lock/Unlock
	err = crwm.Lock(context.Background())
	if err != nil {
		t.FailNow()
	}
	crwm.Unlock()

	// test RLock/canceled Lock/RUnlock
	err = crwm.RLock(context.Background())
	if err != nil {
		t.FailNow()
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancelFunc()
	err = crwm.Lock(ctx)
	if err == nil {
		t.FailNow()
	}

	crwm.RUnlock()

}

func TestSemaphoreBug(t *testing.T) {
	crwm := NewCRWMutex()
	// hold 1 read lock
	crwm.RLock(context.Background())

	go func() {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*300)
		defer cancelFunc()
		// start a Lock request that will giveup after 300ms
		err := crwm.Lock(ctx)
		if err == nil {
			t.FailNow()
		}
	}()

	// sleep 100ms, long enough for the Lock request to be queued
	time.Sleep(time.Millisecond * 100)
	// this channel will be closed if the following RLock succeeded
	doneCh := make(chan struct{})
	go func() {
		// try to grab a read lock, it will be queued after the Lock request
		// but should be notified when the Lock request is canceled
		// this doesn't happen because there's a bug in semaphore
		err := crwm.RLock(context.Background())
		if err != nil {
			t.FailNow()
		}
		crwm.RUnlock()
		close(doneCh)
	}()

	// because of the bug in semaphore, doneCh is never closed
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.FailNow()
	}

}
