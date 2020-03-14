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
	crwm.RLock(context.Background())
	crwm.RLock(context.Background())

	go func() {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*300)
		defer cancelFunc()
		err := crwm.Lock(ctx)
		if err == nil {
			t.FailNow()
		}
	}()

	time.Sleep(time.Millisecond * 100)
	doneCh := make(chan struct{})
	go func() {
		err := crwm.RLock(context.Background())
		if err != nil {
			t.FailNow()
		}
		crwm.RUnlock()
		close(doneCh)
	}()

	crwm.RUnlock()

	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.FailNow()
	}

}
