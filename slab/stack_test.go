package slab

import (
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestStackNew(t *testing.T) {
	s := newStack(2)
	if n := s.items[0].next; n != 1 {
		t.Fatalf("#0 next is %d; want 1", n)
	}
	if n := s.items[1].next; n != -1 {
		t.Fatalf("#1 next is %d; want -1", n)
	}
}

func TestStackParallel(t *testing.T) {
	for _, test := range []struct {
		size        int
		parallelism int
	}{
		{1 << 16, runtime.NumCPU()},
		{1 << 16, 100},
		{1 << 24, runtime.NumCPU()},
		{1 << 24, 100},
	} {
		name := fmt.Sprintf("%d@%d", test.size, test.parallelism)
		t.Run(name, func(t *testing.T) {
			s := newStack(test.size)

			run := make(chan struct{})
			ret := make(chan []int, test.parallelism)

			for i := 0; i < test.parallelism; i++ {
				go func(j int) {
					var ids []int
					defer func() { ret <- ids }()

					<-run

					for i := 0; i < 1000+rand.Intn(test.size); i++ {
						i := s.pop()
						if i == -1 {
							break
						}
						ids = append(ids, int(i))
					}
					for _, i := range rand.Perm(len(ids)) {
						s.push(int32ptr(ids[i]))
					}
					ids = ids[:0]
					for {
						i := s.pop()
						if i == -1 {
							for _, t := range []int{0, 1, 5, 10, 15, 20, 50} {
								time.Sleep(time.Duration(t) * time.Millisecond)
								if i = s.pop(); i != -1 {
									break
								}
							}
							if i == -1 {
								break
							}
						}
						ids = append(ids, int(i))
					}
				}(i)
			}

			close(run)

			var ids []int
			for i := 0; i < test.parallelism; i++ {
				x := <-ret
				ids = append(ids, x...)
			}
			if n := len(ids); n != test.size {
				t.Fatalf("recevied %d; want %d", n, test.size)
			}
			sort.Ints(ids)
			for i := 0; i < len(ids); i++ {
				if int(ids[i]) != i {
					t.Fatalf("#%d item is %v; want %v", i, ids[i], i)
				}
			}
		})
	}
}

func BenchmarkStack(b *testing.B) {
	for _, size := range []int{
		128,
	} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			s := newStack(size)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if i := s.pop(); i != -1 {
						s.push(i)
					}
				}
			})
		})
	}
}
