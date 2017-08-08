# pool

[![GoDoc][godoc-image]][godoc-url]

> Tiny memory reuse helpers for Go.

## generic

Without use of subpackages, `pool` allows to reuse any struct distinguishable
by size in generic way:

```go
package main

import "github.com/gobwas/pool"

func main() {
	x, n := pool.Get(100) // Returns object with size 128 or nil.
	if x == nil {
		// Create x somehow.
	}
	defer pool.Put(x, n)
	
	// Work with x.
}
```

Like `sync.Pool`, generic way allows you to pass initialization function for custom pool:

```go
package main

import "github.com/gobwas/pool"

func main() {
	p := pool.New(0, 128, func(n int) interface{} {
		// Create some object with size n.
	})
	x, n := p.Get(100) // Returns object with size 128.defer pool.Put(x, n)
	defer pool.Put(x, n)
	
	// Work with x.
}
```

Note that there are few non-generic pooling implementations.

## pbytes

Subpackage `pbytes` is intended for `[]byte` reuse.

```go
package main

import "github.com/gobwas/pool/pbytes"

func main() {
	bts := pbytes.GetCap(100) // Returns make([]byte, 0, 128).
	defer pbytes.Put(bts)

	// Work with bts.
}
```

You can also create your own range for pooling:

```go
package main

import "github.com/gobwas/pool/pbytes"

func main() {
	// Reuse only slices whose capacity is 128, 256, 512 or 1024.
	pool := pbytes.New(128, 1024) 

	bts := pool.GetCap(100) // Returns make([]byte, 0, 128).
	defer pool.Put(bts)

	// Work with bts.
}
```

## pbufio

Subpackage `pbufio` is intended for `*bufio.{Reader, Writer}` reuse.

```go
package main

import "github.com/gobwas/pool/pbufio"

func main() {
	bw := pbufio.GetWriter(os.Stdout, 100) // Returns bufio.NewWriterSize(128).
	defer pbufio.PutWriter(bw)

	// Work with bw.
}
```

Like with `pbytes`, you can also create pool with custom reuse bounds.



[godoc-image]: https://godoc.org/github.com/gobwas/pool?status.svg
[godoc-url]:   https://godoc.org/github.com/gobwas/pool
