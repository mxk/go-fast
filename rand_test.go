package fast

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
