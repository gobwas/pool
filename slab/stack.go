package slab

import (
	"math"
	"sync/atomic"
)

type int32ptr int

type stack struct {
	// items contains fixed-size slice of items, which helps to use 32-bit
	// indexes in atomic operations as pointers.
	items []item

	// head contains 32-bit index of head item in items array and 32-bit tag
	// defining consistency state.
	head uint64

	// tag is used as atomic counter which value is used as "tag" in CAS
	// operations to solve ABA problem.
	// See https://en.wikipedia.org/wiki/ABA_problem
	tag uint32
}

// item represents object stored in stack.
type item struct {
	next int32ptr
}

func newStack(n int) *stack {
	if n <= 0 {
		panic("stack size is too small")
	}
	if n > math.MaxInt32 {
		panic("stack size is too big")
	}
	items := make([]item, n)
	for i := n - 1; i > 0; i-- {
		items[i-1].next = int32ptr(i)
	}
	items[n-1].next = -1

	return &stack{
		items: items,
		head:  encode(0, 0),
	}
}

func (s *stack) pop() int32ptr {
	tag := atomic.AddUint32(&s.tag, 1)

	for {
		head := s.head
		i := decode(head)
		if i < 0 {
			return -1
		}

		next := s.items[i].next

		if atomic.CompareAndSwapUint64(
			&s.head, head,
			encode(next, tag),
		) {
			return i
		}
	}
}

func (s *stack) push(i int32ptr) {
	tag := atomic.AddUint32(&s.tag, 1)

	for {
		head := s.head
		next := decode(head)
		s.items[i].next = next

		if atomic.CompareAndSwapUint64(
			&s.head, head,
			encode(i, tag),
		) {
			break
		}
	}
}

func decode(v uint64) (ptr int32ptr) {
	return int32ptr(int32(v >> 32))
}
func encode(ptr int32ptr, tag uint32) uint64 {
	return (uint64(ptr) << 32) | uint64(tag)
}
