package router

import (
	"context"
	"net"
	"sync"
)

// Routable describes requirements to be routable
type Routable interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Routes directs names to Dialers
type Routes struct {
	sync.RWMutex
	endpoints map[string]Routable
}

// Set inserts a route, fails if already exist
func (r *Routes) Set(n string, d Routable) error {
	r.Lock()
	r.endpoints[n] = d
	r.Unlock()
}

// Find fetches a route
func (r *Routes) Find(n string) (Routable, bool) {
	r.RLock()
	d, ok := r.endpoints[n]
	r.RUnlock()
	return d, ok
}
