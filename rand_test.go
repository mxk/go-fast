package fast

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandUint64(t *testing.T) {
	a := RandUint64()
	b := RandUint64()
	if a == b {
		b = RandUint64()
	}
	require.NotEqual(t, a, b)
}

func TestRandID(t *testing.T) {
	require.Empty(t, RandID(0))
	n, seen := 0, [256]bool{}
	start := time.Now()
	for time.Since(start) < 3*time.Second {
		id := RandID(2)
		a, b := id[0], id[1]
		require.True(t, 'A' <= a && a <= 'Z')
		require.True(t, ('A' <= b && b <= 'Z') ||
			('a' <= b && b <= 'z') ||
			('0' <= b && b <= '9'))
		if !seen[b] {
			seen[b] = true
			if n++; n == 62 {
				break
			}
		}
	}
	require.Equal(t, 62, n)
}

func TestRandStr(t *testing.T) {
	require.Empty(t, RandStr(0, func(int) string { return "x" }))
	require.Equal(t, "---", RandStr(3, func(int) string { return "" }))
	require.Equal(t, "xxx", RandStr(3, func(int) string { return "x" }))
	s := RandStr(1024, func(i int) string {
		switch i {
		case 0:
			return "a"
		case 1023:
			return "z"
		default:
			return "12"
		}
	})
	require.Len(t, s, 1024)
	assert.Equal(t, "a", s[:1])
	assert.Equal(t, "z", s[len(s)-1:])
	assert.Equal(t, 1, strings.Count(s, "a"))
	assert.Equal(t, 1, strings.Count(s, "z"))
	assert.True(t, strings.Count(s, "1") > 256)
	assert.True(t, strings.Count(s, "2") > 256)
}
