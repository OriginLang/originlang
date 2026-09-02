package stdio_test

import (
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"originlang/host/rpc"
	"originlang/host/transport/stdio"
)

func TestStdioSendRead(t *testing.T) {
	// Simulate bidirectional pipe.
	clientIn, serverOut := io.Pipe()
	serverIn, clientOut := io.Pipe()

	tr := stdio.New("test", serverIn, serverOut)
	defer tr.Close()

	// Send a message.
	msg := &rpc.Message{
		Version: "2.0",
		ID:      json.RawMessage(`"m1"`),
		Method:  "test.ping",
	}

	// io.Pipe is synchronous: the writer blocks until the reader consumes.
	// Consume clientIn concurrently so Send does not deadlock.
	type result struct {
		got rpc.Message
		err error
	}
	resCh := make(chan result, 1)
	go func() {
		dec := json.NewDecoder(clientIn)
		var got rpc.Message
		resCh <- result{got: got, err: dec.Decode(&got)}
	}()

	if err := tr.Send(msg); err != nil {
		t.Fatalf("Send: %v", err)
	}

	res := <-resCh
	if res.err != nil {
		t.Fatalf("Decode: %v", res.err)
	}
	if res.got.Method != "test.ping" {
		t.Errorf("method = %q, want test.ping", res.got.Method)
	}
	if string(res.got.ID) != `"m1"` {
		t.Errorf("id = %s, want m1", res.got.ID)
	}

	// Cleanup.
	_ = clientOut.Close()
	_ = clientIn.Close()
}

func TestStdioBidirectional(t *testing.T) {
	// Simulate two endpoints connected via pipes.
	clientIn, serverOut := io.Pipe()
	serverIn, clientOut := io.Pipe()

	serverTr := stdio.New("server", serverIn, serverOut)
	clientTr := stdio.New("client", clientIn, clientOut)

	var wg sync.WaitGroup

	// Server: read request, send response.
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg, err := serverTr.Read()
		if err != nil {
			t.Errorf("server Read: %v", err)
			return
		}
		resp, _ := rpc.NewResponse(msg.ID, map[string]string{"pong": "true"})
		_ = serverTr.Send(resp)
	}()

	// Client: send request, read response.
	wg.Add(1)
	go func() {
		defer wg.Done()
		req, _ := rpc.NewRequest(json.RawMessage(`"bid"`), "echo", nil)
		if err := clientTr.Send(req); err != nil {
			t.Errorf("client Send: %v", err)
			return
		}
		resp, err := clientTr.Read()
		if err != nil {
			t.Errorf("client Read: %v", err)
			return
		}
		if resp.Err != nil {
			t.Errorf("client got error: %v", resp.Err)
		}
		var result struct {
			Pong string `json:"pong"`
		}
		if err := resp.UnmarshalResult(&result); err != nil {
			t.Errorf("UnmarshalResult: %v", err)
			return
		}
		if result.Pong != "true" {
			t.Errorf("pong = %q, want true", result.Pong)
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out")
	}

	_ = serverTr.Close()
	_ = clientTr.Close()
}

func TestStdioKind(t *testing.T) {
	r, w := io.Pipe()
	tr := stdio.New("kind-test", r, w)
	if tr.Kind() != "stdio" {
		t.Errorf("Kind() = %q, want stdio", tr.Kind())
	}
	_ = tr.Close()
}
