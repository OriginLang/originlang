package transport_test

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"originlang/host/rpc"
	"originlang/host/transport"
)

// fakeTransport is a test transport backed by in-memory channels.
type fakeTransport struct {
	sendCh chan *rpc.Message
	recvCh chan *rpc.Message
	closed bool
	mu     sync.Mutex
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{
		sendCh: make(chan *rpc.Message, 64),
		recvCh: make(chan *rpc.Message, 64),
	}
}

func (f *fakeTransport) Kind() transport.Kind { return transport.Kind("fake") }

func (f *fakeTransport) Send(msg *rpc.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return io.EOF
	}
	f.sendCh <- msg
	return nil
}

func (f *fakeTransport) Read() (*rpc.Message, error) {
	msg, ok := <-f.recvCh
	if !ok {
		return nil, io.EOF
	}
	return msg, nil
}

func (f *fakeTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	close(f.recvCh)
	return nil
}

func (f *fakeTransport) String() string { return "fake" }

func TestEndpointRequestResponse(t *testing.T) {
	ft := newFakeTransport()
	ep := transport.NewEndpoint(ft)
	defer ep.Close()

	ep.Register("echo", func(_ context.Context, m *rpc.Message) (any, error) {
		var p struct {
			Msg string `json:"msg"`
		}
		_ = m.UnmarshalParams(&p)
		return map[string]string{"reply": p.Msg}, nil
	})

	ctx := context.Background()
	go ep.Run(ctx)

	// Inject a request; endpoint should produce a response.
	req, _ := rpc.NewRequest(json.RawMessage(`"test-id-1"`), "echo", map[string]string{"msg": "hello"})
	ft.recvCh <- req

	select {
	case resp := <-ft.sendCh:
		if resp.Err != nil {
			t.Fatalf("unexpected error: %v", resp.Err)
		}
		var result struct {
			Reply string `json:"reply"`
		}
		if err := resp.UnmarshalResult(&result); err != nil {
			t.Fatalf("UnmarshalResult: %v", err)
		}
		if result.Reply != "hello" {
			t.Errorf("reply = %q, want hello", result.Reply)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for response")
	}
}

func TestEndpointNotification(t *testing.T) {
	ft := newFakeTransport()
	ep := transport.NewEndpoint(ft)
	defer ep.Close()

	received := make(chan string, 1)
	ep.Register("notify.me", func(_ context.Context, m *rpc.Message) (any, error) {
		var p struct{ Val string `json:"val"` }
		_ = m.UnmarshalParams(&p)
		received <- p.Val
		return nil, nil
	})

	ctx := context.Background()
	go ep.Run(ctx)

	n, _ := rpc.NewNotification("notify.me", map[string]string{"val": "got-it"})
	ft.recvCh <- n

	select {
	case v := <-received:
		if v != "got-it" {
			t.Errorf("val = %q, want got-it", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for notification")
	}
}

func TestEndpointMethodNotFound(t *testing.T) {
	ft := newFakeTransport()
	ep := transport.NewEndpoint(ft)
	defer ep.Close()

	ctx := context.Background()
	go ep.Run(ctx)

	req, _ := rpc.NewRequest(json.RawMessage(`"id-mnf"`), "nonexistent.method", nil)
	ft.recvCh <- req

	select {
	case resp := <-ft.sendCh:
		if resp.Err == nil {
			t.Fatal("expected error response for unknown method")
		}
		if resp.Err.Code != rpc.MethodNotFound {
			t.Errorf("code = %d, want %d", resp.Err.Code, rpc.MethodNotFound)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}

func TestEndpointConcurrentCalls(t *testing.T) {
	ft := newFakeTransport()
	ep := transport.NewEndpoint(ft)
	defer ep.Close()

	ep.Register("slow.echo", func(_ context.Context, m *rpc.Message) (any, error) {
		var p struct{ Msg string `json:"msg"` }
		_ = m.UnmarshalParams(&p)
		return map[string]string{"reply": p.Msg}, nil
	})

	ctx := context.Background()
	go ep.Run(ctx)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := rpc.NumID(int64(n))
			req, _ := rpc.NewRequest(id.Raw(), "slow.echo", map[string]int{"n": n})
			ft.recvCh <- req

			select {
			case resp := <-ft.sendCh:
				if resp.Err != nil {
					t.Errorf("call %d: error: %v", n, resp.Err)
				}
			case <-time.After(2 * time.Second):
				t.Errorf("call %d: timed out", n)
			}
		}(i)
	}
	wg.Wait()
}
