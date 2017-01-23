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
// Note that size is rounded up to nearest highest power of two.
func GetWriter(w io.Writer, size int) *bufio.Writer {
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
// Note that size should be the same as used to acquire writer.
// If you have acquired writer from AcquireWriter function, set size to 0.
// If size == 0 then default buffer size is used.
func PutWriter(w *bufio.Writer, size int) {
	if size == 0 {
		size = defaultBufSize
	}
	n := pool.CeilToPowerOfTwo(size)
	if p, ok := writers[n]; ok {
		w.Reset(nil)
		p.Put(w)
	}
}

// GetReader returns bufio.Reader with given buffer size.
// Note that size is rounded up to nearest highest power of two.
func GetReader(r io.Reader, size int) *bufio.Reader {
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
// Note that size should be the same as used to acquire reader.
// If you have acquired reader from AcquireReader function, set size to 0.
// If size == 0 then default buffer size is used.
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
