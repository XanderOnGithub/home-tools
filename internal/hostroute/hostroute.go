// Package hostroute sends each request to a tool by the first label of its
// Host: fitness.example.com → "fitness" (decision #5, ADR 0003).
package hostroute

import (
	"net"
	"net/http"
	"slices"
	"strings"
)

// Router picks a tool's handler by subdomain. Build it with New, add tools
// with Handle before serving; it's read-only afterwards, so it needs no lock.
type Router struct {
	tools map[string]http.Handler
	only  string // set: every host goes to this tool (local dev)
}

// New returns a router. only = "" routes by subdomain; a tool name sends
// every request to that tool, whatever the host: localhost and bare IPs
// have no subdomain, so local dev runs one tool at a time (-tool flag).
func New(only string) *Router {
	return &Router{tools: make(map[string]http.Handler), only: only}
}

// Handle mounts h as the tool called name ("fitness" serves fitness.*).
func (rt *Router) Handle(name string, h http.Handler) {
	rt.tools[name] = h
}

// Names returns the mounted tool names, sorted.
func (rt *Router) Names() []string {
	names := make([]string, 0, len(rt.tools))
	for n := range rt.tools {
		names = append(names, n)
	}
	slices.Sort(names)
	return names
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := rt.only
	if name == "" {
		name = subdomain(r.Host)
	}
	h, ok := rt.tools[name]
	if !ok {
		http.Error(w, "no tool here; try one of: "+strings.Join(rt.Names(), ", ")+" (as <tool>.<your domain>)", http.StatusNotFound)
		return
	}
	h.ServeHTTP(w, r)
}

// subdomain returns host's first label, without a port: "fitness" for
// "fitness.example.com:443". Bare hosts ("localhost") and IPs have none.
func subdomain(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if net.ParseIP(host) != nil {
		return ""
	}
	label, _, found := strings.Cut(host, ".")
	if !found {
		return ""
	}
	return strings.ToLower(label)
}
