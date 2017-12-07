// THIS FILE WAS AUTOGENERATED.
// DO NOT EDIT!

package slab

type recordheap struct {
	x *slab
	w int
}

type heap struct {
	d     int
	data  []recordheap
	index map[*slab]int
}

func newSlabHeap(d int) *heap {
	return &heap{
		d:     d,
		index: make(map[*slab]int),
	}
}

func newSlabHeapFromSlice(data []*slab, d int) *heap {
	records := make([]recordheap, len(data))
	for i, x := range data {
		records[i] = recordheap{x: x}
	}
	h := &heap{
		d:     d,
		data:  records,
		index: make(map[*slab]int),
	}
	for i := len(h.data)/h.d - 1; i >= 0; i-- {
		h.siftDown(i)
	}
	return h
}

func (h *heap) Top() *slab {
	if len(h.data) == 0 {
		return nil
	}
	return h.data[0].x
}

func (h *heap) Pop() *slab {
	n := len(h.data)
	ret := h.data[0].x
	a, b := h.data[0], h.data[n-1]
	h.index[a.x], h.index[b.x] = n-1, 0
	h.data[0], h.data[n-1] = h.data[n-1], h.data[0]
	h.data[n-1] = recordheap{}
	h.data = h.data[:n-1]
	delete(h.index, ret)
	h.siftDown(0)
	return ret
}

func (h *heap) Slice() []*slab {
	cp := *h
	cp.data = make([]recordheap, len(h.data))
	copy(cp.data, h.data)
	cp.index = make(map[*slab]int) // prevent reordering original index
	ret := make([]*slab, len(cp.data))
	n := len(cp.data)
	for i := 0; i < n; i++ {
		ret[i] = cp.data[0].x
		cp.data = cp.data[1:]
		cp.siftDown(0)
	}
	return ret
}

func (h *heap) Len() int    { return len(h.data) }
func (h *heap) Empty() bool { return len(h.data) == 0 }

func (h *heap) Push(x *slab, w int) {
	_, ok := h.index[x]
	if ok {
		panic("could not push value that is already present in heap")
	}
	r := recordheap{x, w}
	i := len(h.data)
	if cap(h.data) == len(h.data) {
		h.data = append(h.data, r)
	} else {
		h.data = h.data[:i+1]
		h.data[i] = r
	}
	h.index[x] = i
	h.siftUp(i)
}

func (h *heap) Heapify() {
	for i := len(h.data)/h.d - 1; i >= 0; i-- {
		h.siftDown(i)
	}
}

func (h *heap) WithPriority(x *slab, fn func(int) int) {
	i, ok := h.index[x]
	if !ok {
		panic("could not update value that is not present in heap")
	}
	h.update(i, recordheap{x, fn(h.data[i].w)})
}

func (h *heap) ChangePriority(x *slab, w int) {
	i, ok := h.index[x]
	if !ok {
		panic("could not update value that is not present in heap")
	}
	h.update(i, recordheap{x, w})
}

func (h *heap) Compare(a, b *slab) int {
	var i, j int
	i, ok := h.index[a]
	if ok {
		j, ok = h.index[b]
	}
	if !ok {
		panic("comparing record that not in heap")
	}
	return h.data[i].w - h.data[j].w
}

func (h *heap) Remove(x *slab) {
	i, ok := h.index[x]
	if !ok {
		return
	}
	h.siftTop(i)
	h.Pop()
}

// Ascend calls it for every item in heap in order.
func (h *heap) Ascend(it func(x *slab, w int) bool) {
	n := len(h.data)
	restore := h.data
	for i := 0; i < n; i++ {
		if !it(h.data[0].x, h.data[0].w) {
			break
		}
		h.data = h.data[1:]
		h.siftDown(0)
	}
	h.data = restore
	// No need to make h.Heapify() cause we get top element the same.
	// The rest of heap will be rebuilt during lifetime.
}

// ForEach calls it for every item in heap not in order.
func (h *heap) ForEach(it func(x *slab, w int) bool) {
	for _, r := range h.data {
		if !it(r.x, r.w) {
			return
		}
	}
}

func (h *heap) update(i int, r recordheap) {
	prev := h.data[i]
	h.data[i] = r
	if !(r.w <= prev.w) {
		h.siftDown(i)
	} else {
		h.siftUp(i)
	}
}

func (h heap) siftDown(root int) {
	for {
		min := root
		for i := 1; i <= h.d; i++ {
			child := h.d*root + i
			if child >= len(h.data) { // out of bounds
				break
			}
			if !(h.data[min].w <= h.data[child].w) {
				min = child
			}
		}
		if min == root {
			return
		}
		a, b := h.data[root], h.data[min]
		h.index[a.x], h.index[b.x] = min, root
		h.data[root], h.data[min] = h.data[min], h.data[root]
		root = min
	}
}

func (h heap) siftUp(root int) {
	for root > 0 {
		parent := (root - 1) / h.d
		if !(h.data[root].w <= h.data[parent].w) {
			return
		}
		a, b := h.data[parent], h.data[root]
		h.index[a.x], h.index[b.x] = root, parent
		h.data[parent], h.data[root] = h.data[root], h.data[parent]
		root = parent
	}
}

func (h heap) siftTop(root int) {
	for root > 0 {
		parent := (root - 1) / h.d
		a, b := h.data[parent], h.data[root]
		h.index[a.x], h.index[b.x] = root, parent
		h.data[parent], h.data[root] = h.data[root], h.data[parent]
		root = parent
	}
}