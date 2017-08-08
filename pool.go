// Package pool contains helpers for pooling structures distinguishable by
// size.
//
// Quick example:
//
//   import "github.com/gobwas/pool"
//
//   func main() {
//      // Reuse objects in logarithmic range from 0 to 64 (0,1,2,4,6,8,16,32,64).
//      p := pool.New(0, 64, func(n int) interface{} {
//          return bytes.NewBuffer(make([]byte, 0, n))
//      })
//
//      buf, n := p.Get(10) // Returns buffer with 16 capacity.
//      defer p.Put(buf, n)
//
//      // Work with buf.
//   }
//
// There are non-generic implementations for pooling:
// - pool/pbytes for []byte reuse;
// - pool/pbufio for *bufio.Reader and *bufio.Writer reuse;
//
package pool

import "sync"

const (
	bitsize       = 32 << (^uint(0) >> 63)
	maxint        = int(1<<(bitsize-1) - 1)
	maxintHeadBit = 1 << (bitsize - 2)
)

// MakePoolMap makes map[int]*sync.Pool where map keys are logarithmic range
// from min to max.
func MakePoolMap(min, max int) map[int]*sync.Pool {
	ret := make(map[int]*sync.Pool)
	LogarithmicRange(min, max, func(n int) {
		ret[n] = new(sync.Pool)
	})
	return ret
}

// LogarithmicRange iterates from ceiled to power of two min to max,
// calling cb on each iteration.
func LogarithmicRange(min, max int, cb func(int)) {
	if min == 0 {
		min = 1
	}
	for n := CeilToPowerOfTwo(min); n <= max; n <<= 1 {
		cb(n)
	}
}

// IsPowerOfTwo reports whether given integer is a power of two.
func IsPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}

// CeilToPowerOfTwo returns the least power of two integer value greater than
// or equal to n.
func CeilToPowerOfTwo(n int) int {
	if n&maxintHeadBit != 0 && n > maxintHeadBit {
		panic("argument is too large")
	}
	if n <= 2 {
		return n
	}
	n--
	n = fillBits(n)
	n++
	return n
}

// FloorToPowerOfTwo returns the greatest power of two integer value less than
// or equal to n.
func FloorToPowerOfTwo(n int) int {
	if n <= 2 {
		return n
	}
	n = fillBits(n)
	n >>= 1
	n++
	return n
}

func fillBits(n int) int {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n
}
