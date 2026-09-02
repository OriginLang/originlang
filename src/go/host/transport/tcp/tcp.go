// Package tcp implements the TCP transport. It supports both dialing a
// remote peer and listening for inbound connections, which is the bridge to
// future distributed (HTTP) deployments: a remote plugin can be reachable by
// address rather than as a local child process.
package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"

	"originlang/host/rpc"
	"originlang/host/transport"
)

// Transport is a JSON-RPC transport over a single net.Conn. Messages are
// newline-delimited JSON objects (mirroring the stdio framing).
type Transport struct {
	conn net.Conn
	dec  *json.Decoder

	writeMu sync.Mutex
	sendErr error
	closed  bool
	clMu    sync.Mutex
}

// Dial connects to a remote listener and returns a ready transport.
func Dial(address string, timeout time.Duration) (*Transport, error) {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return newFromConn(conn), nil
}

func newFromConn(conn net.Conn) *Transport {
	return &Transport{
		conn: conn,
		dec:  json.NewDecoder(bufio.NewReader(conn)),
	}
}

// Kind implements transport.Transport.
func (t *Transport) Kind() transport.Kind { return transport.KindTCP }

// Send implements transport.Transport.
func (t *Transport) Send(msg *rpc.Message) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	if t.sendErr != nil {
		return t.sendErr
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if _, err := t.conn.Write(b); err != nil {
		t.sendErr = err
		return err
	}
	return nil
}

// Read implements transport.Transport.
func (t *Transport) Read() (*rpc.Message, error) {
	var msg rpc.Message
	if err := t.dec.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Close implements transport.Transport. It also closes the underlying conn.
func (t *Transport) Close() error {
	t.clMu.Lock()
	defer t.clMu.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	return t.conn.Close()
}

// String implements transport.Transport.
func (t *Transport) String() string { return "tcp:" + t.conn.RemoteAddr().String() }

var _ transport.Transport = (*Transport)(nil)

// Server accepts inbound TCP connections and hands each ready transport to
// onConn. onConn is expected to run its own endpoint/read loop. Server
// returns when ctx is cancelled and the listener is closed.
type Server struct {
	listener net.Listener
	onConn   func(*Transport)
	wg       sync.WaitGroup
}

// Listen starts a TCP server on address and returns immediately. Each
// accepted connection is delivered to onConn once it becomes ready.
func Listen(ctx context.Context, address string, onConn func(*Transport)) (*Server, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	s := &Server{listener: ln, onConn: onConn}
	go s.acceptLoop(ctx)
	return s, nil
}

func (s *Server) acceptLoop(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return
			}
		}
		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.onConn(newFromConn(c))
		}(conn)
	}
}

// Addr returns the bound address.
func (s *Server) Addr() net.Addr { return s.listener.Addr() }

// Close stops accepting and closes all accepted connections.
func (s *Server) Close() error {
	_ = s.listener.Close()
	s.wg.Wait()
	return nil
}

var _ io.Closer = (*Server)(nil)
