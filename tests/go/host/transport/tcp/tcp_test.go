package tcp_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"originlang/host/rpc"
	"originlang/host/transport/tcp"
)

func TestTCPRoundTrip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var serverTr *tcp.Transport

	received := make(chan *rpc.Message, 1)
	srv, err := tcp.Listen(ctx, "127.0.0.1:0", func(tr *tcp.Transport) {
		mu.Lock()
		serverTr = tr
		mu.Unlock()
		msg, err := tr.Read()
		if err == nil {
			received <- msg
		}
	})
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer srv.Close()

	addr := srv.Addr().String()

	client, err := tcp.Dial(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer client.Close()

	// Wait for server to accept and be available.
	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		ready := serverTr != nil
		mu.Unlock()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server never accepted connection")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Send request from client.
	req, _ := rpc.NewRequest(json.RawMessage(`"tcp-id"`), "echo", map[string]string{"x": "1"})
	if err := client.Send(req); err != nil {
		t.Fatalf("client Send: %v", err)
	}

	select {
	case msg := <-received:
		if msg.Method != "echo" {
			t.Errorf("method = %q, want echo", msg.Method)
		}
		// Reply from server side.
		mu.Lock()
		_ = serverTr.Send(rpcMessage("tcp-reply"))
		mu.Unlock()
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server read")
	}

	// Client should receive the response.
	resp, err := client.Read()
	if err != nil {
		t.Fatalf("client Read: %v", err)
	}
	if string(resp.Result) == "" {
		t.Error("expected non-empty result")
	}
}

func rpcMessage(id string) *rpc.Message {
	return &rpc.Message{
		Version: "2.0",
		ID:      json.RawMessage(`"` + id + `"`),
		Result:  json.RawMessage(`{"ok":true}`),
	}
}

func TestTCPKind(t *testing.T) {
	srv, err := tcp.Listen(context.Background(), "127.0.0.1:0", func(*tcp.Transport) {})
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer srv.Close()

	conn, err := tcp.Dial(srv.Addr().String(), time.Second)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()
	if conn.Kind() != "tcp" {
		t.Errorf("Kind() = %q, want tcp", conn.Kind())
	}
}
