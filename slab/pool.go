package slab

import "github.com/gobwas/pool"

type Pool struct {
	cache map[int]*Cache
}

func NewPool(min, max int) *Pool {
	c := make(map[int]*Cache)
	pool.LogarithmicRange(min, max, func(n int) {
		c[n] = New(n, nil)
	})
	return &Pool{c}
}

func (p *Pool) Get(n, c int) []byte {
	if n > c {
		panic("requested length is greater than capacity")
	}

	x := pool.CeilToPowerOfTwo(c)

	cache, ok := p.cache[x]
	if !ok {
		// No such cache that could store such capacity.
		return make([]byte, n, c)
	}

	bts := cache.Alloc()
	return bts[:n]
}

func (p *Pool) Put(bts []byte) {
	c := cap(bts)
	if cache, ok := p.cache[c]; ok {
		cache.Free(bts)
	}
}

func (p *Pool) GetLen(n int) []byte {
	return p.Get(n, n)
}

func (p *Pool) GetCap(n int) []byte {
	return p.Get(0, n)
}
