// Package transport defines the pluggable transport abstraction used for all
// host<->plugin messaging.
//
// A Transport is responsible for moving JSON-RPC messages over some medium
// (stdio, TCP, or in the future HTTP). Transports are symmetric: the same
// interface is implemented by the host kernel and by plugins, which is what
// lets a plugin run either as a local child process (stdio/TCP) or as a
// remote service (HTTP) without changing the host's plugin-manager logic.
package transport

import (
	"io"

	"originlang/host/rpc"
)

// Kind enumerates the supported (or planned) transport kinds.
type Kind string

const (
	// KindStdio is a local child-process transport over stdin/stdout.
	KindStdio Kind = "stdio"
	// KindTCP is a TCP socket transport, suitable for local or remote peers.
	KindTCP Kind = "tcp"
	// KindHTTP is reserved for the future HTTP transport enabling distributed
	// plugin deployments on top of the same message protocol.
	KindHTTP Kind = "http"
)

// Transport moves encoded JSON-RPC messages over a medium.
//
// Implementations must be safe for concurrent Send calls and may return a
// non-nil io.EOF from Read when the peer closes the connection cleanly.
type Transport interface {
	// Kind identifies the transport medium.
	Kind() Kind
	// Send writes a message to the peer. It must be safe for concurrent use.
	Send(msg *rpc.Message) error
	// Read blocks until the next full message is available or the transport
	// is closed. It returns io.EOF on a clean peer shutdown.
	Read() (*rpc.Message, error)
	// Close tears down the underlying medium.
	Close() error
	// String returns a human-readable description for logs.
	String() string
}

// Codec is the message framing strategy for a transport. Most transports can
// share the newline-delimited JSON codec; TCP framing may stack a length
// prefix on top.
type Codec interface {
	io.ReadWriteCloser
}
