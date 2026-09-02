package transport

import (
	"context"
	"sync"

	"originlang/host/rpc"
)

// Handler processes an incoming request or notification for a registered
// method. It returns the result value to be marshalled into a response, or a
// non-nil error to produce an error response. For notifications the result is
// ignored.
type Handler func(ctx context.Context, params *rpc.Message) (result any, err error)

// Endpoint wraps a Transport and provides the message-level session logic
// shared by host kernel and plugin: method registration, request/response
// correlation, and notifications.
//
// A single reader goroutine consumes messages off the transport while any
// number of goroutines may issue concurrent calls.
type Endpoint struct {
	tr Transport

	mu        sync.Mutex
	handlers  map[string]Handler
	pending   map[string]chan *rpc.Message
	closed    bool
	closeOnce sync.Once
	closeCh   chan struct{}
}

// NewEndpoint wraps the given transport with a fresh session. Callers must
// invoke Run to start consuming messages.
func NewEndpoint(tr Transport) *Endpoint {
	return &Endpoint{
		tr:       tr,
		handlers: make(map[string]Handler),
		pending:  make(map[string]chan *rpc.Message),
		closeCh:  make(chan struct{}),
	}
}

// Transport returns the underlying transport.
func (e *Endpoint) Transport() Transport { return e.tr }

// Register registers a handler for the given JSON-RPC method.
func (e *Endpoint) Register(method string, h Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[method] = h
}

// Run consumes messages until the transport is closed or ctx is cancelled.
// It blocks; call it in its own goroutine.
func (e *Endpoint) Run(ctx context.Context) {
	for {
		msg, err := e.tr.Read()
		if err != nil {
			e.shutdown()
			return
		}
		if err := e.dispatch(ctx, msg); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
}

func (e *Endpoint) dispatch(ctx context.Context, msg *rpc.Message) error {
	switch {
	case msg.IsResponse():
		return e.resolveResponse(msg)
	case msg.IsRequest():
		e.handleRequest(ctx, msg)
	case msg.IsNotification():
		e.handleNotification(ctx, msg)
	}
	return nil
}

// resolveResponse matches an inbound response to a waiting Call by ID.
func (e *Endpoint) resolveResponse(msg *rpc.Message) error {
	k := string(msg.ID)
	e.mu.Lock()
	ch, ok := e.pending[k]
	if ok {
		delete(e.pending, k)
	}
	e.mu.Unlock()
	if ch != nil {
		ch <- msg
	}
	return nil
}

func (e *Endpoint) handleRequest(ctx context.Context, msg *rpc.Message) {
	h, ok := e.handlerFor(msg.Method)
	if !ok {
		resp, _ := rpc.NewErrorResponse(msg.ID, rpc.MethodNotFound, "method not found: "+msg.Method, nil)
		_ = e.tr.Send(resp)
		return
	}
	result, err := h(ctx, msg)
	if err != nil {
		resp, _ := rpc.NewErrorResponse(msg.ID, rpc.InternalError, err.Error(), nil)
		_ = e.tr.Send(resp)
		return
	}
	resp, err := rpc.NewResponse(msg.ID, result)
	if err != nil {
		resp, _ = rpc.NewErrorResponse(msg.ID, rpc.InternalError, err.Error(), nil)
	}
	_ = e.tr.Send(resp)
}

func (e *Endpoint) handleNotification(ctx context.Context, msg *rpc.Message) {
	if h, ok := e.handlerFor(msg.Method); ok {
		_, _ = h(ctx, msg)
	}
}

func (e *Endpoint) handlerFor(method string) (Handler, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	h, ok := e.handlers[method]
	return h, ok
}

// Call sends a request and blocks until the matching response arrives, the
// context is cancelled, or the transport closes.
func (e *Endpoint) Call(ctx context.Context, id rpc.ID, method string, params any) (*rpc.Message, error) {
	req, err := rpc.NewRequest(id.Raw(), method, params)
	if err != nil {
		return nil, err
	}
	k := string(id.Raw())

	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return nil, rpc.NewError(rpc.ErrCodePluginStopped, "endpoint closed", nil)
	}
	ch := make(chan *rpc.Message, 1)
	e.pending[k] = ch
	e.mu.Unlock()

	if err := e.tr.Send(req); err != nil {
		e.mu.Lock()
		delete(e.pending, k)
		e.mu.Unlock()
		return nil, err
	}

	select {
	case <-ctx.Done():
		e.mu.Lock()
		delete(e.pending, k)
		e.mu.Unlock()
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.IsError() {
			return resp, resp.Error()
		}
		return resp, nil
	case <-e.closeCh:
		return nil, rpc.NewError(rpc.ErrCodePluginStopped, "endpoint closed during call", nil)
	}
}

// Notify sends a one-way notification.
func (e *Endpoint) Notify(method string, params any) error {
	msg, err := rpc.NewNotification(method, params)
	if err != nil {
		return err
	}
	return e.tr.Send(msg)
}

// Close shuts down the endpoint and its transport.
func (e *Endpoint) Close() error {
	e.closeOnce.Do(func() {
		e.mu.Lock()
		e.closed = true
		e.mu.Unlock()
		close(e.closeCh)
		_ = e.tr.Close()
	})
	return nil
}

func (e *Endpoint) shutdown() {
	e.closeOnce.Do(func() {
		e.mu.Lock()
		e.closed = true
		e.mu.Unlock()
		close(e.closeCh)
	})
}
