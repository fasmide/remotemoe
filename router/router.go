package router

import (
	"context"
	"net"
	"sync"
)

// Routable describes requirements to be routable
type Routable interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)

	// Replaced is used to indicate to a ssh session that it's routes
	// have been replaced with another ssh session
	Replaced()
}

// Router directs names to Dialers
type Router struct {
	sync.RWMutex
	endpoints map[string]Routable
}

// New returns a new router
func New() *Router {
	return &Router{
		endpoints: make(map[string]Routable),
	}
}

// Replace replaces a route, if another route was found
// the old route will have its .Replaced function called
// and this method will return true - if this was a new
// route, it will return false
func (r *Router) Replace(n string, d Routable) bool {
	r.Lock()
	oldRoute, exists := r.endpoints[n]
	if exists {
		go oldRoute.Replaced()
	}
	r.endpoints[n] = d
	r.Unlock()
	return exists
}

// Remove removes a route by name and Routable
func (r *Router) Remove(n string, d Routable) {
	r.Lock()
	defer r.Unlock()

	endpoint, ok := r.endpoints[n]
	if !ok {
		return
	}

	if d == endpoint {
		delete(r.endpoints, n)
	}

}

// Find fetches a route
func (r *Router) Find(n string) (Routable, bool) {
	r.RLock()
	d, ok := r.endpoints[n]
	r.RUnlock()
	return d, ok
}
