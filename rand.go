package fast

import (
	"crypto/rand"
	"encoding/binary"
	"math/bits"
)

// RandBytes returns b after filling it with random bytes from a CSPRNG.
func RandBytes(b []byte) []byte {
	// Keeping this simple for now, maybe optimize later
	if _, err := rand.Read(b); err != nil {
		panic("fast: crypto/rand error: " + err.Error())
	}
	return b
}

// RandUint64 returns a random uint64 value from a CSPRNG.
func RandUint64() uint64 {
	var b [8]byte
	return binary.LittleEndian.Uint64(RandBytes(b[:]))
}

// RandID returns a random alphanumeric (base62) string of length n with the
// first character always an upper-case letter. The total amount of entropy is:
//
//	4.7 + (n-1)*5.954
func RandID(n int) string {
	return RandStr(n, func(i int) string {
		if i == 0 {
			return "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		}
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	})
}

// RandStr generates a random ASCII string of length n. It calls dict with i in
// the range [0,n) and selects a random byte from the returned string, which
// must be between 1 and 256 bytes long.
func RandStr(n int, dict func(i int) string) string {
	b := RandBytes(make([]byte, n+n+4))
	for i := range b[:n] {
		s := dict(i)
		if len(s) <= 1 {
			if len(s) == 0 {
				b[i] = '-'
			} else {
				b[i] = s[0]
			}
			continue
		}
		mask := byte(1<<bits.Len8(uint8(len(s)-1)) - 1)
		for {
			if r := b[i] & mask; int(r) < len(s) {
				b[i] = s[r]
				break
			}
			if len(b) == n {
				b = b[:cap(b)]
				RandBytes(b[n:])
			}
			b[i], b = b[len(b)-1], b[:len(b)-1]
		}
	}
	return string(b[:n])
}
