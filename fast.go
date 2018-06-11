package fast

import (
	"runtime"
	"sync"
)

// Call calls each fn in a separate goroutine, waits for completion of all
// calls, and returns any non-nil error. It is undefined which error is returned
// if multiple calls fail.
func Call(fn ...func() error) error {
	var ctx callCtx
	ctx.Add(len(fn))
	for _, f := range fn {
		go func(ctx *callCtx, f func() error) {
			ctx.done(f())
		}(&ctx, f)
	}
	ctx.Wait()
	return ctx.err
}

// ForEachIO executes n IO-bound tasks using up-to 64 worker goroutines.
func ForEachIO(n int, fn func(i int) error) error {
	return ForEach(n, 0, fn)
}

// ForEachCPU executes n CPU-bound tasks using up-to NumCPU() worker goroutines.
func ForEachCPU(n int, fn func(i int) error) error {
	return ForEach(n, runtime.NumCPU(), fn)
}

// ForEach executes n tasks using at most batch goroutines. If batch is <= 0,
// 64 goroutines are used (more may be used if n is close to 64). Function fn is
// called for each task with i in the range [0,n). If fn returns an error, all
// pending tasks are canceled and the error is returned. It is undefined which
// error is returned if multiple concurrent tasks fail.
func ForEach(n, batch int, fn func(i int) error) error {
	if n <= 1 || batch == 1 {
		for i := 0; i < n; i++ {
			if err := fn(i); err != nil {
				return err
			}
		}
		return nil
	}
	if batch <= 0 {
		if n < 96 {
			batch = n
		} else {
			batch = 64
		}
	}

	// Avoid channels if there is only one batch
	if n <= batch {
		var ctx callCtx
		ctx.Add(n)
		for i := 0; i < n; i++ {
			go func(ctx *callCtx, fn func(i int) error, i int) {
				ctx.done(fn(i))
			}(&ctx, fn, i)
		}
		ctx.Wait()
		return ctx.err
	}

	// Start waiter goroutine
	var wg sync.WaitGroup
	ich := make(chan int)
	ech := make(chan error)
	wg.Add(batch)
	go func() {
		defer close(ech)
		wg.Wait()
	}()

	// Start worker goroutines for the first batch
	for i := 0; i < batch; i++ {
		go func(i int) {
			defer wg.Done()
			if err := fn(i); err != nil {
				ech <- err
			} else {
				for i = range ich {
					if err = fn(i); err != nil {
						ech <- err
						break
					}
				}
			}
		}(i)
	}

	// Send remaining tasks while waiting for error
	var err error
	for i := batch; i < n; i++ {
		select {
		case ich <- i:
			continue
		case err = <-ech:
		}
		break
	}
	close(ich)

	// Wait for completion of all tasks
	for err = range ech {
	}
	return err
}

// callCtx synchronizes function calls in separate goroutines. Access to err is
// protected by a mutex instead of atomic.Value to allow mixing concrete types.
type callCtx struct {
	sync.WaitGroup
	sync.Mutex
	err error
}

func (ctx *callCtx) done(err error) {
	if err != nil {
		ctx.Lock()
		ctx.err = err
		ctx.Unlock()
	}
	ctx.Done()
}
