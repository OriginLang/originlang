// Package rpc implements the JSON-RPC 2.0 wire protocol used for all
// host<->plugin communication.
//
// It is deliberately dependency-free (stdlib only) so that both the host
// kernel and plugins built in any language can speak the same protocol, and
// so the same messages may flow over stdio, TCP, or (future) HTTP transports
// for distributed deployments.
package rpc

import (
	"encoding/json"
	"fmt"
)

// Version is the JSON-RPC protocol version identifier.
const Version = "2.0"

// Message is the generic envelope for every JSON-RPC 2.0 message: a request,
// a notification, a response, or an error response.
//
// Fields are kept as json.RawMessage so that params/result payloads are
// forwarded without lossy re-marshalling across transports.
type Message struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Err     *ErrorObject    `json:"error,omitempty"`
}

// ErrorObject is the JSON-RPC error object carried in a Message.Err.
type ErrorObject struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewRequest builds a request message with the given ID.
func NewRequest(id json.RawMessage, method string, params any) (*Message, error) {
	p, err := marshal(params)
	if err != nil {
		return nil, err
	}
	return &Message{
		Version: Version,
		ID:      id,
		Method:  method,
		Params:  p,
	}, nil
}

// NewNotification builds a notification (no ID) message.
func NewNotification(method string, params any) (*Message, error) {
	p, err := marshal(params)
	if err != nil {
		return nil, err
	}
	return &Message{
		Version: Version,
		Method:  method,
		Params:  p,
	}, nil
}

// NewResponse builds a successful response for the given request ID.
func NewResponse(id json.RawMessage, result any) (*Message, error) {
	r, err := marshal(result)
	if err != nil {
		return nil, err
	}
	return &Message{
		Version: Version,
		ID:      id,
		Result:  r,
	}, nil
}

// NewErrorResponse builds an error response for the given request ID.
func NewErrorResponse(id json.RawMessage, code int64, message string, data any) (*Message, error) {
	d, err := marshal(data)
	if err != nil {
		return nil, err
	}
	return &Message{
		Version: Version,
		ID:      id,
		Err: &ErrorObject{
			Code:    code,
			Message: message,
			Data:    d,
		},
	}, nil
}

// IsRequest reports whether the message is a request (has an ID and a method).
func (m *Message) IsRequest() bool {
	return m.Method != "" && len(m.ID) > 0
}

// IsNotification reports whether the message is a notification (method, no ID).
func (m *Message) IsNotification() bool {
	return m.Method != "" && len(m.ID) == 0
}

// IsResponse reports whether the message is a response (has an ID, no method).
func (m *Message) IsResponse() bool {
	return m.Method == "" && len(m.ID) > 0
}

// IsError reports whether the message carries an error object.
func (m *Message) IsError() bool {
	return m.Err != nil
}

// UnmarshalParams decodes the params into out using json.RawMessage passthrough.
// When params is null or empty, it leaves out untouched and returns nil.
func (m *Message) UnmarshalParams(out any) error {
	if len(m.Params) == 0 || string(m.Params) == "null" {
		return nil
	}
	return json.Unmarshal(m.Params, out)
}

// UnmarshalResult decodes the result into out.
func (m *Message) UnmarshalResult(out any) error {
	return json.Unmarshal(m.Result, out)
}

// Error returns a Go error describing the response error, or nil on success.
func (m *Message) Error() error {
	if m.Err == nil {
		return nil
	}
	return m.Err
}

func marshal(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// Error implements error for ErrorObject.
func (e *ErrorObject) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "rpc error"
	}
	return fmt.Sprintf("jsonrpc %d: %s", e.Code, msg)
}
