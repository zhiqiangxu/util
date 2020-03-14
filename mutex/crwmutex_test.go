package mutex

import (
	"context"
	"testing"
	"time"
)

func TestCRWMutex(t *testing.T) {
	crwm := NewCRWMutex()
	err := crwm.RLock(context.Background())
	if err != nil {
		t.FailNow()
	}

	crwm.RUnlock()

	err = crwm.Lock(context.Background())
	if err != nil {
		t.FailNow()
	}
	crwm.Unlock()

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
