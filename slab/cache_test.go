package slab

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
)

func TestGetSlabSize(t *testing.T) {
	for _, test := range []struct {
		name string
		size int
		page int
		exp  int
	}{
		{
			size: 1,
			page: 64,
			exp:  64,
		},
		{
			size: 10,
			page: 64,
			exp:  128,
		},
		{
			size: 64,
			page: 64,
			exp:  512,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			act := getSlabSize(test.size, test.page)
			if exp := test.exp; act != exp {
				t.Errorf(
					"getSlabSize(%d, %d) = %d; want %d",
					test.size, test.page, act, exp,
				)
			}
		})
	}
}

func TestCache(t *testing.T) {
	for _, test := range []struct {
		size  int
		alloc int
	}{
		{
			size:  1,
			alloc: 10,
		},
		{
			size:  128,
			alloc: 10000,
		},
		{
			size:  4096,
			alloc: 108,
		},
	} {
		t.Run("", func(t *testing.T) {
			c := New(test.size, nil)

			var b [][]byte
			for i := 0; i < test.alloc; i++ {
				p := c.Alloc()
				if n := len(p); n != test.size {
					t.Fatalf(
						"c.Alloc() returned %d-len slice; want %d",
						n, test.size,
					)
				}
				if n := cap(p); n != test.size {
					t.Fatalf(
						"c.Alloc() returned %d-cap slice; want %d",
						n, test.size,
					)
				}
				log.Println(info(c))
				b = append(b, p)
			}
			for _, p := range b {
				c.Free(p)
				log.Println(info(c))
			}
		})
	}
}

func info(c *Cache) (slabs, complete, bytes, capacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.heap.ForEach(func(s *slab, w int) bool {
		slabs += 1
		bytes += len(s.buf)
		capacity += len(s.stack.items) - int(atomic.LoadInt32(&s.ref))
		return true
	})
	complete = len(c.complete)
	for _, s := range c.complete {
		slabs += 1
		bytes += len(s.buf)
		capacity += len(s.stack.items)
	}
	return
}

var benchmarkCases = []struct {
	size      int
	freeAfter int
}{
	{
		size:      128,
		freeAfter: 1000,
	},
	{
		size:      100,
		freeAfter: 4096 / 100,
	},
	{
		size:      32,
		freeAfter: 4096 / 32,
	},
	{
		size:      512,
		freeAfter: 4096 / 512,
	},
	{
		size:      1024,
		freeAfter: 4096 / 1024,
	},
}

func BenchmarkAlloc(b *testing.B) {
	for _, test := range benchmarkCases {
		b.Run(
			fmt.Sprintf("%d/%d(slab)", test.size, test.freeAfter),
			func(b *testing.B) {
				c := New(test.size, nil)
				f := make([][]byte, test.freeAfter)
				for i := range f {
					f[i] = c.Alloc()
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					j := i % len(f)
					if j == 0 {
						for _, p := range f {
							c.Free(p)
						}
					}
					f[j] = c.Alloc()
				}
			},
		)
		b.Run(
			fmt.Sprintf("%d/%d(sync.Pool)", test.size, test.freeAfter),
			func(b *testing.B) {
				pool := sync.Pool{
					New: func() interface{} {
						return make([]byte, test.size)
					},
				}

				f := make([][]byte, test.freeAfter)
				for i := range f {
					f[i] = pool.Get().([]byte)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					j := i % len(f)
					if j == 0 {
						for _, p := range f {
							pool.Put(p)
						}
					}
					f[j] = pool.Get().([]byte)
				}
			},
		)
	}
}

func BenchmarkAllocParallel(b *testing.B) {
	for _, test := range benchmarkCases {
		b.Run(
			fmt.Sprintf("%d/%d(slab)", test.size, test.freeAfter),
			func(b *testing.B) {
				c := New(test.size, nil)
				f := make(chan []byte, test.freeAfter)
				go func() {
					ps := make([][]byte, test.freeAfter)
					var i int
					for p := range f {
						ps[i] = p
						i++
						if i < len(ps) {
							continue
						}
						for _, p := range ps {
							c.Free(p)
						}
						i = 0
					}
				}()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						f <- c.Alloc()
					}
				})
			},
		)
		b.Run(
			fmt.Sprintf("%d/%d(sync.Pool)", test.size, test.freeAfter),
			func(b *testing.B) {
				pool := sync.Pool{
					New: func() interface{} {
						return make([]byte, test.size)
					},
				}

				f := make(chan []byte, test.freeAfter)
				go func() {
					ps := make([][]byte, test.freeAfter)
					var i int
					for p := range f {
						ps[i] = p
						i++
						if i < len(ps) {
							continue
						}
						for _, p := range ps {
							pool.Put(p)
						}
						i = 0
					}
				}()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						f <- pool.Get().([]byte)
					}
				})
			},
		)
	}
}
