// Package pbufio contains tools for pooling bufio.Reader and bufio.Writers.
package pbufio

import (
	"bufio"
	"io"
	"sync"

	"github.com/gobwas/pool"
)

var (
	DefaultWriterPool = NewWriterPool(256, 65536)
	DefaultReaderPool = NewReaderPool(256, 65536)
)

// GetWriter returns bufio.Writer whose buffer has at least size bytes.
// Note that size could be ceiled to the next power of two.
// GetWriter is a wrapper around DefaultWriterPool.Get().
func GetWriter(w io.Writer, size int) *bufio.Writer { return DefaultWriterPool.Get(w, size) }

// PutWriter takes bufio.Writer for future reuse.
// It does not reuse bufio.Writer which underlying buffer size is not power of
// PutWriter is a wrapper around DefaultWriterPool.Put().
func PutWriter(bw *bufio.Writer) { DefaultWriterPool.Put(bw) }

// GetReader returns bufio.Reader whose buffer has at least size bytes. It returns
// its capacity for further pass to Put().
// Note that size could be ceiled to the next power of two.
// GetReader is a wrapper around DefaultReaderPool.Get().
func GetReader(w io.Reader, size int) *bufio.Reader { return DefaultReaderPool.Get(w, size) }

// PutReader takes bufio.Reader and its size for future reuse.
// It does not reuse bufio.Reader if size is not power of two or is out of pool
// min/max range.
// PutReader is a wrapper around DefaultReaderPool.Put().
func PutReader(bw *bufio.Reader) { DefaultReaderPool.Put(bw) }

// WriterPool contains logic of *bufio.Writer reuse with various size.
type WriterPool struct {
	pool map[int]*sync.Pool
}

// NewWriterPool creates new WriterPool which reuses buffers from at least min
// and up to max buffer sizes.
func NewWriterPool(min, max int) *WriterPool {
	return &WriterPool{
		pool: pool.MakePoolMap(min, max),
	}
}

// Get returns bufio.Writer whose buffer has at least size bytes.
func (wp *WriterPool) Get(w io.Writer, size int) *bufio.Writer {
	n := pool.CeilToPowerOfTwo(size)

	pool := wp.pool[n]
	if pool == nil {
		// No such pool that could store such size.
		return bufio.NewWriterSize(w, size)
	}
	if v := pool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}

	return bufio.NewWriterSize(w, n)
}

// Put takes ownership of bufio.Writer for further reuse.
func (wp *WriterPool) Put(bw *bufio.Writer) {
	n := writerSize(bw)
	if pool := wp.pool[n]; pool != nil {
		// Should reset even if we do Reset() inside Get().
		// This is done to prevent locking underlying io.Writer from GC.
		bw.Reset(nil)
		pool.Put(bw)
	}
}

// ReaderPool contains logic of *bufio.Reader reuse with various size.
type ReaderPool struct {
	pool map[int]*sync.Pool
}

// NewReaderPool creates new ReaderPool which reuses buffers from at least min
// and up to max buffer sizes.
func NewReaderPool(min, max int) *ReaderPool {
	return &ReaderPool{
		pool: pool.MakePoolMap(min, max),
	}
}

// Get returns bufio.Reader whose buffer has at least size bytes.
func (rp *ReaderPool) Get(r io.Reader, size int) *bufio.Reader {
	n := pool.CeilToPowerOfTwo(size)

	pool := rp.pool[n]
	if pool == nil {
		// No such pool that could store such size.
		return bufio.NewReaderSize(r, size)
	}
	if v := pool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}

	return bufio.NewReaderSize(r, n)
}

// Put takes ownership of bufio.Reader for further reuse.
func (rp *ReaderPool) Put(br *bufio.Reader) {
	size := readerSize(br)
	if pool := rp.pool[size]; pool != nil {
		// Should reset even if we do Reset() inside Get().
		// This is done to prevent locking underlying io.Reader from GC.
		br.Reset(nil)
		pool.Put(br)
	}
}

// writerSize returns buffer size of the given buffered writer.
func writerSize(bw *bufio.Writer) int {
	return bw.Available() + bw.Buffered()
}

// readerSize returns buffer size of the given buffered reader.
// NOTE: current implementation reset underlying io.Reader.
//
// TODO(gobwas): this workaround should be moved under tag when go 1.10
//               will be released.
func readerSize(br *bufio.Reader) int {
	br.Reset(sizeReader)
	br.ReadByte()
	n := br.Buffered() + 1
	br.Reset(nil)
	return n
}

var sizeReader optimisticReader

type optimisticReader struct{}

func (optimisticReader) Read(p []byte) (int, error) {
	return len(p), nil
}
