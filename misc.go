package util

import (
	"sync"
	"time"

	"github.com/zhiqiangxu/util/logger"
	"go.uber.org/zap"
)

// GoFunc runs a goroutine under WaitGroup
func GoFunc(routinesGroup *sync.WaitGroup, f func()) {
	routinesGroup.Add(1)
	go func() {
		defer routinesGroup.Done()
		f()
	}()
}

// TryUntilSuccess will try f until success
func TryUntilSuccess(f func() bool, duration time.Duration) {
	var r bool
	for {
		r = f()
		if r {
			return
		}
		time.Sleep(duration)
	}
}

// RunWithRecovery for run exec, calls recoverFn when panic
func RunWithRecovery(exec func(), recoverFn func(r interface{})) {
	defer func() {
		r := recover()
		if r != nil {
			logger.Instance().Error("panic in the recoverable goroutine",
				zap.Reflect("r", r),
				zap.Stack("stack trace"))

			if recoverFn != nil {
				recoverFn(r)
			}
		}
	}()
	exec()
}

// RunWithRetry will run the f with backoff and retry.
// retryCnt: Max retry count
// backoff: When run f failed, it will sleep backoff * triedCount time.Millisecond.
// Function f should have two return value. The first one is an bool which indicate if the err if retryable.
// The second is if the f meet any error.
func RunWithRetry(retryCnt int, backoff uint64, f func() (bool, error)) (err error) {
	var retryAble bool
	for i := 1; i <= retryCnt; i++ {
		retryAble, err = f()
		if err == nil || !retryAble {
			return
		}
		sleepTime := time.Duration(backoff*uint64(i)) * time.Millisecond
		time.Sleep(sleepTime)
	}
	return
}
