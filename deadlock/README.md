# deadlock

This package implements an easy to prove deadlock detector, just replace `sync.Mutex` and `sync.RWMutex` with `deadlock.NewMutex()` and `deadlock.NewRWMutex()`.