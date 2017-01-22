// Package pbytes contains tools for pooling byte buffers.
package pbytes

import (
	"sync"

	"github.com/gobwas/pool"
)

const (
	btsMinPooledSize = 128
	btsMaxPooledSize = 65536
)

var (
	buffers = map[int]*sync.Pool{}
)

func init() {
	for n := btsMinPooledSize; n <= btsMaxPooledSize; n <<= 1 {
		buffers[n] = new(sync.Pool)
	}
}

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

func PutBuf(buf []byte) {
	n := pool.CeilToPowerOfTwo(cap(buf))
	if p, ok := buffers[n]; ok {
		p.Put(buf[:0])
	}
}
