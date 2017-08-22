#include "ppgo/struct/indexed_heap.h"

#define LESS_OR_EQUAL(a, b) a <= b
#define STRUCT(a) heap
#define CTOR() newSlabHeap
#define EMPTY() nil 
#define COMPARE(a, b) a - b 

package slab

MAKE_INDEXED_HEAP(*slab, int)
