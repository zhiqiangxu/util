package diskqueue

import (
	"sync"
	"time"
)

var (
	timeLock sync.Mutex
	lastTime int64
)

// NowNano returns monotonic nano time
func NowNano() int64 {
	timeLock.Lock()

	thisTime := time.Now().UnixNano()

	if thisTime <= lastTime {
		thisTime = lastTime + 1
	}

	lastTime = thisTime

	timeLock.Unlock()

	return thisTime
}
