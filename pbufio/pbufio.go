// Package pbufio contains tools for pooling bufio.Reader and bufio.Writers.
package pbufio

import (
	"bufio"
	"io"

	"github.com/gobwas/pool"
)

const defaultBufSize = 4096

var (
	writers = pool.MakePoolMap(256, 65536)
	readers = pool.MakePoolMap(256, 65536)
)

// GetWriter returns bufio.Writer with given buffer size.
// If size <= 0 the default buffer size is used.
// Note that size is rounded up to nearest highest power of two.
func GetWriter(w io.Writer, size int) *bufio.Writer {
	if size <= 0 {
		size = defaultBufSize
	}
	n := pool.CeilToPowerOfTwo(size)

	if p, ok := writers[n]; ok {
		if v := p.Get(); v != nil {
			ret := v.(*bufio.Writer)
			ret.Reset(w)
			return ret
		}
	}

	return bufio.NewWriterSize(w, size)
}

// PutWriter takes bufio.Writer for future reuse.
func PutWriter(w *bufio.Writer) {
	w.Reset(nil)
	n := pool.CeilToPowerOfTwo(w.Available())
	if p, ok := writers[n]; ok {
		p.Put(w)
	}
}

// GetReader returns bufio.Reader with given buffer size.
// If size <= 0 the default buffer size is used.
// Note that size is rounded up to nearest highest power of two.
func GetReader(r io.Reader, size int) *bufio.Reader {
	if size <= 0 {
		size = defaultBufSize
	}
	n := pool.CeilToPowerOfTwo(size)

	if p, ok := readers[n]; ok {
		if v := p.Get(); v != nil {
			ret := v.(*bufio.Reader)
			ret.Reset(r)
			return ret
		}
	}

	return bufio.NewReaderSize(r, size)
}

// PutReader takes bufio.Reader for future reuse.
// Note that size should be the same as in GetReader call.
// If size <= 0 the default buffer size is used.
// Note that size is rounded up to nearest highest power of two.
func PutReader(r *bufio.Reader, size int) {
	if size == 0 {
		size = defaultBufSize
	}
	n := pool.CeilToPowerOfTwo(size)

	if p, ok := readers[n]; ok {
		r.Reset(nil)
		p.Put(r)
	}
}
