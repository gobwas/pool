package slab

import (
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// Config contains options of particular object allocation cache.
type Config struct {
	Name string
	Ctor func([]byte)
	Dtor func([]byte)
}

func (config *Config) withDefaults() (c Config) {
	if config != nil {
		c = *config
	}
	return c
}

// Cache contains logic of particular object allocation cache.
type Cache struct {
	mu       sync.RWMutex
	complete []*slab
	heap     *heap
	magic    uint32
	bufSize  int
	slabSize int
	config   Config
}

const ctlsize = int(unsafe.Sizeof(bufctl{}))

func New(size int, config *Config) *Cache {
	if size <= 0 {
		panic("size is too small")
	}
	bufsize := size + ctlsize
	return &Cache{
		heap:     newSlabHeap(2),
		slabSize: getSlabSize(bufsize),
		bufSize:  bufsize,
		magic:    rand.Uint32(),
		config:   config.withDefaults(),
	}
}

// Alloc returns at least c.Size bytes. If user has defined c.Ctor then given
// bytes contains "constructed" object.
//
// Caller must call c.Free() when usage of p ends.
func (c *Cache) Alloc() (p []byte) {
	c.mu.RLock()
	top := c.heap.Top()
	if top != nil {
		p = salloc(top)
	}
	c.mu.RUnlock()
	if p != nil {
		return
	}

	// Below is a slow path.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reorder heap to get exaclty most free slab.
	c.heap.Heapify()

	t := c.heap.Top()
	if t != nil && t != top {
		// Slabs has changed. Lookup again.
		if p = salloc(t); p != nil {
			return p
		}
	}

	s := c.grow()
	c.heap.Push(s, 0)

	if p = salloc(s); p == nil {
		panic("slab: can not alloc with new slab")
	}

	return p
}

func (c *Cache) Free(p []byte) {
	s, ref := sfree(p, c.magic)
	if ref != 0 {
		// Slab is in partial state.
		return
	}
	// Slab is complete. So give a try to move it to complete list.
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.refCount() != 0 {
		// Slab become non-complete while taking the mutex.
		return
	}

	c.heap.Remove(s)
	c.complete = append(c.complete, s)
}

// write mutex must be held.
func (c *Cache) grow() (s *slab) {
	if n := len(c.complete); n != 0 {
		s = c.complete[n-1]
		c.complete[n-1] = nil
		c.complete = c.complete[:n-1]
		return s
	}

	buf, err := syscall.Mmap(-1, 0, c.slabSize,
		syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC,
		syscall.MAP_SHARED|syscall.MAP_ANONYMOUS,
	)
	if err != nil {
		panic(err)
	}

	ctor := c.config.Ctor

	s = new(slab)
	s.init(buf, c.bufSize, func(i int32ptr, p []byte) {
		ctl, buf := splitCtl(p)
		*ctl = bufctl{
			magic: c.magic,
			slab:  s,
			index: i,
		}
		if ctor != nil {
			ctor(buf)
		}
	})

	return s
}

//func (c *Cache) Destroy() {
//
//}

// In general, if a slab contains n buffers, then the internal fragmentation is
// at most 1/n; thus the allo- cator can actually control the amount of
// internal fragmentation by controlling the slab size. How- ever, larger slabs
// are more likely to cause external fragmentation, since the probability of
// being able to reclaim a slab decreases as the number of buffers per slab
// increases. The SunOS 5.4 implementation limits internal fragmentation to
// 12.5% (1/8), since this was found to be the empirical sweet-spot in the
// trade-off between internal and external fragmenta- tion.
func getSlabSize(sz int) int {
	p := uint(syscall.Getpagesize())
	n := uint(sz) * 8 // bufsize should be at least 1/8 of slab;
	if n < p {
		return int(p)
	}
	n ^= (p - 1) // and must be multiple of page size.
	if n&(p-1) != 0 {
		n <<= 1
	}
	return int(n)
}

func salloc(s *slab) []byte {
	_, buf := s.alloc()
	if buf == nil {
		return nil
	}
	ctl, p := splitCtl(buf)
	if r := atomic.AddInt32(&ctl.ref, 1); r != 1 {
		panic("inconsistent slab state: obtained buffer has references")
	}
	return p
}

func sfree(p []byte, magic uint32) (*slab, int) {
	ctl := getCtl(p)
	if ctl.magic != magic {
		panic("freeing not known to cache bytes")
	}
	if r := atomic.AddInt32(&ctl.ref, -1); r != 0 {
		panic("inconsistent slab state: reclaimed buffer has non-zero references")
	}
	return ctl.slab, ctl.slab.free(ctl.index)
}

type bufctl struct {
	magic uint32
	slab  *slab
	index int32ptr
	ref   int32
}

type slab struct {
	buf   []byte
	size  int
	stack *stack
	ref   int32
}

func (s *slab) init(buf []byte, size int, init func(int32ptr, []byte)) {
	n := len(buf) / size

	s.buf = buf
	s.size = size
	s.stack = newStack(n)

	for i := 0; i < n; i++ {
		low := i * size
		max := low + size
		init(int32ptr(i), buf[low:max:max])
	}
}

func (s *slab) alloc() (int32ptr, []byte) {
	i := s.stack.pop()
	if i == -1 {
		return -1, nil
	}
	atomic.AddInt32(&s.ref, 1)

	low := int(i) * s.size
	max := low + s.size

	return i, s.buf[low:max:max]
}

func (s *slab) free(i int32ptr) int {
	s.stack.push(i)
	c := atomic.AddInt32(&s.ref, -1)
	return int(c)
}

func (s *slab) refCount() int {
	c := atomic.LoadInt32(&s.ref)
	return int(c)
}

func splitCtl(buf []byte) (ctl *bufctl, p []byte) {
	ctl = (*bufctl)(unsafe.Pointer(&buf[0]))
	p = buf[ctlsize:]
	return
}

func getCtl(p []byte) *bufctl {
	data := (uintptr)(unsafe.Pointer(&p[0]))
	ctl := (*bufctl)(unsafe.Pointer(data - uintptr(ctlsize)))
	return ctl
}

// using data is more safe than unsafe.Pointer(&p[0]) with nil or zero-len
// slices.
func data(p []byte) uintptr {
	return (*(*reflect.SliceHeader)(unsafe.Pointer(&p))).Data
}
