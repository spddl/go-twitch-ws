package twitch

import (
	"bytes"
	"sync"
)

// stolen from https://github.com/nhooyr/websocket/blob/master/internal/bpool/bpool.go
var bpool sync.Pool // https://golang.org/pkg/sync/#Pool

// Get returns a buffer from the pool or creates a new one if
// the pool is empty.
func bpoolGet() *bytes.Buffer {
	b := bpool.Get()
	if b == nil {
		return &bytes.Buffer{}
	}
	return b.(*bytes.Buffer)
}

// Put returns a buffer into the pool.
func bpoolPut(b *bytes.Buffer) {
	b.Reset()
	bpool.Put(b)
}
