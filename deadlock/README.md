# deadlock

This package implements a deadlock detector with easy to prove correctness, just replace `sync.Mutex` and `sync.RWMutex` with `deadlock.NewMutex()` and `deadlock.NewRWMutex()` and you're ready to go!!!

When something wrong happend, panic will happen.

Just catch the `panic` and handle it with `ParsePanicError`, like following:

```golang
defer func() {
    panicErr := recover()
    errDeadlock, errUsage := ParsePanicError(panicErr)
}()

```

If everything went well, both `errDeadlock` and `errUsage` should be nil.

If deadlock happend, `errDeadlock` is non nil.

If usage problem happens, like the same goroutine calls `Locks` on the same `Mutex` multiple times, `errUsage` is non nil.