package fast

import (
	"crypto/rand"
	"sync"
)

// randPool contains pages of random bytes.
var randPool = sync.Pool{New: func() interface{} { return new(randPage) }}

// RandBytes returns b after filling it with random bytes from a CSPRNG.
func RandBytes(b []byte) []byte {
	if len(b) > len((*randPage)(nil).r)/2 {
		randRead(b)
	} else {
		p := randPool.Get().(*randPage)
		p.copy(b)
		randPool.Put(p)
	}
	return b
}

// randPage contains random bytes.
type randPage struct {
	n int
	r [4088]byte
}

// copy transfers random bytes from p to b.
func (p *randPage) copy(b []byte) {
	if p.n < len(b) {
		randRead(p.r[p.n:])
		p.n = len(p.r)
	}
	p.n -= copy(b, p.r[p.n-len(b):p.n])
}

// randRead fills b with random bytes from crypto/rand.
func randRead(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("fast: crypto/rand error: " + err.Error())
	}
}
