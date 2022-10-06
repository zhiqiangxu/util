package parallel

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	zlog "github.com/rs/zerolog/log"
)

type Result struct {
	R interface{}
	I int
}

func First[R any](ctx context.Context, workers int, handleFunc func(ctx context.Context, workerIdx int) (R, error), cd time.Duration) (r R, i int, err error) {
	if workers == 0 {
		err = fmt.Errorf("workers == 0")
		return
	}

	if workers == 1 {
		for {
			r, err = handleFunc(ctx, 0)
			if err == nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
			time.Sleep(cd)
		}
	}

	resultCh := make(chan Result, 1)
	doneCh := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(i int) {
			var (
				r   R
				err error
			)
			for {
				r, err = handleFunc(ctx, i)
				if err != nil {
					select {
					case <-doneCh:
						return
					case <-ctx.Done():
						return
					default:
					}
					time.Sleep(cd)
					continue
				}

				select {
				case resultCh <- Result{R: r, I: i}:
				default:
				}
				return
			}

		}(i)
	}

	select {
	case result := <-resultCh:
		r = result.R.(R)
		i = result.I
		close(doneCh)
	case <-ctx.Done():
		err = ctx.Err()
	}

	return
}

type subTask struct {
	from, to int
}

func All(ctx context.Context, total, unit, workers int, handleFunc func(ctx context.Context, workerIdx int, from int, to int) error, dispatchCB func(int, int), retry int, cd time.Duration) {
	if workers == 0 {
		panic("workers == 0")
	}

	taskChan := make(chan subTask, workers)
	doneCh := make(chan struct{})

	var (
		wg        sync.WaitGroup
		doneCount int64
	)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			var (
				err  error
				task subTask
			)
			for {
				select {
				case task = <-taskChan:
					for k := 0; k < retry; k++ {
						err = handleFunc(ctx, i, task.from, task.to)
						if err != nil {
							if k == retry-1 {
								break
							}
							zlog.Warn().Int("i", i).Err(err).Msg("handleFunc")
							time.Sleep(cd)
							continue
						}
						if atomic.AddInt64(&doneCount, int64(task.to-task.from)) == int64(total) {
							close(doneCh)
							return
						}
						break
					}
					if err != nil {
						// switch a worker and try again
						select {
						case taskChan <- task:
						case <-ctx.Done():
							return
						}
						time.Sleep(cd)
					}

				case <-ctx.Done():
					return
				case <-doneCh:
					return
				}
			}
		}(i)
	}

	for from := 0; from < total; from += unit {
		to := from + unit
		if to > total {
			to = total
		}
		if dispatchCB != nil {
			dispatchCB(from, to)
		}
		select {
		case taskChan <- subTask{from: from, to: to}:
		case <-ctx.Done():
			return
		}
	}

	select {
	case <-doneCh:
	case <-ctx.Done():
	}

	wg.Wait()
	return
}
