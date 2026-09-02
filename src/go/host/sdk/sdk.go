// Package sdk provides the plugin-side companion to the host kernel. A plugin
// process calls Serve to establish the stdio transport, answer the register
// handshake, and respond to host-driven lifecycle calls (ping, shutdown).
//
// The same protocol is used by plugins written in other languages; this
// package just removes the boilerplate for Go plugins.
package sdk

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"originlang/host/plugin"
	"originlang/host/rpc"
	"originlang/host/transport"
	"originlang/host/transport/stdio"
)

// Options configures a plugin session.
type Options struct {
	// Name/Version are reported to the host during registration.
	Name    string
	Version string
	// Capabilities are the RPC methods this plugin serves.
	Capabilities []plugin.Capability
	// OnRegister, when set, is invoked after successful registration.
	OnRegister func()
	// ErrorOutput receives any internal error messages (defaults to os.Stderr).
	ErrorOutput io.Writer
}

// Session is a running plugin session bound to the host.
type Session struct {
	ep *transport.Endpoint
}

// Endpoint exposes the RPC endpoint for callbacks and custom handlers.
func (s *Session) Endpoint() *transport.Endpoint { return s.ep }

// Serve runs the plugin session until the host sends shutdown, the context is
// cancelled, or an OS interrupt is received. It blocks.
func Serve(ctx context.Context, opts Options) error {
	if opts.ErrorOutput == nil {
		opts.ErrorOutput = os.Stderr
	}
	tr := stdio.New("sdk", os.Stdin, os.Stdout)
	ep := transport.NewEndpoint(tr)

	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var once sync.Once
	notifyRegister := func() {
		if opts.OnRegister != nil {
			once.Do(opts.OnRegister)
		}
	}

	ep.Register(plugin.MethodRegister, func(_ context.Context, _ *rpc.Message) (any, error) {
		notifyRegister()
		return plugin.RegisterReply{
			Name:         opts.Name,
			Version:      opts.Version,
			Capabilities: opts.Capabilities,
		}, nil
	})
	ep.Register(plugin.MethodPing, func(_ context.Context, _ *rpc.Message) (any, error) {
		return map[string]any{"pong": true, "name": opts.Name}, nil
	})
	ep.Register(plugin.MethodShutdown, func(_ context.Context, _ *rpc.Message) (any, error) {
		cancel()
		return nil, nil
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	go func() {
		select {
		case <-stop:
			cancel()
		case <-reqCtx.Done():
		}
	}()

	go ep.Run(reqCtx)

	<-reqCtx.Done()
	_ = tr.Close()
	return nil
}
