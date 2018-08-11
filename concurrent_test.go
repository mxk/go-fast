package fast

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	var a, b bool
	assert.NoError(t, Call())
	assert.NoError(t, Call(func() error {
		a = true
		return nil
	}))
	assert.True(t, a)

	a = false
	assert.NoError(t, Call(
		func() error {
			a = true
			return nil
		},
		func() error {
			b = true
			return nil
		},
	))
	assert.True(t, a)
	assert.True(t, b)

	// Error return
	a = false
	e1, e2 := fmt.Errorf("e1"), fmt.Errorf("e2")
	assert.Equal(t, e2, Call(
		func() error {
			a = true
			return nil
		},
		func() error { return e2 },
	))
	assert.True(t, a)

	for i := time.Duration(0); i < 5; i++ {
		require.Equal(t, e1, Call(
			func() error {
				time.Sleep(i * time.Millisecond)
				return e1
			},
			func() error {
				time.Sleep((4 - i) * time.Millisecond)
				return e2
			},
		))
	}
}

func TestForEach(t *testing.T) {
	// Various combinations of different n and batch values
	for n := 0; n <= 256; {
		b, c := bytes.Repeat([]byte{'0'}, n), byte('1')
		for batch := range []int{0, 1, 2, n / 2, n, n * 2} {
			ForEach(n, batch, func(i int) error {
				require.Equal(t, c-1, b[i])
				b[i] = c
				return nil
			})
			require.Equal(t, strings.Repeat(string(c), n), string(b))
			c++
		}
		if n *= 2; n == 0 {
			n = 1
		}
	}

	// Error return
	e := fmt.Errorf("e")
	require.NoError(t, ForEachIO(0, func(i int) error { return e }))
	require.Equal(t, e, ForEachCPU(int(^uint(0)>>1), func(i int) error {
		return e
	}))
	for j, n := 0, 8; j < n; j++ {
		for batch := range []int{0, 1, 2, 5, 10} {
			require.EqualError(t, ForEach(n, batch, func(i int) error {
				if i >= j {
					return fmt.Errorf("%d", i)
				}
				return nil
			}), fmt.Sprint(j))
		}
	}

	// One batch, no early termination
	b := bytes.Repeat([]byte{'0'}, 50)
	require.EqualError(t, ForEach(len(b), len(b), func(i int) error {
		if i == 10 || i == 40 {
			return fmt.Errorf("%d", i)
		}
		b[i] = '1'
		return nil
	}), "10")
	want := bytes.Repeat([]byte{'1'}, 50)
	want[10] = '0'
	want[40] = '0'
	require.Equal(t, string(want), string(b))

	// Multiple batches, early termination
	b = bytes.Repeat([]byte{'0'}, 50)
	require.EqualError(t, ForEach(len(b), 2, func(i int) error {
		if i == 10 || i == 40 {
			return fmt.Errorf("%d", i)
		}
		b[i] = '1'
		return nil
	}), "10")
	require.Equal(t, strings.Repeat("1", 10), string(b[:10]))
	require.Equal(t, strings.Repeat("0", 10), string(b[40:]))
}
