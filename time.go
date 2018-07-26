package fast

import (
	"flag"
	"sync"
	"sync/atomic"
	"time"
)

var (
	timeMu sync.Mutex
	now    atomic.Value
	stop   bool
	sleep  int64
)

const rate = 250 * time.Millisecond

// Time returns the current local time. It is faster than time.Now(), but has a
// resolution of 250 ms.
func Time() time.Time {
	if t := now.Load(); t != nil {
		return t.(time.Time)
	}
	return startClock()
}

// MockTime overrides the timestamp returned by Time() and returns the current
// clock value. Zero value restarts the clock. This function can only be called
// when running unit tests.
func MockTime(t time.Time) time.Time {
	testOnly()
	restart := t.IsZero()
	if now.Load() == nil {
		if now := startClock(); restart {
			return now
		}
	}
	timeMu.Lock()
	defer timeMu.Unlock()
	if stop = !restart; restart {
		t = time.Now()
	}
	now.Store(t)
	return t
}

// Sleep pauses the current goroutine for at least the duration d. MockSleep may
// be used to control the sleep duration for unit testing.
func Sleep(d time.Duration) {
	if v := atomic.LoadInt64(&sleep); v != 0 {
		d = time.Duration(v)
	}
	time.Sleep(d)
}

// MockSleep overrides the Sleep duration. Zero duration disables the override.
// Negative duration causes Sleep to return immediately. This function can only
// be called when running unit tests.
func MockSleep(d time.Duration) {
	testOnly()
	atomic.StoreInt64(&sleep, int64(d))
}

func startClock() time.Time {
	timeMu.Lock()
	defer timeMu.Unlock()
	if t := now.Load(); t != nil {
		return t.(time.Time)
	}
	t := time.Now() // No truncation to keep the monotonic timestamp
	now.Store(t)
	go func() {
		t := time.Now()
		d := t.Truncate(rate).Add(rate).Sub(t)
		if d < rate/4 {
			d += rate
		}
		tick(<-time.After(d))
		for t := range time.Tick(rate) {
			tick(t)
		}
	}()
	return t
}

func tick(t time.Time) {
	timeMu.Lock()
	if !stop {
		now.Store(t)
	}
	timeMu.Unlock()
}

func testOnly() {
	if flag.Lookup("test.v") == nil {
		panic("fast: mock function called outside of unit testing")
	}
}
