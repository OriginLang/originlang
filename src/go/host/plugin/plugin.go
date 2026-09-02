// Package plugin defines the host-side abstraction of a loaded plugin: its
// identity, capabilities, lifecycle state, and the RPC surface used to drive
// it.
package plugin

import (
	"context"
	"os"
	"sync"

	"originlang/host/rpc"
	"originlang/host/transport"
)

// Well-known RPC methods the host sends to a plugin.
const (
	MethodRegister = "plugin.register"
	MethodShutdown = "plugin.shutdown"
	MethodPing     = "plugin.ping"
)

// State is the lifecycle state of a plugin instance.
type State int

const (
	// StateCreated is the initial state after construction.
	StateCreated State = iota
	// StateStarting is set while the process/connection is being brought up.
	StateStarting
	// StateRegistered is set after the plugin has reported its capabilities.
	StateRegistered
	// StateReady is set when the plugin signals it is ready to serve.
	StateReady
	// StateRunning is the steady state while serving RPCs.
	StateRunning
	// StateStopped is set after a clean shutdown.
	StateStopped
	// StateFailed is set when startup or a crash leaves the plugin unusable.
	StateFailed
)

// String renders a State.
func (s State) String() string {
	switch s {
	case StateCreated:
		return "created"
	case StateStarting:
		return "starting"
	case StateRegistered:
		return "registered"
	case StateReady:
		return "ready"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Capability declares a single RPC method a plugin serves.
type Capability struct {
	Method      string `json:"method"`
	Description string `json:"description,omitempty"`
}

// RegisterReply is the payload of the plugin.register handshake.
type RegisterReply struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Capabilities []Capability `json:"capabilities"`
}

// Plugin is a host-side handle to a loaded plugin instance.
type Plugin struct {
	ID    string
	Spec  Spec
	State State

	endpoint *transport.Endpoint
	idgen    *rpc.Allocator

	mu     sync.Mutex
	caps   []Capability
	reg    *RegisterReply
	proc   *os.Process
	closed bool
}

// Spec describes how to launch a plugin.
type Spec struct {
	// ID is a stable unique identifier.
	ID string
	// Name is a human-readable name.
	Name string
	// Path is the executable/script path for child-process plugins, or an
	// empty string for a pre-connected remote plugin.
	Path string
	// Args are passed to the plugin executable.
	Args []string
	// AttachEndpoint, when non-nil, pre-wires an already-established
	// transport (e.g. a TCP connection to a remote plugin) instead of
	// launching a child process.
	AttachEndpoint *transport.Endpoint
}

// New constructs a Plugin in StateCreated.
func New(spec Spec) *Plugin {
	ep := spec.AttachEndpoint
	if ep == nil {
		ep = nil
	}
	return &Plugin{
		ID:       spec.ID,
		Spec:     spec,
		State:    StateCreated,
		endpoint: ep,
		idgen:    rpc.NewAllocator(),
	}
}

// Endpoint returns the RPC endpoint for this plugin.
func (p *Plugin) Endpoint() *transport.Endpoint { return p.endpoint }

// SetEndpoint wires an endpoint after construction (used when a child process
// is launched and its stdio transport becomes available).
func (p *Plugin) SetEndpoint(ep *transport.Endpoint) {
	p.mu.Lock()
	p.endpoint = ep
	p.mu.Unlock()
}

// SetProcess records the child process handle.
func (p *Plugin) SetProcess(proc *os.Process) {
	p.mu.Lock()
	p.proc = proc
	p.mu.Unlock()
}

// Process returns the child process handle, or nil for an attached plugin.
func (p *Plugin) Process() *os.Process {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.proc
}

// SetState transitions the plugin to the given state.
func (p *Plugin) SetState(s State) {
	p.mu.Lock()
	p.State = s
	p.mu.Unlock()
}

// StateValue returns the current state.
func (p *Plugin) StateValue() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.State
}

// Register records the handshake result.
func (p *Plugin) Register(reg *RegisterReply) {
	p.mu.Lock()
	p.reg = reg
	p.caps = reg.Capabilities
	p.State = StateRegistered
	p.mu.Unlock()
}

// Capabilities returns the plugin's declared capabilities.
func (p *Plugin) Capabilities() []Capability {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Capability, len(p.caps))
	copy(out, p.caps)
	return out
}

// Call invokes an arbitrary RPC method on the plugin.
func (p *Plugin) Call(ctx context.Context, method string, params any) (*rpc.Message, error) {
	ep := p.endpoint
	if ep == nil {
		return nil, rpc.NewError(rpc.ErrCodePluginStopped, "plugin has no endpoint", nil)
	}
	id := p.newID()
	return ep.Call(ctx, id, method, params)
}

// Shutdown asks the plugin to stop gracefully and marks it stopped.
func (p *Plugin) Shutdown(ctx context.Context) error {
	ep := p.endpoint
	if ep == nil {
		return nil
	}
	_ = ep.Notify(MethodShutdown, nil)
	_ = ep.Close()
	p.SetState(StateStopped)
	return nil
}

func (p *Plugin) newID() rpc.ID {
	// The endpoint's Call owns correlation by raw id string; we just need a
	// per-connection unique id, which the allocator provides.
	return p.idgen.Next()
}