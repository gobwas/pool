// +build !pool_sanitize

package pbytes

import (
	"sync"

	"github.com/gobwas/pool"
)

// Pool contains logic of reusing byte slices of various size.
type Pool struct {
	pool map[int]*sync.Pool
}

// New creates new Pool which reuses min and max sized slices.
// Note that min is ceiled to the next power of two.
func New(min, max int) *Pool {
	return &Pool{
		pool: pool.MakePoolMap(min, max),
	}
}

// Get returns probably reused slice of bytes with at least capacity of c and
// exactly len of n.
func (p *Pool) Get(n, c int) []byte {
	if n > c {
		panic("requested length is greater than capacity")
	}

	x := pool.CeilToPowerOfTwo(c)

	pool, ok := p.pool[x]
	if !ok {
		// No such pool that could store such capacity.
		return make([]byte, n, c)
	}
	if v := pool.Get(); v != nil {
		bts := v.([]byte)
		bts = bts[:n]
		return bts
	}

	return make([]byte, n, x)
}

// Put returns given slice to reuse pool.
// It does not reuse bytes whose size is not power of two or is out of pool
// min/max range.
func (p *Pool) Put(bts []byte) {
	c := cap(bts)
	if pool, ok := p.pool[c]; ok {
		pool.Put(bts)
	}
}

// GetCap returns probably reused slice of bytes with at least capacity of n.
func (p *Pool) GetCap(c int) []byte {
	return p.Get(0, c)
}

// GetLen returns probably reused slice of bytes with at least capacity of n
// and exactly len of n.
func (p *Pool) GetLen(n int) []byte {
	return p.Get(n, n)
}
