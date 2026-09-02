package rpc

import (
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
)

// ID is a JSON-RPC request identifier. JSON-RPC allows a String, Number, or
// Null id. We wrap the raw JSON representation so it round-trips unchanged.
type ID json.RawMessage

// StringID builds an ID from a plain string.
func StringID(s string) ID {
	b, _ := json.Marshal(s)
	return ID(b)
}

// NumID builds an ID from an integer.
func NumID(n int64) ID {
	return ID(strconv.AppendInt(nil, n, 10))
}

// Raw returns the underlying raw JSON bytes.
func (id ID) Raw() json.RawMessage {
	return json.RawMessage(id)
}

// Allocator issues monotonically increasing, unique-in-process IDs.
// It is safe for concurrent use.
type Allocator struct {
	mu     sync.Mutex
	prefix uint64
	next   uint64
}

// NewAllocator returns an ID allocator with a random process-unique prefix.
func NewAllocator() *Allocator {
	a := &Allocator{}
	a.prefix = uint64(newPrefix())
	a.next = 0
	return a
}

// Next returns the next unique ID as a JSON-RPC string id.
func (a *Allocator) Next() ID {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.next++
	// "p:<prefix>-<n>" keeps ids short, sortable, and process-unique.
	return StringID(strconv.FormatUint(a.prefix, 36) + "-" + strconv.FormatUint(a.next, 36))
}

var counter uint64

// newPrefix derives a unique process-wide number.
func newPrefix() uint64 {
	// Mostly a debug aid; uniqueness across processes is best-effort and
	// reinforced by the per-conn allocator ownership model in the transport.
	return atomic.AddUint64(&counter, 1)
}
