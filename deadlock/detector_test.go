package deadlock

import (
	"testing"
	"time"
)

func TestDetector(t *testing.T) {
	var (
		lock1, lock2 Mutex
	)
	lock1.Init()
	lock2.Init()

	var (
		errDL    *ErrorDeadlock
		errUsage *ErrorUsage
	)

	go func() {
		lock1.Lock()

		time.Sleep(time.Millisecond * 200)

		defer func() {
			panicErr := recover()
			errDL, errUsage = ParsePanicError(panicErr)
		}()

		lock2.Lock()

		select {}
	}()

	time.Sleep(time.Millisecond * 100)

	go func() {
		lock2.Lock()

		lock1.Lock()
	}()

	time.Sleep(time.Millisecond * 300)
	if errDL == nil {
		t.FailNow()
	}
	if errUsage != nil {
		t.FailNow()
	}

}

func TestMap(t *testing.T) {
	m := make(map[int]int)

	m[1]++

	if m[1] != 1 {
		t.FailNow()
	}

	m[1]--
	m[1]--

	if m[1] != -1 {
		t.FailNow()
	}

	var nilMap map[int]int
	if _, ok := nilMap[1]; ok {
		t.FailNow()
	}
}
