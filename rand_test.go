package fast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandBytes(t *testing.T) {
	p := randPool.New().(*randPage)
	randPool.Put(p)

	// p not used
	RandBytes(make([]byte, 4096))
	require.Equal(t, 0, p.n)

	// p used
	want := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	copy(p.r[:], want)
	p.n = len(want)
	require.Equal(t, want, RandBytes(make([]byte, 10)))
	require.Equal(t, 0, p.n)

	// refill
	var seen [256]bool
	b := make([]byte, 512)
	for i := 0; i < 8; i++ {
		l := RandBytes(b[:256])
		h := RandBytes(b[256:])
		require.NotEqual(t, l, h)
		for _, v := range b {
			seen[v] = true
		}
	}
	require.NotContains(t, seen[:], false)
}
