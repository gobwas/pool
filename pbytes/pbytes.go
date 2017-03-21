// Package pbytes contains tools for pooling byte buffers.
package pbytes

import "github.com/gobwas/pool"

var buffers = pool.MakePoolMap(128, 65536)

func GetBufCap(n int) []byte {
	n = pool.CeilToPowerOfTwo(n)

	if p, ok := buffers[n]; ok {
		if buf := p.Get(); buf != nil {
			return buf.([]byte)
		}
	}

	return make([]byte, 0, n)
}

func GetBufLen(n int) []byte {
	bts := GetBufCap(n)
	bts = bts[:n]
	return bts
}

func GetBuf(length, capacity int) []byte {
	bts := GetBufCap(capacity)
	bts = bts[:length]
	return bts
}

func PutBuf(buf []byte) {
	n := pool.CeilToPowerOfTwo(cap(buf))
	if p, ok := buffers[n]; ok {
		p.Put(buf[:0])
	}
}
