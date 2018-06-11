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
