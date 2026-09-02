package rpc_test

import (
	"encoding/json"
	"testing"

	"originlang/host/rpc"
)

func TestNewRequest(t *testing.T) {
	id := rpc.NumID(42)
	msg, err := rpc.NewRequest(id.Raw(), "foo.bar", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if msg.Version != "2.0" {
		t.Errorf("version = %q, want 2.0", msg.Version)
	}
	if !msg.IsRequest() {
		t.Error("expected IsRequest()")
	}
	if msg.IsNotification() || msg.IsResponse() || msg.IsError() {
		t.Error("request should not be notification/response/error")
	}
	if string(msg.Method) != "foo.bar" {
		t.Errorf("method = %q, want foo.bar", msg.Method)
	}
	if string(msg.Params) == "" {
		t.Error("params should not be empty")
	}
}

func TestNewNotification(t *testing.T) {
	msg, err := rpc.NewNotification("test.hello", nil)
	if err != nil {
		t.Fatalf("NewNotification: %v", err)
	}
	if !msg.IsNotification() {
		t.Error("expected IsNotification()")
	}
	if msg.IsRequest() || msg.IsResponse() || msg.IsError() {
		t.Error("notification should not be request/response/error")
	}
	if string(msg.ID) != "" {
		t.Errorf("id = %q, want empty", msg.ID)
	}
}

func TestNewResponse(t *testing.T) {
	id := rpc.StringID("abc")
	msg, err := rpc.NewResponse(id.Raw(), "result-value")
	if err != nil {
		t.Fatalf("NewResponse: %v", err)
	}
	if !msg.IsResponse() {
		t.Error("expected IsResponse()")
	}
	if msg.IsRequest() || msg.IsNotification() {
		t.Error("response should not be request/notification")
	}
	if msg.Error() != nil {
		t.Error("response error should be nil")
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := rpc.NumID(1)
	msg, err := rpc.NewErrorResponse(id.Raw(), rpc.MethodNotFound, "not found", nil)
	if err != nil {
		t.Fatalf("NewErrorResponse: %v", err)
	}
	if !msg.IsError() {
		t.Error("expected IsError()")
	}
	if msg.Error() == nil {
		t.Error("error should not be nil")
	}
	if msg.Err.Code != rpc.MethodNotFound {
		t.Errorf("code = %d, want %d", msg.Err.Code, rpc.MethodNotFound)
	}
}

func TestErrorObjectInterface(t *testing.T) {
	e := &rpc.ErrorObject{Code: rpc.InternalError, Message: "oops"}
	s := e.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
}

func TestUnmarshalParams(t *testing.T) {
	type req struct {
		Name string `json:"name"`
	}
	msg := &rpc.Message{Params: json.RawMessage(`{"name":"alice"}`)}
	var out req
	if err := msg.UnmarshalParams(&out); err != nil {
		t.Fatalf("UnmarshalParams: %v", err)
	}
	if out.Name != "alice" {
		t.Errorf("name = %q, want alice", out.Name)
	}
}

func TestUnmarshalParamsNil(t *testing.T) {
	msg := &rpc.Message{}
	var out struct{ Name string }
	if err := msg.UnmarshalParams(&out); err != nil {
		t.Fatalf("UnmarshalParams(nil params): %v", err)
	}
}

func TestUnmarshalResult(t *testing.T) {
	type res struct {
		Value int `json:"value"`
	}
	msg := &rpc.Message{Result: json.RawMessage(`{"value":99}`)}
	var out res
	if err := msg.UnmarshalResult(&out); err != nil {
		t.Fatalf("UnmarshalResult: %v", err)
	}
	if out.Value != 99 {
		t.Errorf("value = %d, want 99", out.Value)
	}
}

func TestAllocatorUniqueness(t *testing.T) {
	a := rpc.NewAllocator()
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := a.Next()
		s := string(id.Raw())
		if seen[s] {
			t.Fatalf("duplicate ID at iteration %d: %s", i, s)
		}
		seen[s] = true
	}
}

func TestAllocatorDifferentProcesses(t *testing.T) {
	a1 := rpc.NewAllocator()
	a2 := rpc.NewAllocator()
	id1 := a1.Next()
	id2 := a2.Next()
	if string(id1.Raw()) == string(id2.Raw()) {
		t.Error("two allocators should produce different IDs")
	}
}

func TestMessageRoundTrip(t *testing.T) {
	orig := &rpc.Message{
		Version: "2.0",
		ID:      json.RawMessage(`"m1"`),
		Method:  "test.echo",
		Params:  json.RawMessage(`{"n":1}`),
	}
	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded rpc.Message
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Method != "test.echo" {
		t.Errorf("method = %q, want test.echo", decoded.Method)
	}
	if !decoded.IsRequest() {
		t.Error("expected IsRequest()")
	}
}
