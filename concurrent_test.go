package fast

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	var a, b bool
	require.NoError(t, Call())
	require.NoError(t, Call(func() error {
		a = true
		return nil
	}))
	require.True(t, a)

	a = false
	require.NoError(t, Call(
		func() error {
			a = true
			return nil
		},
		func() error {
			b = true
			return nil
		},
	))
	require.True(t, a)
	require.True(t, b)

	require.Error(t, Call(
		func() error {
			a = false
			return nil
		},
		func() error {
			return errors.New("")
		},
	))
	require.False(t, a)
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
	Err := errors.New("")
	require.NoError(t, ForEachIO(0, func(i int) error {
		return Err
	}))
	require.Error(t, ForEachCPU(int(^uint(0)>>1), func(i int) error {
		return Err
	}))
	for j := 0; j < 4; j++ {
		for batch := range []int{0, 1, 2, 5} {
			require.Error(t, ForEach(4, batch, func(i int) error {
				if i == j {
					return Err
				}
				return nil
			}))
		}
	}

	// One batch, no early termination
	b := bytes.Repeat([]byte{'0'}, 50)
	require.Error(t, ForEach(len(b), len(b), func(i int) error {
		if i == 10 {
			return errors.New("")
		}
		b[i] = '1'
		return nil
	}))
	want := bytes.Repeat([]byte{'1'}, 50)
	want[10] = '0'
	require.Equal(t, string(want), string(b))

	// Multiple batches, early termination
	b = bytes.Repeat([]byte{'0'}, 50)
	require.Error(t, ForEach(len(b), 2, func(i int) error {
		if i == 10 {
			return errors.New("")
		}
		b[i] = '1'
		return nil
	}))
	require.Equal(t, strings.Repeat("1", 10), string(b[:10]))
	require.Equal(t, strings.Repeat("0", 10), string(b[40:]))
}
