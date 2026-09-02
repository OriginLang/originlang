// Command hostd is the host daemon kernel: it loads configured plugins,
// drives their lifecycle, and exposes the host RPC surface so that external
// components can query and invoke plugins.
//
// By default hostd speaks JSON-RPC over stdio (for embedding / piping). With
// -tcp it instead serves over a TCP listener, which is the stepping stone to
// the future distributed (HTTP) deployment: the same plugin manager backs
// both mediums.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"originlang/host/manager"
	"originlang/host/plugin"
	"originlang/host/rpc"
	"originlang/host/transport"
	"originlang/host/transport/stdio"
	"originlang/host/transport/tcp"
)

// hostAPI implements the methods an external client calls on the host.
type hostAPI struct {
	mgr *manager.Manager
}

// initEndpoint registers the host's public RPC methods on an endpoint.
func (a *hostAPI) initEndpoint(ep *transport.Endpoint) {
	ep.Register("host.plugin.list", a.handleList)
	ep.Register("host.plugin.call", a.handleCall)
}

func (a *hostAPI) handleList(_ context.Context, m *rpc.Message) (any, error) {
	_ = m
	plugins := a.mgr.List()
	type item struct {
		ID           string              `json:"id"`
		Name         string              `json:"name"`
		State        string              `json:"state"`
		Capabilities []plugin.Capability `json:"capabilities"`
	}
	res := make([]item, 0, len(plugins))
	for _, p := range plugins {
		res = append(res, item{
			ID:           p.ID,
			Name:         p.Spec.Name,
			State:        p.StateValue().String(),
			Capabilities: p.Capabilities(),
		})
	}
	return res, nil
}

func (a *hostAPI) handleCall(ctx context.Context, m *rpc.Message) (any, error) {
	var req struct {
		PluginID string          `json:"plugin_id"`
		Method   string          `json:"method"`
		Params   json.RawMessage `json:"params"`
	}
	if err := m.UnmarshalParams(&req); err != nil {
		return nil, err
	}
	p, err := a.mgr.Get(req.PluginID)
	if err != nil {
		return nil, err
	}
	msg, err := p.Call(ctx, req.Method, unpack(req.Params))
	if err != nil {
		return nil, err
	}
	return msg.Result, nil
}

func unpack(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return v
}

func main() {
	var (
		tcpAddr    string
		pluginsCSV string
	)
	flag.StringVar(&tcpAddr, "tcp", "", "expose host over TCP on this address (e.g. ':7600'); empty means stdio")
	flag.StringVar(&pluginsCSV, "plugins", "", "comma-separated plugin executables to load")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mgr := manager.NewManagerWith(manager.Options{Logger: stdLogger{}})
	defer mgr.Close()

	for _, p := range parsePlugins(pluginsCSV) {
		if _, err := mgr.Load(ctx, p); err != nil {
			log.Fatalf("load plugin %q: %v", p.ID, err)
		}
	}

	api := &hostAPI{mgr: mgr}

	if tcpAddr != "" {
		if err := serveTCP(ctx, tcpAddr, api); err != nil {
			log.Fatalf("tcp serve: %v", err)
		}
		return
	}
	serveStdio(ctx, api)
}

func serveStdio(ctx context.Context, api *hostAPI) {
	tr := stdio.New("hostd", os.Stdin, os.Stdout)
	ep := transport.NewEndpoint(tr)
	api.initEndpoint(ep)
	go ep.Run(ctx)
	<-ctx.Done()
	_ = ep.Close()
}

func serveTCP(ctx context.Context, addr string, api *hostAPI) error {
	onConn := func(tr *tcp.Transport) {
		ep := transport.NewEndpoint(tr)
		api.initEndpoint(ep)
		go ep.Run(ctx)
	}
	srv, err := tcp.Listen(ctx, addr, onConn)
	if err != nil {
		return err
	}
	log.Printf("[hostd] tcp listening on %s", srv.Addr())
	<-ctx.Done()
	srv.Close()
	return nil
}

// parsePlugins converts a comma-separated path list into plugin specs.
func parsePlugins(csv string) []plugin.Spec {
	var out []plugin.Spec
	for _, p := range splitCSV(csv) {
		if p == "" {
			continue
		}
		out = append(out, plugin.Spec{
			ID:   fmt.Sprintf("plugin-%d", len(out)+1),
			Name: baseName(p),
			Path: p,
		})
	}
	return out
}

func splitCSV(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func baseName(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}

// stdLogger logs host lifecycle events to stderr.
type stdLogger struct{}

// Logf implements manager.Logger.
func (stdLogger) Logf(id, format string, args ...any) {
	log.Printf("[%s] %s", id, fmt.Sprintf(format, args...))
}
