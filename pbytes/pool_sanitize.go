// +build pool_sanitize

package pbytes

import (
	"reflect"
	"syscall"
	"unsafe"
)

const magic = uint64(0x777742)

type guard struct {
	magic uint64
	size  int
}

const guardSize = int(unsafe.Sizeof(guard{}))

type Pool struct {
	min, max int
}

func New(min, max int) *Pool {
	return &Pool{min, max}
}

// Get returns probably reused slice of bytes with at least capacity of c and
// exactly len of n.
func (p *Pool) Get(n, c int) []byte {
	if n > c {
		panic("requested length is greater than capacity")
	}
	if p.min > c || c > p.max {
		return make([]byte, n, c)
	}

	pageSize := syscall.Getpagesize()
	pages := (c+guardSize)/pageSize + 1
	size := pages * pageSize

	bts := make([]byte, 0, size)
	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&bts))
	data := hdr.Data

	g := (*guard)(unsafe.Pointer(data))
	*g = guard{
		magic: magic,
		size:  size,
	}

	return bts[guardSize : guardSize+n]
}

func (p *Pool) GetCap(c int) []byte { return p.Get(0, c) }
func (p *Pool) GetLen(n int) []byte { return Get(n, n) }

// Put returns given slice to reuse pool.
func (p *Pool) Put(bts []byte) {
	if cap(bts) < p.min {
		return
	}

	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&bts))
	data := hdr.Data - uintptr(guardSize)

	g := (*guard)(unsafe.Pointer(data))
	if g.magic != magic {
		panic("unknown slice returned to the pool")
	}

	// Disable read and write on bytes memory pages. This will cause panic on
	// incorrect access to returned slice.
	mprotect(data, g.size)
}

func mprotect(ptr uintptr, size int) {
	var (
		// Need to avoid "EINVAL addr is not a valid pointer,
		// or not a multiple of PAGESIZE."
		start = ptr & ^(uintptr(syscall.Getpagesize() - 1))
		prot  = uintptr(syscall.PROT_NONE)
	)
	_, _, err := syscall.Syscall(syscall.SYS_MPROTECT,
		start, uintptr(size), prot,
	)
	if err != 0 {
		panic(err.Error())
	}
}
