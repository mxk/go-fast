package fast

import (
	"runtime"
	"sync"
)

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
	var wg sync.WaitGroup
	if n <= batch {
		var mu sync.Mutex
		var anyErr error
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func(i int) {
				defer wg.Done()
				if err := fn(i); err != nil {
					// Using mutex instead of atomic.Value to allow mixed types
					mu.Lock()
					anyErr = err
					mu.Unlock()
				}
			}(i)
		}
		wg.Wait()
		return anyErr
	}

	// Start waiter goroutine
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
