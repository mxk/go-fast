package fast

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	// startClock() only runs once, so pick a path at random
	if RandUint64()&1 == 0 {
		Time()
	} else {
		MockTime(time.Now())
		MockTime(time.Time{})
	}

	t1 := Time()
	d := time.Now().Sub(t1)
	assert.True(t, 0 <= d && d < 10*time.Millisecond, "d=%v", d)

	time.Sleep(rate + rate/2)
	d = Time().Sub(t1)
	assert.True(t, rate/4 <= d && d < 2*rate-rate/4, "d=%v", d)

	for _, mt := range []time.Time{time.Unix(1, 0), time.Unix(2, 2)} {
		assert.Equal(t, mt, MockTime(mt), "mt=%v", mt)
		assert.Equal(t, mt, Time(), "mt=%v", mt)
	}

	assert.Equal(t, MockTime(time.Time{}), Time())
	d = time.Now().Sub(Time())
	assert.True(t, 0 <= d && d < 10*time.Millisecond, "d=%v", d)
}

func TestSleep(t *testing.T) {
	const d = 50 * time.Millisecond
	MockSleep(d)
	start := time.Now()
	Sleep(0)
	assert.True(t, time.Since(start) >= d)

	MockSleep(-1)
	start = time.Now()
	Sleep(d)
	assert.True(t, time.Since(start) < d/2)

	MockSleep(0)
	start = time.Now()
	Sleep(d)
	assert.True(t, time.Since(start) >= d)
}

func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now()
	}
}

func BenchmarkFastTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Time()
	}
}
