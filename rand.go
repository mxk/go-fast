package fast

import "crypto/rand"

// RandBytes returns b after filling it with random bytes from a CSPRNG.
func RandBytes(b []byte) []byte {
	// Keeping this simple for now, maybe optimize later
	if _, err := rand.Read(b); err != nil {
		panic("fast: crypto/rand error: " + err.Error())
	}
	return b
}

// RandID returns a random alphanumeric (base62) string of length n with the
// first character always an upper-case letter.
func RandID(n int) string {
	const b62 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	buf := RandBytes(make([]byte, n+n))
	max, mask := byte(26), byte(31)
	for i := range buf[:n] {
		for {
			if r := buf[i] & mask; r < max {
				buf[i] = b62[r]
				max, mask = 62, 63
				break
			}
			if len(buf) == n {
				buf = buf[:cap(buf)]
				RandBytes(buf[n:])
			}
			j := len(buf) - 1
			buf[i] = buf[j]
			buf = buf[:j]
		}
	}
	return string(buf[:n])
}
