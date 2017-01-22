// Package pool contains helpers for pooling structures.
package pool

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
