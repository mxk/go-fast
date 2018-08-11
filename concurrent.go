// Package fast contains utility functions and types aimed at improving
// performance.
package fast

import (
	"runtime"
	"sync"
)

// Call calls each fn in a separate goroutine, waits for completion of all
// calls, and returns the first (in parameter order) non-nil error.
func Call(fn ...func() error) error {
	var err error
	if len(fn) > 1 {
		var ctx callCtx
		ctx.Add(len(fn))
		last := len(fn) - 1
		for i, f := range fn[:last] {
			go func(ctx *callCtx, i int, f func() error) {
				ctx.done(i, f())
			}(&ctx, i, f)
		}
		ctx.done(last, fn[last]())
		ctx.Wait()
		err = ctx.err
	} else if len(fn) == 1 {
		err = fn[0]()
	}
	return err
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
// called for each task with i in the range [0,n). If any call returns an error,
// all pending tasks are canceled and the error associated with the lowest i is
// returned.
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
		last := n - 1
		for i := 0; i < last; i++ {
			go func(ctx *callCtx, i int, fn func(i int) error) {
				ctx.done(i, fn(i))
			}(&ctx, i, fn)
		}
		ctx.done(last, fn(last))
		ctx.Wait()
		return ctx.err
	}

	// Start waiter goroutine
	var wg sync.WaitGroup
	ich := make(chan int)
	ech := make(chan ordErr)
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
				ech <- ordErr{i, err}
			} else {
				for i = range ich {
					if err = fn(i); err != nil {
						ech <- ordErr{i, err}
						break
					}
				}
			}
		}(i)
	}

	// Send remaining tasks while waiting for the first error
	var err ordErr
	for i := batch; i < n; i++ {
		select {
		case ich <- i:
			continue
		case e := <-ech:
			err.set(e.ord, e.err)
		}
		break
	}
	close(ich)

	// Wait for completion of all running tasks
	for e := range ech {
		err.set(e.ord, e.err)
	}
	return err.err
}

// callCtx synchronizes function calls in separate goroutines.
type callCtx struct {
	sync.WaitGroup
	sync.Mutex
	ordErr
}

func (ctx *callCtx) done(ord int, err error) {
	if err != nil {
		ctx.Lock()
		ctx.set(ord, err)
		ctx.Unlock()
	}
	ctx.Done()
}

// ordErr is an ordered error.
type ordErr struct {
	ord int
	err error
}

func (e *ordErr) set(ord int, err error) {
	if ord < e.ord || e.err == nil {
		e.ord = ord
		e.err = err
	}
}
