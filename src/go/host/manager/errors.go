package manager

import (
	"fmt"

	"originlang/host/rpc"
)

// NotFoundError is returned when a plugin ID is not loaded.
type NotFoundError struct {
	ID string
}

// Error implements error.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("plugin %q not found", e.ID)
}

// Is reports whether the target error is a NotFoundError.
func (e *NotFoundError) Is(target error) bool {
	_, ok := target.(*NotFoundError)
	return ok
}

func errNotFound(id string) error {
	return &NotFoundError{ID: id}
}

var _ error = (*NotFoundError)(nil)

// toRPCError normalizes a manager error into a JSON-RPC error code for
// forwarding across a transport boundary.
func toRPCError(err error) *rpc.ErrorObject {
	if err == nil {
		return nil
	}
	e := &rpc.ErrorObject{
		Code:    rpc.ErrCodeOperationFailed,
		Message: err.Error(),
	}
	var nfe *NotFoundError
	if asError(err, &nfe) {
		e.Code = rpc.ErrCodePluginNotFound
	}
	return e
}

func asError(err error, target any) bool {
	v := err
	for v != nil {
		switch t := target.(type) {
		case **NotFoundError:
			if nf, ok := v.(*NotFoundError); ok {
				*t = nf
				return true
			}
		}
		u, ok := v.(interface{ Unwrap() error })
		if !ok {
			break
		}
		v = u.Unwrap()
	}
	return false
}
