// Package stdio implements a Transport over stdin/stdout, the default channel
// for local child-process plugins.
package stdio

import (
	"bufio"
	"encoding/json"
	"io"
	"sync"

	"originlang/host/rpc"
	"originlang/host/transport"
)

// Transport is a JSON-RPC transport that reads from in and writes to out
// (typically plugin stdin and stdout). Each message is a single JSON object
// terminated by a newline.
type Transport struct {
	name string
	in   io.Reader
	out  io.Writer

	dec *json.Decoder

	writeMu sync.Mutex
	sendErr error
	closed  bool
	clMu    sync.Mutex
}

// New builds a stdio transport. name is used only for logs.
func New(name string, in io.Reader, out io.Writer) *Transport {
	return &Transport{
		name: name,
		in:   in,
		out:  out,
		dec:  json.NewDecoder(bufio.NewReader(in)),
	}
}

// Kind implements transport.Transport.
func (t *Transport) Kind() transport.Kind { return transport.KindStdio }

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
	if _, err := t.out.Write(b); err != nil {
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

// Close implements transport.Transport.
func (t *Transport) Close() error {
	t.clMu.Lock()
	defer t.clMu.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	if c, ok := t.in.(io.Closer); ok {
		_ = c.Close()
	}
	// Do not close out: stdio streams may be owned by the OS process/exec.
	return nil
}

// String implements transport.Transport.
func (t *Transport) String() string { return "stdio:" + t.name }

var _ transport.Transport = (*Transport)(nil)
