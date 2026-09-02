package rpc

// Standard JSON-RPC 2.0 error codes.
const (
	ParseError     = -32700 // Invalid JSON was received by the server.
	InvalidRequest = -32600 // The JSON sent is not a valid Request object.
	MethodNotFound = -32601 // The method does not exist / is not available.
	InvalidParams  = -32602 // Invalid method parameter(s).
	InternalError  = -32603 // Internal JSON-RPC error.
)

// Server error reserved range.
const (
	ServerErrorMin = -32099
	ServerErrorMax = -32000
)

// Application-level error codes. Messages in this range (-32000..-31900)
// are host/plugin specific and stable across transports.
const (
	ErrCodePluginNotFound  = -32001
	ErrCodePluginStopped   = -32002
	ErrCodeTimeout         = -32003
	ErrCodeUnknownMethod   = -32004
	ErrCodeOperationFailed = -32005
)

// NewError constructs a JSON-RPC-style error (an *ErrorObject, which also
// implements error) for convenience when returning a plain error value.
func NewError(code int64, message string, data any) error {
	d, err := marshal(data)
	if err != nil {
		d = nil
	}
	return &ErrorObject{
		Code:    code,
		Message: message,
		Data:    d,
	}
}
