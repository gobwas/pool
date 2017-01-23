// Package pool contains helpers for pooling structures.
package pool

import "sync"

// MakePoolMap makes map[int]*sync.Pool where map keys are
// powers of two from ceiled to power of two min to max.
func MakePoolMap(min, max int) map[int]*sync.Pool {
	ret := make(map[int]*sync.Pool)
	PowerOfTwoRange(min, max, func(n int) {
		ret[n] = new(sync.Pool)
	})
	return ret
}

// PowerOfTwoRange iterates from ceiled to power of two min to max,
// calling cb on each iteration.
func PowerOfTwoRange(min, max int, cb func(int)) {
	if min == 0 {
		min = 1
	}
	for n := CeilToPowerOfTwo(min); n <= max; n <<= 1 {
		cb(n)
	}
}

// CeilToPowerOfTwo rounds n to the highest power of two integer.
func CeilToPowerOfTwo(n int) int {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
