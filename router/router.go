package router

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Routable describes requirements to be routable
type Routable interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)

	// A routable should be able to identify it self
	FQDN() string

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

// ErrNotFound errors are returned when searching for something
// that does not exist
type ErrNotFound error

// Replace replaces a route, if another route was found
// the old route will have its .Replaced function called
// and this method will return true - if this was a new
// route, it will return false
func (r *Router) Replace(d Routable) bool {
	n := d.FQDN()

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
func (r *Router) Remove(d Routable) {
	n := d.FQDN()

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

// Exists allows the acme/autocert to figure out if it should make certificate requests
func (r *Router) Exists(_ context.Context, s string) error {
	r.RLock()
	_, exists := r.endpoints[s]
	r.RUnlock()

	if !exists {
		return fmt.Errorf("%s does not exist", s).(ErrNotFound)
	}

	return nil
}

// DialContext is used by stuff that what to dial something up
func (r *Router) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("router: could not split host from port: %w", err)
	}

	r.RLock()
	d, exists := r.endpoints[host]
	r.RUnlock()

	if !exists {
		return nil, fmt.Errorf("router: %s not found", host).(ErrNotFound)
	}

	return d.DialContext(ctx, network, address)
}
