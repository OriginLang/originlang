// Package manager is the host-kernel core: it loads, registers, runs, looks
// up, and tears down plugin instances.
//
// A plugin is presented to the rest of the host as a named, capability-
// exposing RPC peer. Whether the plugin is a local child process speaking
// stdio, or a remote service reached over TCP (and later HTTP), the manager
// treats it identically once its transport is established.
package manager

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"originlang/host/plugin"
	"originlang/host/transport"
	"originlang/host/transport/stdio"
)

// Logger receives lifecycle events. It is kept minimal and nil-safe so the
// manager has no hard dependency on a logging framework.
type Logger interface {
	Logf(pluginID, format string, args ...any)
}

// NopLogger discards all log output.
type NopLogger struct{}

// Logf implements Logger.
func (NopLogger) Logf(string, string, ...any) {}

// Options configures the Manager.
type Options struct {
	// Logger is optional; when nil a NopLogger is used.
	Logger Logger
	// RegisterTimeout bounds the plugin.register handshake.
	RegisterTimeout time.Duration
	// StartupGrace bounds process wait time on load.
	StartupGrace time.Duration
}

// Manager owns the set of loaded plugins.
type Manager struct {
	log Logger
	registerTimeout time.Duration

	mu      sync.Mutex
	plugins map[string]*plugin.Plugin
	order   []string
}

// NewManager returns an empty Manager.
func NewManager() *Manager {
	return NewManagerWith(Options{})
}

// NewManagerWith returns a Manager configured with opts.
func NewManagerWith(opts Options) *Manager {
	if opts.Logger == nil {
		opts.Logger = NopLogger{}
	}
	if opts.RegisterTimeout == 0 {
		opts.RegisterTimeout = 10 * time.Second
	}
	if opts.StartupGrace == 0 {
		opts.StartupGrace = 5 * time.Second
	}
	return &Manager{
		log:             opts.Logger,
		registerTimeout: opts.RegisterTimeout,
		plugins:         make(map[string]*plugin.Plugin),
	}
}

// Attach registers a plugin whose transport is already established (a remote
// plugin reached over TCP, or any externally connected endpoint) and drives
// its register handshake.
func (m *Manager) Attach(ctx context.Context, spec plugin.Spec, ep *transport.Endpoint) (*plugin.Plugin, error) {
	p := plugin.New(spec)
	p.SetEndpoint(ep)
	// Run the inbound read loop.
	go ep.Run(ctx)
	return m.adopt(ctx, p, ep)
}

// Load launches a child-process plugin over stdio and completes the register
// handshake. spec.AttachEndpoint, if set, takes precedence and lets Load act
// like Attach (useful for test doubles).
func (m *Manager) Load(ctx context.Context, spec plugin.Spec) (*plugin.Plugin, error) {
	if spec.AttachEndpoint != nil {
		p := plugin.New(spec)
		p.SetEndpoint(spec.AttachEndpoint)
		go spec.AttachEndpoint.Run(ctx)
		return m.handshake(ctx, p)
	}

	p := plugin.New(spec)
	p.SetState(plugin.StateStarting)

	cmd := exec.CommandContext(ctx, spec.Path, spec.Args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	p.SetProcess(cmd.Process)

	tr := stdio.New(spec.Name, stdout, stdin)
	ep := transport.NewEndpoint(tr)
	p.SetEndpoint(ep)
	go ep.Run(ctx)

	// Release the process handles asynchronously so we don't leak waiters.
	go func() {
		_ = cmd.Wait()
	}()

	return m.handshake(ctx, p)
}

func (m *Manager) adopt(ctx context.Context, p *plugin.Plugin, _ *transport.Endpoint) (*plugin.Plugin, error) {
	// Remote/attached plugins still complete a handshake.
	return m.handshake(ctx, p)
}

// handshake completes the register exchange and stores the plugin.
func (m *Manager) handshake(ctx context.Context, p *plugin.Plugin) (*plugin.Plugin, error) {
	regCtx, cancel := context.WithTimeout(ctx, m.registerTimeout)
	defer cancel()

	msg, err := p.Call(regCtx, plugin.MethodRegister, map[string]any{
		"id":   p.ID,
		"name": p.Spec.Name,
	})
	if err != nil {
		p.SetState(plugin.StateFailed)
		_ = p.Endpoint().Close()
		return nil, err
	}

	var reg plugin.RegisterReply
	if err := msg.UnmarshalResult(&reg); err != nil {
		p.SetState(plugin.StateFailed)
		_ = p.Endpoint().Close()
		return nil, err
	}
	if reg.Name == "" {
		reg.Name = p.Spec.Name
	}
	p.Register(&reg)
	p.SetState(plugin.StateReady)

	m.mu.Lock()
	m.plugins[p.ID] = p
	m.order = append(m.order, p.ID)
	m.mu.Unlock()

	m.log.Logf(p.ID, "plugin ready with %d capabilities", len(reg.Capabilities))
	return p, nil
}

// Get returns a plugin by ID, or an error if not found.
func (m *Manager) Get(id string) (*plugin.Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.plugins[id]
	if !ok {
		return nil, errNotFound(id)
	}
	return p, nil
}

// List returns all loaded plugins in load order.
func (m *Manager) List() []*plugin.Plugin {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*plugin.Plugin, 0, len(m.order))
	for _, id := range m.order {
		if p, ok := m.plugins[id]; ok {
			out = append(out, p)
		}
	}
	return out
}

// Call dispatches an RPC to the named plugin.
func (m *Manager) Call(ctx context.Context, pluginID, method string, params any) (any, error) {
	p, err := m.Get(pluginID)
	if err != nil {
		return nil, err
	}
	msg, err := p.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}
	// Return the raw result bytes; the caller decides how to decode.
	return msg.Result, nil
}

// Shutdown stops a single plugin.
func (m *Manager) Shutdown(ctx context.Context, id string) error {
	p, err := m.Get(id)
	if err != nil {
		return err
	}
	if proc := p.Process(); proc != nil {
		// Give the plugin a moment to drain; then ask it to stop.
		_ = p.Shutdown(ctx)
		m.forceKill(proc)
	} else {
		_ = p.Shutdown(ctx)
	}
	m.mu.Lock()
	delete(m.plugins, id)
	m.mu.Unlock()
	m.log.Logf(id, "plugin stopped")
	return nil
}

// StopAll gracefully shuts down all plugins.
func (m *Manager) StopAll(ctx context.Context) {
	m.mu.Lock()
	ids := make([]string, len(m.order))
	copy(ids, m.order)
	m.mu.Unlock()
	for _, id := range ids {
		_ = m.Shutdown(ctx, id)
	}
}

func (m *Manager) forceKill(proc *os.Process) {
	done := make(chan struct{})
	go func() {
		_, _ = proc.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = proc.Kill()
	}
}

// Close releases all plugins; any remaining processes are killed.
func (m *Manager) Close() error {
	m.mu.Lock()
	plugs := make([]*plugin.Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugs = append(plugs, p)
	}
	m.plugins = make(map[string]*plugin.Plugin)
	m.order = nil
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, p := range plugs {
		if proc := p.Process(); proc != nil {
			_ = p.Shutdown(ctx)
			m.forceKill(proc)
		} else {
			_ = p.Shutdown(ctx)
		}
	}
	return nil
}

var _ io.Closer = (*Manager)(nil)
